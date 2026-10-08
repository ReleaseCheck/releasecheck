// Package cli wires the supported registry paths to report renderers.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/releasecheck/releasecheck/internal/domain"
	"github.com/releasecheck/releasecheck/internal/npm"
	"github.com/releasecheck/releasecheck/internal/pypi"
	"github.com/releasecheck/releasecheck/internal/report"
	"github.com/releasecheck/releasecheck/internal/security"
)

var (
	version = "0.1.0-dev"
	commit  = "unknown"
)

// Run executes ReleaseCheck CLI arguments and returns a stable process code.
// It never invokes a package manager or executes package content.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		writeUsage(stdout)
		return 0
	}
	if args[0] == "--version" || args[0] == "-v" {
		fmt.Fprintf(stdout, "%s (commit %s)\n", version, commit)
		return 0
	}
	if args[0] != "verify" || len(args) < 2 {
		fmt.Fprintln(stderr, "releasecheck: expected `verify npm|pypi [flags] NAME [VERSION]`")
		writeUsage(stderr)
		return int(domain.ExitUsage)
	}

	ecosystem := args[1]
	set := flag.NewFlagSet("releasecheck verify "+ecosystem, flag.ContinueOnError)
	set.SetOutput(stderr)
	jsonOutput := set.Bool("json", false, "render the stable JSON report")
	sarifOutput := set.Bool("sarif", false, "render SARIF 2.1.0 output")
	outputPath := set.String("output", "", "write the report to a file instead of stdout")
	timeout := set.Duration("timeout", 30*time.Second, "maximum registry/source operation duration")
	artifactFilename := set.String("artifact", "", "PyPI release filename to verify")
	if err := set.Parse(args[2:]); err != nil {
		return int(domain.ExitUsage)
	}
	if *jsonOutput && *sarifOutput {
		fmt.Fprintln(stderr, "releasecheck: --json and --sarif cannot be used together")
		return int(domain.ExitUsage)
	}
	positional := set.Args()
	if len(positional) < 1 || len(positional) > 2 || (ecosystem != "npm" && ecosystem != "pypi") {
		fmt.Fprintln(stderr, "releasecheck: expected NAME and optional VERSION for npm or pypi")
		return int(domain.ExitUsage)
	}
	if *timeout <= 0 {
		fmt.Fprintln(stderr, "releasecheck: --timeout must be positive")
		return int(domain.ExitUsage)
	}

	operationContext, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()
	request := domain.ReleaseRequest{Registry: domain.Registry(ecosystem), Name: positional[0]}
	if len(positional) == 2 {
		request.Version = positional[1]
	}

	var (
		result  domain.Report
		err     error
		cleanup func() error
	)
	switch ecosystem {
	case "npm":
		client, clientErr := npm.NewClient(npm.Options{})
		if clientErr != nil {
			err = clientErr
			break
		}
		verification, verifyErr := client.Verify(operationContext, request)
		if verifyErr != nil {
			err = verifyErr
			break
		}
		result, err = reportFromNPM(verification)
		cleanup = verification.Cleanup
	case "pypi":
		client, clientErr := pypi.NewClient(pypi.Options{})
		if clientErr != nil {
			err = clientErr
			break
		}
		verification, verifyErr := client.Verify(operationContext, request, pypi.VerifyOptions{ArtifactFilename: *artifactFilename})
		if verifyErr != nil {
			err = verifyErr
			break
		}
		result, err = reportFromPyPI(verification)
		cleanup = verification.Cleanup
	}
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		fmt.Fprintf(stderr, "releasecheck: %v\n", err)
		return int(domain.ExitError)
	}
	normalized, err := report.Normalize(result)
	if err != nil {
		fmt.Fprintf(stderr, "releasecheck: normalize report: %v\n", err)
		return int(domain.ExitError)
	}

	data, err := render(normalized, *jsonOutput, *sarifOutput)
	if err != nil {
		fmt.Fprintf(stderr, "releasecheck: render report: %v\n", err)
		return int(domain.ExitError)
	}
	if err := writeOutput(*outputPath, data, stdout); err != nil {
		fmt.Fprintf(stderr, "releasecheck: write report: %v\n", err)
		return int(domain.ExitError)
	}
	return int(report.ExitCode(normalized))
}

func reportFromNPM(result npm.VerificationResult) (domain.Report, error) {
	if len(result.Metadata.Artifacts) != 1 {
		return domain.Report{}, errors.New("npm verification returned no single artifact")
	}
	securityFindings, err := security.AnalyzeInventory(result.ArtifactInventory, security.Options{})
	if err != nil {
		return domain.Report{}, err
	}
	artifact := result.Metadata.Artifacts[0]
	return domain.Report{
		Identity:    result.Metadata.Identity,
		Artifact:    artifact,
		Source:      result.Metadata.Source,
		Git:         result.Metadata.Git,
		Comparisons: result.Comparisons,
		Security:    securityFindings,
		Evidence:    result.Metadata.Evidence,
		Provenance:  result.Metadata.Provenance,
		Limitations: result.Limitations,
	}, nil
}

func reportFromPyPI(result pypi.VerificationResult) (domain.Report, error) {
	securityFindings, err := security.AnalyzeInventory(result.ArtifactInventory, security.Options{})
	if err != nil {
		return domain.Report{}, err
	}
	artifact := result.Selected
	for _, candidate := range result.Metadata.Artifacts {
		if candidate.Filename == result.Selected.Filename {
			artifact = candidate
			break
		}
	}
	return domain.Report{
		Identity:    result.Metadata.Identity,
		Artifact:    artifact,
		Source:      result.Metadata.Source,
		Git:         result.Metadata.Git,
		Comparisons: result.Comparisons,
		Security:    securityFindings,
		Evidence:    result.Metadata.Evidence,
		Provenance:  result.Metadata.Provenance,
		Limitations: result.Limitations,
	}, nil
}

func render(input domain.Report, jsonOutput, sarifOutput bool) ([]byte, error) {
	if sarifOutput {
		return report.SARIF(input)
	}
	if jsonOutput {
		return report.JSON(input)
	}
	human, err := report.Human(input)
	return []byte(human), err
}

func writeOutput(path string, data []byte, stdout io.Writer) error {
	if path == "" {
		_, err := stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func writeUsage(writer io.Writer) {
	fmt.Fprintln(writer, "ReleaseCheck: deterministic package release evidence")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Usage:")
	fmt.Fprintln(writer, "  releasecheck verify npm [flags] NAME [VERSION]")
	fmt.Fprintln(writer, "  releasecheck verify pypi [flags] NAME [VERSION]")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Flags:")
	fmt.Fprintln(writer, "  --json                 render JSON schema 1.0")
	fmt.Fprintln(writer, "  --sarif                render SARIF 2.1.0")
	fmt.Fprintln(writer, "  --output PATH          write the report to PATH")
	fmt.Fprintln(writer, "  --artifact FILENAME    select a PyPI release file")
	fmt.Fprintln(writer, "  --timeout DURATION     bound registry and source requests")
}

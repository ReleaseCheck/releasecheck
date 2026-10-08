// Package report normalizes and renders ReleaseCheck evidence.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/releasecheck/releasecheck/internal/domain"
)

const (
	// SchemaVersion is the stable JSON report contract for Phase 9.
	SchemaVersion = "1.0"
	SarifVersion  = "2.1.0"
)

// Normalize validates, sorts, and evaluates a report without changing the
// caller's value. Report output is deterministic after normalization.
func Normalize(input domain.Report) (domain.Report, error) {
	if err := validate(input); err != nil {
		return domain.Report{}, err
	}
	result := input
	result.SchemaVersion = SchemaVersion
	result.GeneratedAt = input.GeneratedAt
	result.Comparisons = domain.SortComparisons(input.Comparisons)
	result.Security = domain.SortSecurityObservations(input.Security)
	result.Evidence = sortEvidence(input.Evidence)
	result.Provenance = sortProvenance(input.Provenance)
	result.Verdict = Evaluate(result)
	return result, nil
}

// Evaluate maps deterministic evidence to a policy-neutral verdict. MATCH
// means the available comparison evidence agrees; it never means safe.
func Evaluate(report domain.Report) domain.Verdict {
	if len(report.Comparisons) == 0 || report.Source == nil || report.Git == nil {
		return domain.VerdictIncomplete
	}
	if len(report.Limitations) > 0 {
		return domain.VerdictIncomplete
	}
	for _, item := range report.Comparisons {
		if item.Kind != domain.ComparisonIdentical {
			return domain.VerdictReview
		}
	}
	for _, item := range report.Security {
		if item.State == domain.EvidenceInvalid || item.Severity == domain.SecurityWarning {
			return domain.VerdictReview
		}
	}
	for _, item := range report.Evidence {
		if item.State == domain.EvidenceInvalid {
			return domain.VerdictReview
		}
	}
	for _, item := range report.Provenance {
		if item.Status == domain.ProvenanceInvalid || item.Status == domain.ProvenanceInsufficient {
			return domain.VerdictReview
		}
	}
	return domain.VerdictMatch
}

// ExitCode maps a normalized verdict to the stable process contract.
func ExitCode(report domain.Report) domain.ExitCode {
	switch report.Verdict {
	case domain.VerdictMatch:
		return domain.ExitSuccess
	case domain.VerdictReview:
		return domain.ExitReview
	case domain.VerdictIncomplete:
		return domain.ExitReview
	default:
		return domain.ExitError
	}
}

// JSON returns deterministic, indented JSON with a trailing newline. Generated
// timestamps are omitted unless RenderOptions.IncludeGeneratedAt is requested.
func JSON(report domain.Report) ([]byte, error) {
	return JSONWithOptions(report, RenderOptions{})
}

// RenderOptions controls intentionally nondeterministic report metadata.
type RenderOptions struct {
	IncludeGeneratedAt bool
}

// JSONWithOptions renders the versioned JSON report.
func JSONWithOptions(input domain.Report, options RenderOptions) ([]byte, error) {
	report, err := Normalize(input)
	if err != nil {
		return nil, err
	}
	if !options.IncludeGeneratedAt {
		report.GeneratedAt = time.Time{}
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal JSON report: %w", err)
	}
	return append(data, '\n'), nil
}

// WriteJSON writes JSON without retaining report data after the call.
func WriteJSON(writer io.Writer, report domain.Report) error {
	data, err := JSON(report)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

// Human returns a concise deterministic report for terminal and CI logs.
func Human(input domain.Report) (string, error) {
	report, err := Normalize(input)
	if err != nil {
		return "", err
	}
	var output strings.Builder
	fmt.Fprintln(&output, "Release Integrity")
	fmt.Fprintf(&output, "Package: %s@%s (%s)\n", report.Identity.Name, report.Identity.Version, report.Identity.Registry)
	fmt.Fprintf(&output, "Verdict: %s\n", report.Verdict)
	fmt.Fprintf(&output, "Exit code: %d\n", ExitCode(report))

	fmt.Fprintln(&output, "Evidence:")
	for _, item := range report.Evidence {
		fmt.Fprintf(&output, "%s %s: %s", evidenceMarker(item.State), item.ID, item.Subject)
		if item.Value != "" {
			fmt.Fprintf(&output, " = %s", item.Value)
		}
		if item.Description != "" {
			fmt.Fprintf(&output, " (%s)", item.Description)
		}
		output.WriteByte('\n')
	}
	for _, item := range report.Comparisons {
		fmt.Fprintf(&output, "%s comparison.%s: %s\n", comparisonMarker(item.Kind), item.Kind, item.Path)
	}
	for _, item := range report.Security {
		fmt.Fprintf(&output, "%s security.%s: %s", securityMarker(item), item.ID, item.Subject)
		if item.Description != "" {
			fmt.Fprintf(&output, " (%s)", item.Description)
		}
		output.WriteByte('\n')
	}
	for _, item := range report.Provenance {
		fmt.Fprintf(&output, "%s provenance.%s: %s", provenanceMarker(item.Status), item.Status, item.Subject)
		if item.Explanation != "" {
			fmt.Fprintf(&output, " (%s)", item.Explanation)
		}
		output.WriteByte('\n')
	}
	return output.String(), nil
}

// SARIF returns a stable SARIF 2.1.0 log. Findings are observations and
// comparison relationships, not claims that code is malicious.
func SARIF(input domain.Report) ([]byte, error) {
	report, err := Normalize(input)
	if err != nil {
		return nil, err
	}
	type location struct {
		ArtifactLocation struct {
			URI string `json:"uri"`
		} `json:"artifactLocation"`
	}
	type result struct {
		RuleID  string `json:"ruleId"`
		Level   string `json:"level"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
		Locations []location `json:"locations,omitempty"`
	}
	results := make([]result, 0)
	add := func(ruleID, level, message, subject string) {
		item := result{RuleID: ruleID, Level: level}
		item.Message.Text = message
		if subject != "" {
			var where location
			where.ArtifactLocation.URI = subject
			item.Locations = []location{where}
		}
		results = append(results, item)
	}
	for _, item := range report.Comparisons {
		if item.Kind != domain.ComparisonIdentical {
			add("comparison."+string(item.Kind), "warning", "Release artifact and source differ or cannot be fully compared.", item.Path)
		}
	}
	for _, item := range report.Security {
		level := "note"
		if item.Severity == domain.SecurityWarning || item.State == domain.EvidenceInvalid {
			level = "warning"
		}
		add("security."+item.ID, level, item.Description, item.Subject)
	}
	for _, item := range report.Provenance {
		if item.Status == domain.ProvenancePresent {
			continue
		}
		level := "note"
		if item.Status == domain.ProvenanceInvalid || item.Status == domain.ProvenanceInsufficient {
			level = "warning"
		}
		add("provenance."+string(item.Status), level, item.Explanation, item.Subject)
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].RuleID == results[j].RuleID {
			return results[i].Message.Text < results[j].Message.Text
		}
		return results[i].RuleID < results[j].RuleID
	})
	log := struct {
		Version string `json:"version"`
		Schema  string `json:"$schema"`
		Runs    []struct {
			Tool struct {
				Driver struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				} `json:"driver"`
			} `json:"tool"`
			Results []result `json:"results"`
		} `json:"runs"`
	}{Version: SarifVersion, Schema: "https://json.schemastore.org/sarif-2.1.0.json"}
	run := struct {
		Tool struct {
			Driver struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"driver"`
		} `json:"tool"`
		Results []result `json:"results"`
	}{}
	run.Tool.Driver.Name = "ReleaseCheck"
	run.Tool.Driver.Version = SchemaVersion
	run.Results = results
	log.Runs = []struct {
		Tool struct {
			Driver struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"driver"`
		} `json:"tool"`
		Results []result `json:"results"`
	}{run}
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal SARIF report: %w", err)
	}
	return append(data, '\n'), nil
}

func validate(report domain.Report) error {
	if err := report.Identity.Validate(); err != nil {
		return fmt.Errorf("report identity: %w", err)
	}
	if err := report.Artifact.Validate(); err != nil {
		return fmt.Errorf("report artifact: %w", err)
	}
	if report.SchemaVersion != "" && report.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported report schema version %q", report.SchemaVersion)
	}
	return nil
}

func sortEvidence(items []domain.Evidence) []domain.Evidence {
	result := append([]domain.Evidence(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].ID == result[j].ID {
			left := result[i].Subject + "\x00" + string(result[i].State) + "\x00" + result[i].Value + "\x00" + result[i].Description
			right := result[j].Subject + "\x00" + string(result[j].State) + "\x00" + result[j].Value + "\x00" + result[j].Description
			return left < right
		}
		return result[i].ID < result[j].ID
	})
	return result
}

func sortProvenance(items []domain.ProvenanceEvidence) []domain.ProvenanceEvidence {
	result := append([]domain.ProvenanceEvidence(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		left := string(result[i].Status) + "\x00" + result[i].Predicate + "\x00" + result[i].Subject + "\x00" + result[i].Digest + "\x00" + result[i].Identity + "\x00" + result[i].SourceURI + "\x00" + result[i].Commit + "\x00" + result[i].Explanation
		right := string(result[j].Status) + "\x00" + result[j].Predicate + "\x00" + result[j].Subject + "\x00" + result[j].Digest + "\x00" + result[j].Identity + "\x00" + result[j].SourceURI + "\x00" + result[j].Commit + "\x00" + result[j].Explanation
		return left < right
	})
	return result
}

func evidenceMarker(state domain.EvidenceState) string {
	if state == domain.EvidenceKnown {
		return "OK"
	}
	return "INFO"
}

func comparisonMarker(kind domain.ComparisonKind) string {
	if kind == domain.ComparisonIdentical {
		return "OK"
	}
	return "WARN"
}

func securityMarker(item domain.SecurityObservation) string {
	if item.State == domain.EvidenceInvalid || item.Severity == domain.SecurityWarning {
		return "WARN"
	}
	return "INFO"
}

func provenanceMarker(status domain.ProvenanceStatus) string {
	if status == domain.ProvenancePresent {
		return "OK"
	}
	return "INFO"
}

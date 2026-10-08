package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/releasecheck/releasecheck/internal/domain"
)

func TestRenderersNormalizeDeterministically(t *testing.T) {
	report := completeReport()
	report.GeneratedAt = time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	first, err := JSON(report)
	if err != nil {
		t.Fatal(err)
	}
	second, err := JSON(report)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("repeated JSON rendering changed output")
	}
	if bytes.Contains(first, []byte("2026-10-08")) {
		t.Fatal("default JSON rendering included nondeterministic generated time")
	}
	if !bytes.Contains(first, []byte(`"schema_version": "1.0"`)) {
		t.Fatal("schema version missing")
	}

	human, err := Human(report)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(human, "Verdict: MATCH") || !strings.Contains(human, "comparison.identical") {
		t.Fatalf("human report missing stable summary: %s", human)
	}

	sarif, err := SARIF(report)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Version string `json:"version"`
		Runs    []struct {
			Results []json.RawMessage `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(sarif, &document); err != nil {
		t.Fatal(err)
	}
	if document.Version != SarifVersion || len(document.Runs) != 1 || len(document.Runs[0].Results) != 0 {
		t.Fatalf("unexpected SARIF document: %s", sarif)
	}
}

func TestEvaluateReviewAndIncompleteExitCodes(t *testing.T) {
	review := completeReport()
	review.Comparisons[0].Kind = domain.ComparisonModified
	normalized, err := Normalize(review)
	if err != nil {
		t.Fatal(err)
	}
	if normalized.Verdict != domain.VerdictReview || ExitCode(normalized) != domain.ExitReview {
		t.Fatalf("review policy: %#v", normalized)
	}

	incomplete := completeReport()
	incomplete.Source = nil
	normalized, err = Normalize(incomplete)
	if err != nil {
		t.Fatal(err)
	}
	if normalized.Verdict != domain.VerdictIncomplete || ExitCode(normalized) != domain.ExitReview {
		t.Fatalf("incomplete policy: %#v", normalized)
	}
}

func TestNormalizeRejectsUnsupportedSchemaAndInvalidArtifact(t *testing.T) {
	unsupported := completeReport()
	unsupported.SchemaVersion = "9.0"
	if _, err := Normalize(unsupported); err == nil {
		t.Fatal("unsupported schema version was accepted")
	}

	invalid := completeReport()
	invalid.Artifact.URL = "http://example.test/artifact.tgz"
	if _, err := Normalize(invalid); err == nil {
		t.Fatal("invalid artifact URL was accepted")
	}
}

func TestSARIFIncludesNonIdenticalEvidence(t *testing.T) {
	report := completeReport()
	report.Comparisons[0].Kind = domain.ComparisonModified
	report.Security = []domain.SecurityObservation{{
		ID:       "npm.install_script",
		State:    domain.EvidenceKnown,
		Severity: domain.SecurityWarning,
		Subject:  "package.json.scripts.postinstall",
	}}
	report.Provenance = []domain.ProvenanceEvidence{{
		Status:      domain.ProvenanceInsufficient,
		Subject:     "demo-1.0.0.tgz",
		Explanation: "signature was not verified",
	}}
	data, err := SARIF(report)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(`"comparison.modified"`)) || !bytes.Contains(data, []byte(`"security.npm.install_script"`)) || !bytes.Contains(data, []byte(`"provenance.insufficient"`)) {
		t.Fatalf("SARIF findings missing: %s", data)
	}
}

func completeReport() domain.Report {
	identity := domain.PackageIdentity{Registry: domain.RegistryNPM, Name: "demo", Version: "1.0.0"}
	hash := strings.Repeat("a", 64)
	return domain.Report{
		Identity: identity,
		Artifact: domain.Artifact{
			Identity: identity,
			Filename: "demo-1.0.0.tgz",
			URL:      "https://registry.example/demo.tgz",
			SHA256:   hash,
		},
		Source: &domain.SourceReference{URL: "https://github.com/example/demo"},
		Git:    &domain.GitReference{Ref: "0123456789abcdef", Commit: "0123456789abcdef", Immutable: true},
		Comparisons: []domain.FileComparison{{
			Path: "index.js",
			Kind: domain.ComparisonIdentical,
		}},
		Evidence: []domain.Evidence{{
			ID:      "npm.version.selected",
			State:   domain.EvidenceKnown,
			Subject: "package.version",
			Value:   "1.0.0",
		}},
	}
}

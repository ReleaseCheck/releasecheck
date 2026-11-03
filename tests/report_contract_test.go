package tests

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/releasecheck/releasecheck/internal/domain"
	"github.com/releasecheck/releasecheck/internal/report"
)

// TestReportContractSemantics checks the public meaning of each report mode
// through the renderers, independently of registry and archive acquisition.
func TestReportContractSemantics(t *testing.T) {
	cases := []struct {
		name          string
		mutate        func(*domain.Report)
		verdict       domain.Verdict
		exitCode      domain.ExitCode
		humanContains string
		jsonContains  string
		sarifContains string
	}{
		{
			name:          "match is bounded evidence",
			verdict:       domain.VerdictMatch,
			exitCode:      domain.ExitSuccess,
			humanContains: "Verdict: MATCH",
			jsonContains:  `"reason_code": "comparison.identical"`,
		},
		{
			name: "difference is review, not malware verdict",
			mutate: func(value *domain.Report) {
				value.Comparisons[0].Kind = domain.ComparisonModified
			},
			verdict:       domain.VerdictReview,
			exitCode:      domain.ExitReview,
			humanContains: "Verdict: REVIEW",
			jsonContains:  `"reason_code": "evidence.requires_review"`,
			sarifContains: `"comparison.modified"`,
		},
		{
			name: "missing source is incomplete",
			mutate: func(value *domain.Report) {
				value.Source = nil
			},
			verdict:       domain.VerdictIncomplete,
			exitCode:      domain.ExitReview,
			humanContains: "Verdict: INCOMPLETE",
			jsonContains:  `"reason_code": "evidence.incomplete"`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			value := contractReport()
			if testCase.mutate != nil {
				testCase.mutate(&value)
			}
			normalized, err := report.Normalize(value)
			if err != nil {
				t.Fatal(err)
			}
			if normalized.Verdict != testCase.verdict || report.ExitCode(normalized) != testCase.exitCode {
				t.Fatalf("policy: got verdict=%s exit=%d, want verdict=%s exit=%d", normalized.Verdict, report.ExitCode(normalized), testCase.verdict, testCase.exitCode)
			}
			human, err := report.Human(value)
			if err != nil || !strings.Contains(human, testCase.humanContains) {
				t.Fatalf("human report: %v\n%s", err, human)
			}
			jsonData, err := report.JSON(value)
			if err != nil || !strings.Contains(string(jsonData), testCase.jsonContains) {
				t.Fatalf("JSON report: %v\n%s", err, jsonData)
			}
			if testCase.sarifContains != "" {
				sarifData, err := report.SARIF(value)
				if err != nil || !strings.Contains(string(sarifData), testCase.sarifContains) {
					t.Fatalf("SARIF report: %v\n%s", err, sarifData)
				}
			}
			var decoded map[string]any
			if err := json.Unmarshal(jsonData, &decoded); err != nil {
				t.Fatalf("report JSON remains parseable: %v", err)
			}
		})
	}
}

func contractReport() domain.Report {
	identity := domain.PackageIdentity{Registry: domain.RegistryNPM, Name: "contract-fixture", Version: "1.0.0"}
	return domain.Report{
		Identity: identity,
		Artifact: domain.Artifact{Identity: identity, Filename: "contract-fixture.tgz", URL: "https://registry.example/contract-fixture.tgz", SHA256: strings.Repeat("a", 64)},
		Source:   &domain.SourceReference{URL: "https://github.com/example/contract-fixture"},
		Git:      &domain.GitReference{Ref: strings.Repeat("b", 40), Commit: strings.Repeat("b", 40), Immutable: true},
		Comparisons: []domain.FileComparison{{
			Path: "index.js",
			Kind: domain.ComparisonIdentical,
		}},
	}
}

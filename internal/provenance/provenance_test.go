package provenance

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/releasecheck/releasecheck/internal/domain"
)

func TestParsePyPIProvenanceBindsSubjectAndPreservesLimitations(t *testing.T) {
	statement := statementJSON("sampleproject-1.0.0.tar.gz", "sha256", "abc123", PredicateSLSAProvenance)
	statement["predicate"] = map[string]any{
		"buildDefinition": map[string]any{
			"externalParameters": map[string]any{
				"source": map[string]any{
					"uri":    "https://github.com/example/sampleproject",
					"digest": map[string]string{"gitCommit": "0123456789abcdef"},
				},
			},
		},
	}
	document := map[string]any{
		"version": 1,
		"attestation_bundles": []any{map[string]any{
			"publisher": map[string]any{"kind": "GitHub"},
			"attestations": []any{map[string]any{
				"envelope": map[string]any{"payload": payload(statement)},
			}},
		}},
	}
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	got := ParsePyPI(data, Binding{Filename: "sampleproject-1.0.0.tar.gz", Digests: map[string]string{"sha256": "abc123"}})
	if len(got) != 1 {
		t.Fatalf("got %#v", got)
	}
	if got[0].Status != domain.ProvenanceInsufficient {
		t.Fatalf("status: got %q want insufficient", got[0].Status)
	}
	if got[0].Identity != "GitHub" || got[0].Predicate != PredicateSLSAProvenance || got[0].Digest != "sha256:abc123" {
		t.Fatalf("evidence binding: %#v", got[0])
	}
	if got[0].SourceURI != "https://github.com/example/sampleproject" || got[0].Commit != "0123456789abcdef" {
		t.Fatalf("source extraction: %#v", got[0])
	}
	if !strings.Contains(got[0].Explanation, "not verified") {
		t.Fatalf("missing verification limitation: %#v", got[0])
	}
}

func TestParseProvenanceRejectsMismatchesAndMalformedObjects(t *testing.T) {
	statement := statementJSON("other.whl", "sha256", "abc123", PredicatePyPIPublish)
	document := map[string]any{
		"version": 1,
		"attestation_bundles": []any{map[string]any{
			"attestations": []any{map[string]any{"envelope": map[string]any{"payload": payload(statement)}}},
		}},
	}
	data, _ := json.Marshal(document)
	got := ParsePyPI(data, Binding{Filename: "expected.whl", Digests: map[string]string{"sha256": "abc123"}})
	if len(got) != 1 || got[0].Status != domain.ProvenanceInvalid {
		t.Fatalf("filename mismatch: %#v", got)
	}

	malformed := ParsePyPI([]byte("{"), Binding{})
	if len(malformed) != 1 || malformed[0].Status != domain.ProvenanceInvalid {
		t.Fatalf("malformed JSON: %#v", malformed)
	}

	unsupported := statementJSON("expected.whl", "sha256", "abc123", "https://example.invalid/unknown")
	npm := map[string]any{"bundles": []any{map[string]any{
		"dsseEnvelope": map[string]any{"payload": payload(unsupported)},
	}}}
	npmData, _ := json.Marshal(npm)
	npmEvidence := ParseNPM(npmData, Binding{Filename: "expected.whl", Digests: map[string]string{"sha256": "abc123"}})
	if len(npmEvidence) != 1 || npmEvidence[0].Status != domain.ProvenanceInsufficient {
		t.Fatalf("unsupported predicate should remain insufficient: %#v", npmEvidence)
	}
}

func TestParseNPMReportsUnavailableAndAbsentSeparately(t *testing.T) {
	unavailable := ParseNPM(nil, Binding{})
	if len(unavailable) != 1 || unavailable[0].Status != domain.ProvenanceUnavailable {
		t.Fatalf("unavailable: %#v", unavailable)
	}
	absent := Absent("npm", "registry did not return an attestation")
	if absent.Status != domain.ProvenanceAbsent {
		t.Fatalf("absent: %#v", absent)
	}
}

func TestParseNPMSelectsDigestDeterministically(t *testing.T) {
	statement := statementJSON("expected.tgz", "sha512", "good512", PredicateSLSAProvenance)
	statement["subject"] = []any{map[string]any{
		"name":   "expected.tgz",
		"digest": map[string]string{"sha512": "good512", "sha256": "good256"},
	}}
	document := map[string]any{"bundles": []any{map[string]any{
		"dsseEnvelope": map[string]any{"payload": payload(statement)},
	}}}
	data, _ := json.Marshal(document)
	binding := Binding{Filename: "expected.tgz", Digests: map[string]string{"sha512": "good512", "sha256": "good256"}}
	for range 20 {
		got := ParseNPM(data, binding)
		if len(got) != 1 || got[0].Digest != "sha256:good256" {
			t.Fatalf("digest selection was not deterministic: %#v", got)
		}
	}
}

func statementJSON(name, algorithm, digest, predicate string) map[string]any {
	return map[string]any{
		"_type": "https://in-toto.io/Statement/v1",
		"subject": []any{map[string]any{
			"name":   name,
			"digest": map[string]string{algorithm: digest},
		}},
		"predicateType": predicate,
		"predicate":     map[string]any{},
	}
}

func payload(statement map[string]any) string {
	data, _ := json.Marshal(statement)
	return base64.StdEncoding.EncodeToString(data)
}

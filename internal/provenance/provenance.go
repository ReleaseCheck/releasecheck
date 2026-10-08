// Package provenance consumes registry-provided provenance structures without
// recreating the registries' or Sigstore's cryptographic verification systems.
package provenance

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/releasecheck/releasecheck/internal/domain"
)

const (
	PredicateSLSAProvenance = "https://slsa.dev/provenance/v1"
	PredicateSLAsync        = "https://slsa.dev/provenance/v0.2"
	PredicatePyPIPublish    = "https://docs.pypi.org/attestations/publish/v1"
	maxProvenanceBytes      = 16 << 20
)

// Binding contains the artifact identity needed to bind attestation subjects
// to the downloaded release. Digest values are hexadecimal and keyed by
// algorithm name, for example sha256 or sha512.
type Binding struct {
	Filename string
	Digests  map[string]string
}

// ParsePyPI parses a PEP 740 provenance object. It validates its structural
// shape and artifact subject, but deliberately does not verify DSSE signatures
// or Sigstore trust roots.
func ParsePyPI(data []byte, binding Binding) []domain.ProvenanceEvidence {
	if len(bytes.TrimSpace(data)) == 0 {
		return []domain.ProvenanceEvidence{unavailable("pypi", "provenance object is empty")}
	}
	if len(data) > maxProvenanceBytes {
		return []domain.ProvenanceEvidence{invalid("pypi", "provenance object exceeds the metadata limit")}
	}
	var object struct {
		Version            int               `json:"version"`
		AttestationBundles []json.RawMessage `json:"attestation_bundles"`
	}
	if err := json.Unmarshal(data, &object); err != nil {
		return []domain.ProvenanceEvidence{invalid("pypi", "malformed provenance JSON: "+err.Error())}
	}
	if object.Version != 1 || len(object.AttestationBundles) == 0 {
		return []domain.ProvenanceEvidence{invalid("pypi", "provenance version must be 1 with at least one attestation bundle")}
	}
	result := make([]domain.ProvenanceEvidence, 0)
	for _, bundle := range object.AttestationBundles {
		result = append(result, parsePyPIBundle(bundle, binding)...)
	}
	return result
}

// ParseNPM parses one npm registry/CLI attestation document. npm's exposed
// bundle shape may evolve, so this accepts a DSSE envelope, a bundle wrapper,
// or an object containing bundles without claiming that every future shape is
// understood.
func ParseNPM(data []byte, binding Binding) []domain.ProvenanceEvidence {
	if len(bytes.TrimSpace(data)) == 0 {
		return []domain.ProvenanceEvidence{unavailable("npm", "attestation document is empty")}
	}
	if len(data) > maxProvenanceBytes {
		return []domain.ProvenanceEvidence{invalid("npm", "attestation document exceeds the metadata limit")}
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return []domain.ProvenanceEvidence{invalid("npm", "malformed attestation JSON: "+err.Error())}
	}
	objects := collectBundleObjects(value)
	if len(objects) == 0 {
		return []domain.ProvenanceEvidence{insufficient("npm", "attestation shape is unavailable or unsupported")}
	}
	result := make([]domain.ProvenanceEvidence, 0, len(objects))
	for _, object := range objects {
		result = append(result, parseBundle(object, binding, "npm")...)
	}
	return result
}

// Absent returns explicit evidence for a registry response that says no
// provenance exists. A missing attestation is not an invalid attestation.
func Absent(registry, explanation string) domain.ProvenanceEvidence {
	return domain.ProvenanceEvidence{
		Status:      domain.ProvenanceAbsent,
		Explanation: registry + ": " + explanation,
	}
}

func parsePyPIBundle(data json.RawMessage, binding Binding) []domain.ProvenanceEvidence {
	var bundle struct {
		Publisher struct {
			Kind string `json:"kind"`
		} `json:"publisher"`
		Attestations []json.RawMessage `json:"attestations"`
	}
	if err := json.Unmarshal(data, &bundle); err != nil || len(bundle.Attestations) == 0 {
		return []domain.ProvenanceEvidence{invalid("pypi", "attestation bundle is malformed or empty")}
	}
	result := make([]domain.ProvenanceEvidence, 0, len(bundle.Attestations))
	for _, attestation := range bundle.Attestations {
		var value any
		if err := json.Unmarshal(attestation, &value); err != nil {
			result = append(result, invalid("pypi", "attestation is malformed"))
			continue
		}
		items := parseBundle(value, binding, "pypi")
		for index := range items {
			if bundle.Publisher.Kind != "" {
				items[index].Identity = bundle.Publisher.Kind
			}
		}
		result = append(result, items...)
	}
	return result
}

func parseBundle(value any, binding Binding, registry string) []domain.ProvenanceEvidence {
	object, ok := value.(map[string]any)
	if !ok {
		return []domain.ProvenanceEvidence{invalid(registry, "attestation bundle is not an object")}
	}
	if nested, ok := object["bundle"]; ok {
		if nestedObject, ok := nested.(map[string]any); ok {
			object = nestedObject
		}
	}
	envelope, ok := object["dsseEnvelope"].(map[string]any)
	if !ok {
		envelope, ok = object["envelope"].(map[string]any)
	}
	if !ok {
		return []domain.ProvenanceEvidence{insufficient(registry, "DSSE envelope is not present")}
	}
	predicateHint, _ := object["predicateType"].(string)
	return parseEnvelope(envelope, predicateHint, binding, registry)
}

func parseEnvelope(envelope map[string]any, predicateHint string, binding Binding, registry string) []domain.ProvenanceEvidence {
	payload, ok := envelope["payload"].(string)
	if !ok || payload == "" {
		return []domain.ProvenanceEvidence{invalid(registry, "DSSE payload is missing")}
	}
	decoded, err := decodeBase64(payload)
	if err != nil {
		return []domain.ProvenanceEvidence{invalid(registry, "DSSE payload is not valid base64")}
	}
	var statement struct {
		Type    string `json:"_type"`
		Subject []struct {
			Name   string            `json:"name"`
			Digest map[string]string `json:"digest"`
		} `json:"subject"`
		PredicateType string          `json:"predicateType"`
		Predicate     json.RawMessage `json:"predicate"`
	}
	if err := json.Unmarshal(decoded, &statement); err != nil || len(statement.Subject) != 1 {
		return []domain.ProvenanceEvidence{invalid(registry, "DSSE payload is not a single-subject in-toto statement")}
	}
	predicate := statement.PredicateType
	if predicate == "" {
		predicate = predicateHint
	}
	if predicate == "" {
		return []domain.ProvenanceEvidence{insufficient(registry, "attestation predicate type is missing")}
	}
	subject := statement.Subject[0]
	evidence := domain.ProvenanceEvidence{Predicate: predicate, Subject: subject.Name}
	if subject.Name != "" && binding.Filename != "" && subject.Name != binding.Filename {
		return []domain.ProvenanceEvidence{invalid(registry, "attestation subject filename does not match artifact")}
	}
	digestAlgorithm, digest, matched := matchDigest(subject.Digest, binding.Digests)
	if !matched {
		return []domain.ProvenanceEvidence{invalidWithEvidence(registry, evidence, "attestation subject digest does not match the artifact or uses an unavailable digest algorithm")}
	}
	evidence.Digest = digestAlgorithm + ":" + strings.ToLower(digest)
	evidence.SourceURI, evidence.Commit = extractSource(statement.Predicate)
	evidence.Status = domain.ProvenanceInsufficient
	evidence.Explanation = "attestation structure and artifact subject are bound, but DSSE signature and registry trust roots were not verified by ReleaseCheck"
	if !supportedPredicate(predicate) {
		evidence.Explanation = "attestation subject is bound, but the predicate type is not supported for stronger interpretation"
	}
	return []domain.ProvenanceEvidence{evidence}
}

func collectBundleObjects(value any) []any {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	for _, key := range []string{"bundles", "attestations"} {
		if values, ok := object[key].([]any); ok {
			return values
		}
	}
	return []any{object}
}

func matchDigest(subject, expected map[string]string) (string, string, bool) {
	algorithms := make([]string, 0, len(subject))
	for algorithm := range subject {
		algorithms = append(algorithms, algorithm)
	}
	sort.Strings(algorithms)
	for _, algorithm := range algorithms {
		digest := subject[algorithm]
		want, ok := expected[strings.ToLower(algorithm)]
		if ok && strings.EqualFold(strings.TrimSpace(digest), strings.TrimSpace(want)) {
			return strings.ToLower(algorithm), digest, true
		}
	}
	return "", "", false
}

func extractSource(raw json.RawMessage) (string, string) {
	var predicate struct {
		BuildDefinition struct {
			ExternalParameters struct {
				Source struct {
					URI    string            `json:"uri"`
					Digest map[string]string `json:"digest"`
				} `json:"source"`
			} `json:"externalParameters"`
		} `json:"buildDefinition"`
	}
	if json.Unmarshal(raw, &predicate) != nil {
		return "", ""
	}
	commit := predicate.BuildDefinition.ExternalParameters.Source.Digest["gitCommit"]
	return predicate.BuildDefinition.ExternalParameters.Source.URI, commit
}

func decodeBase64(value string) ([]byte, error) {
	for _, encoding := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		decoded, err := encoding.DecodeString(value)
		if err == nil {
			return decoded, nil
		}
	}
	return nil, fmt.Errorf("invalid base64 payload")
}

func supportedPredicate(predicate string) bool {
	return predicate == PredicateSLSAProvenance || predicate == PredicateSLAsync || predicate == PredicatePyPIPublish
}

func unavailable(registry, explanation string) domain.ProvenanceEvidence {
	return domain.ProvenanceEvidence{Status: domain.ProvenanceUnavailable, Explanation: registry + ": " + explanation}
}

func invalid(registry, explanation string) domain.ProvenanceEvidence {
	return domain.ProvenanceEvidence{Status: domain.ProvenanceInvalid, Explanation: registry + ": " + explanation}
}

func insufficient(registry, explanation string) domain.ProvenanceEvidence {
	return domain.ProvenanceEvidence{Status: domain.ProvenanceInsufficient, Explanation: registry + ": " + explanation}
}

func invalidWithEvidence(registry string, evidence domain.ProvenanceEvidence, explanation string) domain.ProvenanceEvidence {
	evidence.Status = domain.ProvenanceInvalid
	evidence.Explanation = registry + ": " + explanation
	return evidence
}

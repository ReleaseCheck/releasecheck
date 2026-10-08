// Package domain defines the contracts shared by ReleaseCheck's acquisition,
// comparison, provenance, reporting, and registry-specific packages.
package domain

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Registry identifies a supported package registry.
type Registry string

const (
	RegistryNPM  Registry = "npm"
	RegistryPyPI Registry = "pypi"
)

// PackageIdentity identifies a package coordinate within a registry.
type PackageIdentity struct {
	Registry Registry `json:"registry"`
	Name     string   `json:"name"`
	Version  string   `json:"version"`
}

// Validate checks required identity fields without applying registry-specific
// version grammar that belongs in an adapter.
func (p PackageIdentity) Validate() error {
	if p.Registry != RegistryNPM && p.Registry != RegistryPyPI {
		return fmt.Errorf("unsupported registry %q", p.Registry)
	}
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("package name is required")
	}
	if strings.TrimSpace(p.Version) == "" {
		return errors.New("package version is required")
	}
	return nil
}

// Artifact describes one published release file.
type Artifact struct {
	Identity       PackageIdentity `json:"identity"`
	Filename       string          `json:"filename"`
	URL            string          `json:"url"`
	ExpectedSHA256 string          `json:"expected_sha256,omitempty"`
	SHA256         string          `json:"sha256,omitempty"`
	Size           int64           `json:"size,omitempty"`
	ContentType    string          `json:"content_type,omitempty"`
	ExpectedFiles  int             `json:"expected_files,omitempty"`
}

// ReleaseRequest identifies the release an adapter should resolve. An empty
// Version asks the adapter to apply its documented default policy.
type ReleaseRequest struct {
	Registry Registry `json:"registry"`
	Name     string   `json:"name"`
	Version  string   `json:"version,omitempty"`
}

// ReleaseMetadata is the normalized registry result consumed by the rest of
// the pipeline.
type ReleaseMetadata struct {
	Identity   PackageIdentity      `json:"identity"`
	Artifacts  []Artifact           `json:"artifacts"`
	Source     *SourceReference     `json:"source,omitempty"`
	Git        *GitReference        `json:"git,omitempty"`
	Evidence   []Evidence           `json:"evidence,omitempty"`
	Provenance []ProvenanceEvidence `json:"provenance,omitempty"`
}

// RegistryAdapter is the only registry-specific contract required by the
// domain. Adapters translate registry behavior into normalized values.
type RegistryAdapter interface {
	Registry() Registry
	Resolve(context.Context, ReleaseRequest) (ReleaseMetadata, error)
}

// Validate checks artifact fields that are independent of a registry.
func (a Artifact) Validate() error {
	if err := a.Identity.Validate(); err != nil {
		return fmt.Errorf("artifact identity: %w", err)
	}
	if strings.TrimSpace(a.Filename) == "" {
		return errors.New("artifact filename is required")
	}
	if err := validateHTTPSURL(a.URL); err != nil {
		return fmt.Errorf("artifact URL: %w", err)
	}
	if a.Size < 0 {
		return errors.New("artifact size cannot be negative")
	}
	if a.ExpectedSHA256 != "" {
		if len(a.ExpectedSHA256) != sha256HexLength {
			return errors.New("expected SHA-256 must contain 64 hexadecimal characters")
		}
		if _, err := hex.DecodeString(a.ExpectedSHA256); err != nil {
			return errors.New("expected SHA-256 must be hexadecimal")
		}
	}
	if a.ExpectedFiles < 0 {
		return errors.New("expected file count cannot be negative")
	}
	return nil
}

const sha256HexLength = 64

// SourceReference identifies the claimed source repository and optional
// subdirectory. It records the claim, not proof that the claim is correct.
type SourceReference struct {
	URL       string `json:"url"`
	Directory string `json:"directory,omitempty"`
}

// Validate checks that a source URL is absolute and uses HTTPS. Registry
// adapters may normalize ecosystem-specific URL forms before validation.
func (s SourceReference) Validate() error {
	if err := validateHTTPSURL(s.URL); err != nil {
		return fmt.Errorf("source URL: %w", err)
	}
	if strings.ContainsAny(s.Directory, "\x00\r\n") {
		return errors.New("source directory contains invalid characters")
	}
	return nil
}

// GitReference records a claimed or resolved git reference.
type GitReference struct {
	Ref       string `json:"ref"`
	Commit    string `json:"commit,omitempty"`
	Source    string `json:"source,omitempty"`
	Immutable bool   `json:"immutable"`
}

// Validate checks reference shape without requiring a particular hosting
// service or assuming that every ref is a commit.
func (g GitReference) Validate() error {
	if strings.TrimSpace(g.Ref) == "" && strings.TrimSpace(g.Commit) == "" {
		return errors.New("git reference or commit is required")
	}
	if strings.ContainsAny(g.Ref+g.Commit, "\x00\r\n") {
		return errors.New("git reference contains invalid characters")
	}
	if g.Commit != "" && !isHexCommit(g.Commit) {
		return errors.New("git commit must be hexadecimal")
	}
	return nil
}

// FileKind identifies the logical kind of an inventory entry.
type FileKind string

const (
	FileRegular   FileKind = "regular"
	FileSymlink   FileKind = "symlink"
	FileSpecial   FileKind = "special"
	FileDirectory FileKind = "directory"
)

// FileEntry is a deterministic inventory entry. Hash is empty for entries
// that are not regular files or were not safely read.
type FileEntry struct {
	Path   string   `json:"path"`
	Kind   FileKind `json:"kind"`
	SHA256 string   `json:"sha256,omitempty"`
	Size   int64    `json:"size,omitempty"`
	Target string   `json:"target,omitempty"`
}

// ComparisonKind describes the relationship of one path between source and
// artifact inventories.
type ComparisonKind string

const (
	ComparisonIdentical    ComparisonKind = "identical"
	ComparisonSourceOnly   ComparisonKind = "source_only"
	ComparisonArtifactOnly ComparisonKind = "artifact_only"
	ComparisonModified     ComparisonKind = "modified"
	ComparisonTypeChange   ComparisonKind = "type_change"
	ComparisonUnverifiable ComparisonKind = "unverifiable"
)

// FileComparison contains one stable comparison result.
type FileComparison struct {
	Path     string         `json:"path"`
	Kind     ComparisonKind `json:"kind"`
	Source   *FileEntry     `json:"source,omitempty"`
	Artifact *FileEntry     `json:"artifact,omitempty"`
}

// SortComparisons returns a copy ordered by path and then comparison kind.
func SortComparisons(items []FileComparison) []FileComparison {
	result := append([]FileComparison(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Path == result[j].Path {
			return result[i].Kind < result[j].Kind
		}
		return result[i].Path < result[j].Path
	})
	return result
}

// EvidenceState describes how a fact was established.
type EvidenceState string

const (
	EvidenceKnown       EvidenceState = "known"
	EvidenceInferred    EvidenceState = "inferred"
	EvidenceUnavailable EvidenceState = "unavailable"
	EvidenceInvalid     EvidenceState = "invalid"
)

// Evidence is a typed observation with a human-readable explanation. Value
// is intentionally interface-free so JSON output remains stable and explicit.
type Evidence struct {
	ID          string        `json:"id"`
	State       EvidenceState `json:"state"`
	Subject     string        `json:"subject"`
	Value       string        `json:"value,omitempty"`
	Description string        `json:"description,omitempty"`
}

// SecuritySeverity expresses the review priority of a deterministic security
// observation. It is not a probability of maliciousness.
type SecuritySeverity string

const (
	SecurityInfo    SecuritySeverity = "info"
	SecurityWarning SecuritySeverity = "warning"
)

// SecurityObservation records a deterministic structural signal. It never
// asserts that a package is malicious or that an observed script was run.
type SecurityObservation struct {
	ID          string           `json:"id"`
	State       EvidenceState    `json:"state"`
	Severity    SecuritySeverity `json:"severity"`
	Subject     string           `json:"subject"`
	Value       string           `json:"value,omitempty"`
	Description string           `json:"description,omitempty"`
}

// SortSecurityObservations returns a copy ordered for stable reports.
func SortSecurityObservations(items []SecurityObservation) []SecurityObservation {
	result := append([]SecurityObservation(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].ID == result[j].ID {
			return result[i].Subject < result[j].Subject
		}
		return result[i].ID < result[j].ID
	})
	return result
}

// ProvenanceStatus describes the availability and verification state of
// registry-provided provenance or attestation evidence.
type ProvenanceStatus string

const (
	ProvenancePresent      ProvenanceStatus = "present"
	ProvenanceAbsent       ProvenanceStatus = "absent"
	ProvenanceUnavailable  ProvenanceStatus = "unavailable"
	ProvenanceInvalid      ProvenanceStatus = "invalid"
	ProvenanceInsufficient ProvenanceStatus = "insufficient"
)

// ProvenanceEvidence records evidence without asserting that the artifact is
// safe or equivalent to source.
type ProvenanceEvidence struct {
	Status      ProvenanceStatus `json:"status"`
	Predicate   string           `json:"predicate,omitempty"`
	Subject     string           `json:"subject,omitempty"`
	Digest      string           `json:"digest,omitempty"`
	Identity    string           `json:"identity,omitempty"`
	SourceURI   string           `json:"source_uri,omitempty"`
	Commit      string           `json:"commit,omitempty"`
	Explanation string           `json:"explanation,omitempty"`
}

// Verdict is a policy-neutral summary of the evidence state.
type Verdict string

const (
	VerdictMatch      Verdict = "MATCH"
	VerdictReview     Verdict = "REVIEW"
	VerdictIncomplete Verdict = "INCOMPLETE"
	VerdictError      Verdict = "ERROR"
)

// ExitCode is the stable process result used by the future CLI.
type ExitCode int

const (
	ExitSuccess ExitCode = 0
	ExitReview  ExitCode = 1
	ExitUsage   ExitCode = 2
	ExitError   ExitCode = 3
)

// Report is the top-level machine-readable result. GeneratedAt is metadata
// about report creation and should be omitted by deterministic fixture tests.
type Report struct {
	SchemaVersion string                `json:"schema_version"`
	Identity      PackageIdentity       `json:"identity"`
	Artifact      Artifact              `json:"artifact"`
	Source        *SourceReference      `json:"source,omitempty"`
	Git           *GitReference         `json:"git,omitempty"`
	Comparisons   []FileComparison      `json:"comparisons,omitempty"`
	Security      []SecurityObservation `json:"security_observations,omitempty"`
	Evidence      []Evidence            `json:"evidence,omitempty"`
	Provenance    []ProvenanceEvidence  `json:"provenance,omitempty"`
	Limitations   []string              `json:"limitations,omitempty"`
	Verdict       Verdict               `json:"verdict"`
	GeneratedAt   time.Time             `json:"generated_at,omitempty"`
}

// ErrorKind classifies operational and evidence-processing errors.
type ErrorKind string

const (
	ErrorInput      ErrorKind = "input"
	ErrorMetadata   ErrorKind = "metadata"
	ErrorNetwork    ErrorKind = "network"
	ErrorIntegrity  ErrorKind = "integrity"
	ErrorArchive    ErrorKind = "archive_safety"
	ErrorSource     ErrorKind = "source_resolution"
	ErrorProvenance ErrorKind = "provenance"
	ErrorComparison ErrorKind = "comparison"
	ErrorReporting  ErrorKind = "reporting"
	ErrorInternal   ErrorKind = "internal"
)

// DomainError preserves a stable category while retaining the underlying
// cause for diagnostics.
type DomainError struct {
	Kind ErrorKind
	Op   string
	Err  error
}

func (e *DomainError) Error() string {
	if e.Op == "" {
		return fmt.Sprintf("%s: %v", e.Kind, e.Err)
	}
	return fmt.Sprintf("%s: %s: %v", e.Kind, e.Op, e.Err)
}

func (e *DomainError) Unwrap() error { return e.Err }

// NewError creates a categorized domain error.
func NewError(kind ErrorKind, op string, err error) error {
	if err == nil {
		return nil
	}
	return &DomainError{Kind: kind, Op: op, Err: err}
}

// ErrorKindOf returns the category of err, or an empty category when err is
// not a ReleaseCheck domain error.
func ErrorKindOf(err error) ErrorKind {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Kind
	}
	return ""
}

// Cache is intentionally small. Implementations must key entries by immutable
// identity such as a digest or resolved commit, not a mutable branch name.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Put(ctx context.Context, key string, value []byte) error
}

func validateHTTPSURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return errors.New("must be an absolute HTTPS URL")
	}
	return nil
}

func isHexCommit(value string) bool {
	if len(value) < 7 || len(value) > 64 {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}

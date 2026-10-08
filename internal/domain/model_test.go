package domain

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestPackageIdentityValidation(t *testing.T) {
	valid := PackageIdentity{Registry: RegistryNPM, Name: "example", Version: "1.0.0"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid identity rejected: %v", err)
	}

	tests := []PackageIdentity{
		{Name: "example", Version: "1.0.0"},
		{Registry: RegistryNPM, Version: "1.0.0"},
		{Registry: RegistryNPM, Name: "example"},
	}
	for _, test := range tests {
		if err := test.Validate(); err == nil {
			t.Fatalf("invalid identity accepted: %+v", test)
		}
	}
}

func TestArtifactValidationRejectsNonHTTPS(t *testing.T) {
	artifact := Artifact{
		Identity: PackageIdentity{Registry: RegistryPyPI, Name: "example", Version: "1.0"},
		Filename: "example-1.0.tar.gz",
		URL:      "http://files.example.test/example.tar.gz",
	}
	if err := artifact.Validate(); err == nil {
		t.Fatal("HTTP artifact URL was accepted")
	}
}

func TestGitReferenceValidation(t *testing.T) {
	if err := (GitReference{Commit: "0123456789abcdef0123456789abcdef01234567", Immutable: true}).Validate(); err != nil {
		t.Fatalf("valid commit rejected: %v", err)
	}
	if err := (GitReference{Commit: "not-a-commit"}).Validate(); err == nil {
		t.Fatal("invalid commit accepted")
	}
}

func TestSortComparisonsReturnsStableCopy(t *testing.T) {
	input := []FileComparison{
		{Path: "z", Kind: ComparisonModified},
		{Path: "a", Kind: ComparisonArtifactOnly},
		{Path: "a", Kind: ComparisonIdentical},
	}
	got := SortComparisons(input)
	want := []FileComparison{
		{Path: "a", Kind: ComparisonArtifactOnly},
		{Path: "a", Kind: ComparisonIdentical},
		{Path: "z", Kind: ComparisonModified},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected ordering: got %#v want %#v", got, want)
	}
	if input[0].Path != "z" {
		t.Fatal("SortComparisons mutated input")
	}
}

func TestReportJSONRoundTrip(t *testing.T) {
	report := Report{
		SchemaVersion: "0.1",
		Identity:      PackageIdentity{Registry: RegistryNPM, Name: "example", Version: "1.0.0"},
		Artifact: Artifact{
			Identity: PackageIdentity{Registry: RegistryNPM, Name: "example", Version: "1.0.0"},
			Filename: "example-1.0.0.tgz",
			URL:      "https://registry.example.test/example.tgz",
		},
		Verdict: VerdictReview,
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	var decoded Report
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}
	if !reflect.DeepEqual(report, decoded) {
		t.Fatalf("round trip changed report: got %#v want %#v", decoded, report)
	}
}

func TestErrorClassification(t *testing.T) {
	original := errors.New("bad archive")
	err := NewError(ErrorArchive, "inspect", original)
	if ErrorKindOf(err) != ErrorArchive {
		t.Fatalf("unexpected error kind: %q", ErrorKindOf(err))
	}
	if !errors.Is(err, original) {
		t.Fatal("domain error did not preserve cause")
	}
}

func TestCacheContractAcceptsStandardContext(t *testing.T) {
	var cache Cache = contextCache{}
	if _, ok, err := cache.Get(context.Background(), "sha256:abc"); err != nil || ok {
		t.Fatalf("unexpected empty cache result: ok=%v err=%v", ok, err)
	}
}

type contextCache struct{}

func (contextCache) Get(context.Context, string) ([]byte, bool, error) { return nil, false, nil }
func (contextCache) Put(context.Context, string, []byte) error         { return nil }

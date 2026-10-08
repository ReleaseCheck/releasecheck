package tests

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/releasecheck/releasecheck/internal/acquire"
	"github.com/releasecheck/releasecheck/internal/compare"
	"github.com/releasecheck/releasecheck/internal/domain"
	"github.com/releasecheck/releasecheck/internal/provenance"
	"github.com/releasecheck/releasecheck/internal/report"
	"github.com/releasecheck/releasecheck/internal/security"
)

func TestOfflineFixtureMatrix(t *testing.T) {
	hashA := strings.Repeat("a", 64)
	hashB := strings.Repeat("b", 64)
	entry := func(path, hash string) domain.FileEntry {
		return domain.FileEntry{Path: path, Kind: domain.FileRegular, SHA256: hash}
	}

	t.Run("matching source and artifact", func(t *testing.T) {
		comparisons, err := compare.Compare([]domain.FileEntry{entry("index.js", hashA)}, []domain.FileEntry{entry("index.js", hashA)})
		if err != nil || len(comparisons) != 1 || comparisons[0].Kind != domain.ComparisonIdentical {
			t.Fatalf("unexpected match: %#v, %v", comparisons, err)
		}
	})
	t.Run("extra artifact file", func(t *testing.T) {
		comparisons, err := compare.Compare(nil, []domain.FileEntry{entry("extra.js", hashA)})
		if err != nil || comparisons[0].Kind != domain.ComparisonArtifactOnly {
			t.Fatalf("unexpected artifact-only result: %#v, %v", comparisons, err)
		}
	})
	t.Run("missing source file", func(t *testing.T) {
		comparisons, err := compare.Compare([]domain.FileEntry{entry("source.py", hashA)}, nil)
		if err != nil || comparisons[0].Kind != domain.ComparisonSourceOnly {
			t.Fatalf("unexpected source-only result: %#v, %v", comparisons, err)
		}
	})
	t.Run("modified file", func(t *testing.T) {
		comparisons, err := compare.Compare([]domain.FileEntry{entry("index.js", hashA)}, []domain.FileEntry{entry("index.js", hashB)})
		if err != nil || comparisons[0].Kind != domain.ComparisonModified {
			t.Fatalf("unexpected modified result: %#v, %v", comparisons, err)
		}
	})
	t.Run("malformed package", func(t *testing.T) {
		if _, err := inspectBytes(t, []byte("not an archive"), acquire.DefaultLimits()); err == nil {
			t.Fatal("malformed package was accepted")
		}
	})
	t.Run("suspicious install metadata", func(t *testing.T) {
		observations := security.AnalyzeNPMManifest([]byte(`{"scripts":{"postinstall":"node setup.js"}}`))
		if len(observations) != 1 || observations[0].ID != "npm.install_script" {
			t.Fatalf("unexpected metadata observation: %#v", observations)
		}
	})
	t.Run("missing repository metadata", func(t *testing.T) {
		result := domain.ReleaseMetadata{Evidence: []domain.Evidence{{ID: "npm.repository.unavailable", State: domain.EvidenceUnavailable, Subject: "source.repository"}}}
		if len(result.Evidence) != 1 || result.Evidence[0].State != domain.EvidenceUnavailable {
			t.Fatalf("missing repository state was not preserved: %#v", result)
		}
	})
	t.Run("unavailable source reference", func(t *testing.T) {
		evidence := provenance.Absent("pypi", "source provenance is unavailable")
		if evidence.Status != domain.ProvenanceAbsent {
			t.Fatalf("unavailable source state was not explicit: %#v", evidence)
		}
	})
	t.Run("unsafe archive path", func(t *testing.T) {
		archive := zipBytes(t, []zipFixture{{name: "../escape", content: []byte("x")}})
		if _, err := inspectBytes(t, archive, acquire.DefaultLimits()); err == nil {
			t.Fatal("unsafe archive path was accepted")
		}
	})
	t.Run("symlink edge case", func(t *testing.T) {
		inventory := acquire.ArchiveInventory{Entries: []domain.FileEntry{{Path: "link", Kind: domain.FileSymlink, Target: "../outside"}}}
		observations, err := security.AnalyzeInventory(inventory, security.Options{})
		if err != nil || !hasObservation(observations, "archive.symlink_target_unsafe") {
			t.Fatalf("symlink signal missing: %#v, %v", observations, err)
		}
	})
	t.Run("large archive boundary", func(t *testing.T) {
		archive := tarGzipBytes(t, []tarFixture{{name: "large.bin", content: []byte("1234")}})
		limits := acquire.DefaultLimits()
		limits.MaxFileBytes = 3
		if _, err := inspectBytes(t, archive, limits); err == nil {
			t.Fatal("oversized archive member was accepted")
		}
	})
	t.Run("provenance variations", func(t *testing.T) {
		cases := []struct {
			name  string
			got   domain.ProvenanceStatus
			input func() []domain.ProvenanceEvidence
		}{
			{name: "unavailable", got: domain.ProvenanceUnavailable, input: func() []domain.ProvenanceEvidence { return provenance.ParseNPM(nil, provenance.Binding{}) }},
			{name: "absent", got: domain.ProvenanceAbsent, input: func() []domain.ProvenanceEvidence {
				return []domain.ProvenanceEvidence{provenance.Absent("npm", "none")}
			}},
			{name: "invalid", got: domain.ProvenanceInvalid, input: func() []domain.ProvenanceEvidence { return provenance.ParsePyPI([]byte("{"), provenance.Binding{}) }},
		}
		for _, testCase := range cases {
			t.Run(testCase.name, func(t *testing.T) {
				got := testCase.input()
				if len(got) != 1 || got[0].Status != testCase.got {
					t.Fatalf("got %#v want %s", got, testCase.got)
				}
			})
		}
	})

	// Keep the report package in this offline matrix so the fixture contract
	// proves that evidence can be rendered without registry access.
	t.Run("report remains offline", func(t *testing.T) {
		identity := domain.PackageIdentity{Registry: domain.RegistryNPM, Name: "fixture", Version: "1.0.0"}
		input := domain.Report{Identity: identity, Artifact: domain.Artifact{Identity: identity, Filename: "fixture.tgz", URL: "https://example.test/fixture.tgz"}}
		if _, err := report.JSON(input); err != nil {
			t.Fatal(err)
		}
	})
}

type zipFixture struct {
	name    string
	content []byte
}

func zipBytes(t *testing.T, entries []zipFixture) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, fixture := range entries {
		entry, err := writer.Create(fixture.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(fixture.content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

type tarFixture struct {
	name    string
	content []byte
}

func tarGzipBytes(t *testing.T, entries []tarFixture) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, fixture := range entries {
		if err := tarWriter.WriteHeader(&tar.Header{Name: fixture.name, Mode: 0600, Size: int64(len(fixture.content))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write(fixture.content); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func inspectBytes(t *testing.T, content []byte, limits acquire.Limits) (acquire.ArchiveInventory, error) {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "fixture.archive")
	if err := os.WriteFile(filename, content, 0600); err != nil {
		t.Fatal(err)
	}
	return acquire.InspectArchive(filename, limits)
}

func hasObservation(items []domain.SecurityObservation, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

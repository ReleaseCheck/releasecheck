package security

import (
	"strings"
	"testing"

	"github.com/releasecheck/releasecheck/internal/acquire"
	"github.com/releasecheck/releasecheck/internal/domain"
)

func TestAnalyzeInventoryReportsStructuralSignalsWithoutFollowingLinks(t *testing.T) {
	observations, err := AnalyzeInventory(acquire.ArchiveInventory{
		Kind:            acquire.ArchiveTarGzip,
		CompressedBytes: 20,
		ExpandedBytes:   200,
		Entries: []domain.FileEntry{
			{Path: "link", Kind: domain.FileSymlink, Target: "../outside"},
			{Path: "device", Kind: domain.FileSpecial},
			{Path: "large.bin", Kind: domain.FileRegular, Size: 100},
		},
	}, Options{LargeFileBytes: 100, LargeArchiveBytes: 150})
	if err != nil {
		t.Fatal(err)
	}

	ids := make(map[string]bool)
	for _, observation := range observations {
		ids[observation.ID] = true
		if observation.State != domain.EvidenceKnown {
			t.Fatalf("unexpected state for %#v", observation)
		}
	}
	for _, id := range []string{"archive.large", "archive.symlink", "archive.symlink_target_unsafe", "archive.special_entry", "archive.large_file"} {
		if !ids[id] {
			t.Fatalf("missing observation %q: %#v", id, observations)
		}
	}
	for i := 1; i < len(observations); i++ {
		if observations[i-1].ID > observations[i].ID {
			t.Fatalf("observations are not stable: %#v", observations)
		}
	}
}

func TestAnalyzeNPMManifestReportsOnlyInstallLifecycleScripts(t *testing.T) {
	data := []byte(`{"scripts":{"test":"go test ./...","postinstall":"node setup.js","prepare":"echo prepare"}}`)
	observations := AnalyzeNPMManifest(data)
	if len(observations) != 2 {
		t.Fatalf("got %#v", observations)
	}
	for _, observation := range observations {
		if observation.ID != "npm.install_script" || observation.State != domain.EvidenceKnown || observation.Severity != domain.SecurityWarning {
			t.Fatalf("unexpected observation: %#v", observation)
		}
		if strings.Contains(observation.Description, "run") == false {
			t.Fatalf("missing non-execution boundary: %#v", observation)
		}
	}
}

func TestAnalyzeMetadataReportsMalformedInputAndDoesNotClaimMalware(t *testing.T) {
	malformed := AnalyzeNPMManifest([]byte("{"))
	if len(malformed) != 1 || malformed[0].ID != "npm.manifest.malformed" || malformed[0].State != domain.EvidenceInvalid {
		t.Fatalf("unexpected malformed result: %#v", malformed)
	}

	setup := AnalyzePythonMetadata("project/setup.py", []byte("print('not executed')"))
	if len(setup) != 1 || setup[0].ID != "python.setup_py_present" {
		t.Fatalf("unexpected setup.py result: %#v", setup)
	}
	if strings.Contains(strings.ToLower(setup[0].Description), "malware") {
		t.Fatal("security observation made an unsupported malware claim")
	}

	emptyProject := AnalyzePythonMetadata("pyproject.toml", nil)
	if len(emptyProject) != 1 || emptyProject[0].ID != "python.pyproject.malformed" {
		t.Fatalf("unexpected empty pyproject result: %#v", emptyProject)
	}
}

func TestAnalyzeInventoryRejectsInvalidThresholds(t *testing.T) {
	if _, err := AnalyzeInventory(acquire.ArchiveInventory{}, Options{LargeFileBytes: -1}); err == nil {
		t.Fatal("invalid threshold was accepted")
	}
}

package compare

import (
	"strings"
	"testing"

	"github.com/releasecheck/releasecheck/internal/acquire"
	"github.com/releasecheck/releasecheck/internal/domain"
)

func TestCompareAllCoreCategories(t *testing.T) {
	sourceHash := strings.Repeat("a", 64)
	artifactHash := strings.Repeat("b", 64)
	source := []domain.FileEntry{
		{Path: "same", Kind: domain.FileRegular, SHA256: sourceHash},
		{Path: "modified", Kind: domain.FileRegular, SHA256: sourceHash},
		{Path: "source-only", Kind: domain.FileRegular, SHA256: sourceHash},
		{Path: "type", Kind: domain.FileRegular, SHA256: sourceHash},
		{Path: "unknown", Kind: domain.FileRegular},
		{Path: "link", Kind: domain.FileSymlink, Target: "target-a"},
	}
	artifact := []domain.FileEntry{
		{Path: "same", Kind: domain.FileRegular, SHA256: sourceHash},
		{Path: "modified", Kind: domain.FileRegular, SHA256: artifactHash},
		{Path: "artifact-only", Kind: domain.FileRegular, SHA256: artifactHash},
		{Path: "type", Kind: domain.FileSymlink, Target: "target"},
		{Path: "unknown", Kind: domain.FileRegular, SHA256: artifactHash},
		{Path: "link", Kind: domain.FileSymlink, Target: "target-b"},
	}

	got, err := Compare(source, artifact)
	if err != nil {
		t.Fatalf("Compare returned error: %v", err)
	}
	gotKinds := make(map[string]domain.ComparisonKind, len(got))
	for _, item := range got {
		gotKinds[item.Path] = item.Kind
	}
	want := map[string]domain.ComparisonKind{
		"same":          domain.ComparisonIdentical,
		"modified":      domain.ComparisonModified,
		"source-only":   domain.ComparisonSourceOnly,
		"artifact-only": domain.ComparisonArtifactOnly,
		"type":          domain.ComparisonTypeChange,
		"unknown":       domain.ComparisonUnverifiable,
		"link":          domain.ComparisonModified,
	}
	for entryPath, kind := range want {
		if gotKinds[entryPath] != kind {
			t.Fatalf("path %q: got %q want %q", entryPath, gotKinds[entryPath], kind)
		}
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Path > got[i].Path {
			t.Fatalf("comparison output is not sorted: %#v", got)
		}
	}

	summary := Summarize(got)
	wantSummary := Summary{
		Total:        7,
		Identical:    1,
		SourceOnly:   1,
		ArtifactOnly: 1,
		Modified:     2,
		TypeChanges:  1,
		Unverifiable: 1,
	}
	if summary != wantSummary {
		t.Fatalf("summary: got %#v want %#v", summary, wantSummary)
	}
}

func TestInventoryStripsRootAndRejectsUnsafeOrDuplicateEntries(t *testing.T) {
	input := []domain.FileEntry{{Path: "package/index.js", Kind: domain.FileRegular}}
	got, err := Inventory(structInventory(input), "package")
	if err != nil || len(got) != 1 || got[0].Path != "index.js" {
		t.Fatalf("unexpected normalized inventory: %#v, %v", got, err)
	}

	for _, entryPath := range []string{
		"other/index.js",
		"package/../escape",
		"/absolute",
		"package\\file",
		"package/./file",
	} {
		if _, err := Inventory(structInventory([]domain.FileEntry{{Path: entryPath}}), "package"); err == nil {
			t.Fatalf("unsafe path %q was accepted", entryPath)
		}
	}

	duplicates := []domain.FileEntry{{Path: "package/a"}, {Path: "package/a"}}
	if _, err := Inventory(structInventory(duplicates), "package"); err == nil {
		t.Fatal("duplicate normalized paths were accepted")
	}
}

func TestCompareRejectsInvalidDirectInput(t *testing.T) {
	if _, err := Compare([]domain.FileEntry{{Path: "a/../b", Kind: domain.FileRegular}}, nil); err == nil {
		t.Fatal("non-canonical path was accepted")
	}
}

func structInventory(entries []domain.FileEntry) acquire.ArchiveInventory {
	return acquire.ArchiveInventory{Entries: entries}
}

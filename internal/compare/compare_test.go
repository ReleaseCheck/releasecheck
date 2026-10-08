package compare

import (
	"testing"

	"github.com/releasecheck/releasecheck/internal/acquire"
	"github.com/releasecheck/releasecheck/internal/domain"
)

func TestCompareAllCoreCategories(t *testing.T) {
	sourceHash := "source"
	artifactHash := "artifact"
	source := []domain.FileEntry{
		{Path: "same", Kind: domain.FileRegular, SHA256: sourceHash},
		{Path: "modified", Kind: domain.FileRegular, SHA256: sourceHash},
		{Path: "source-only", Kind: domain.FileRegular, SHA256: sourceHash},
		{Path: "type", Kind: domain.FileRegular, SHA256: sourceHash},
	}
	artifact := []domain.FileEntry{
		{Path: "same", Kind: domain.FileRegular, SHA256: sourceHash},
		{Path: "modified", Kind: domain.FileRegular, SHA256: artifactHash},
		{Path: "artifact-only", Kind: domain.FileRegular, SHA256: artifactHash},
		{Path: "type", Kind: domain.FileSymlink},
	}

	got := Compare(source, artifact)
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
	}
	for path, kind := range want {
		if gotKinds[path] != kind {
			t.Fatalf("path %q: got %q want %q", path, gotKinds[path], kind)
		}
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Path > got[i].Path {
			t.Fatalf("comparison output is not sorted: %#v", got)
		}
	}
}

func TestInventoryStripsRootAndRejectsOutsideEntries(t *testing.T) {
	input := []domain.FileEntry{{Path: "package/index.js", Kind: domain.FileRegular}}
	got, err := Inventory(structInventory(input), "package")
	if err != nil || len(got) != 1 || got[0].Path != "index.js" {
		t.Fatalf("unexpected normalized inventory: %#v, %v", got, err)
	}

	if _, err := Inventory(structInventory([]domain.FileEntry{{Path: "other/index.js"}}), "package"); err == nil {
		t.Fatal("entry outside root was accepted")
	}
}

func structInventory(entries []domain.FileEntry) acquire.ArchiveInventory {
	return acquire.ArchiveInventory{Entries: entries}
}

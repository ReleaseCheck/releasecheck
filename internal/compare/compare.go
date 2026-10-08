// Package compare performs deterministic comparisons of validated inventories.
package compare

import (
	"encoding/hex"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/releasecheck/releasecheck/internal/acquire"
	"github.com/releasecheck/releasecheck/internal/domain"
)

// Inventory returns a validated, sorted copy whose paths have an optional
// archive-root prefix removed. Root removal is the only path transformation
// performed by this package. Backslashes, traversal, absolute paths, dot
// segments, and duplicate normalized paths are rejected rather than hidden.
func Inventory(source acquire.ArchiveInventory, rootPrefix string) ([]domain.FileEntry, error) {
	rootPrefix = strings.Trim(rootPrefix, "/")
	if rootPrefix != "" {
		if err := validateCanonicalPath(rootPrefix); err != nil {
			return nil, fmt.Errorf("invalid archive root %q: %w", rootPrefix, err)
		}
	}

	result := make([]domain.FileEntry, 0, len(source.Entries))
	seen := make(map[string]struct{}, len(source.Entries))
	for _, entry := range source.Entries {
		entryPath, err := validateEntry(entry)
		if err != nil {
			return nil, err
		}
		if rootPrefix != "" {
			if entryPath == rootPrefix {
				continue
			}
			prefix := rootPrefix + "/"
			if !strings.HasPrefix(entryPath, prefix) {
				return nil, fmt.Errorf("archive entry %q is outside root prefix %q", entry.Path, rootPrefix)
			}
			entryPath = strings.TrimPrefix(entryPath, prefix)
		}
		if entryPath == "" {
			return nil, fmt.Errorf("archive entry %q has an empty normalized path", entry.Path)
		}
		if _, exists := seen[entryPath]; exists {
			return nil, fmt.Errorf("normalized archive path %q is duplicated", entryPath)
		}
		seen[entryPath] = struct{}{}
		entry.Path = entryPath
		result = append(result, entry)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

// Compare validates both inputs and returns path-ordered relationships. It
// never infers intent from differences and never calls a missing hash equal.
func Compare(source, artifact []domain.FileEntry) ([]domain.FileComparison, error) {
	normalizedSource, err := normalizeEntries(source)
	if err != nil {
		return nil, fmt.Errorf("source inventory: %w", err)
	}
	normalizedArtifact, err := normalizeEntries(artifact)
	if err != nil {
		return nil, fmt.Errorf("artifact inventory: %w", err)
	}
	sourceByPath := entriesByPath(normalizedSource)
	artifactByPath := entriesByPath(normalizedArtifact)
	paths := make(map[string]struct{}, len(sourceByPath)+len(artifactByPath))
	for entryPath := range sourceByPath {
		paths[entryPath] = struct{}{}
	}
	for entryPath := range artifactByPath {
		paths[entryPath] = struct{}{}
	}

	result := make([]domain.FileComparison, 0, len(paths))
	for entryPath := range paths {
		sourceEntry, inSource := sourceByPath[entryPath]
		artifactEntry, inArtifact := artifactByPath[entryPath]
		comparison := domain.FileComparison{Path: entryPath}
		if inSource {
			copy := sourceEntry
			comparison.Source = &copy
		}
		if inArtifact {
			copy := artifactEntry
			comparison.Artifact = &copy
		}
		switch {
		case !inSource:
			comparison.Kind = domain.ComparisonArtifactOnly
		case !inArtifact:
			comparison.Kind = domain.ComparisonSourceOnly
		case sourceEntry.Kind != artifactEntry.Kind:
			comparison.Kind = domain.ComparisonTypeChange
		case sourceEntry.Kind == domain.FileRegular:
			comparison.Kind = compareHashes(sourceEntry.SHA256, artifactEntry.SHA256)
		case sourceEntry.Kind == domain.FileSymlink:
			if sourceEntry.Target == "" || artifactEntry.Target == "" {
				comparison.Kind = domain.ComparisonUnverifiable
			} else if sourceEntry.Target == artifactEntry.Target {
				comparison.Kind = domain.ComparisonIdentical
			} else {
				comparison.Kind = domain.ComparisonModified
			}
		default:
			comparison.Kind = domain.ComparisonIdentical
		}
		result = append(result, comparison)
	}
	return domain.SortComparisons(result), nil
}

// Summary is a deterministic count of comparison categories.
type Summary struct {
	Total        int `json:"total"`
	Identical    int `json:"identical"`
	SourceOnly   int `json:"source_only"`
	ArtifactOnly int `json:"artifact_only"`
	Modified     int `json:"modified"`
	TypeChanges  int `json:"type_changes"`
	Unverifiable int `json:"unverifiable"`
}

// Summarize counts comparison categories without changing their meaning.
func Summarize(items []domain.FileComparison) Summary {
	result := Summary{Total: len(items)}
	for _, item := range items {
		switch item.Kind {
		case domain.ComparisonIdentical:
			result.Identical++
		case domain.ComparisonSourceOnly:
			result.SourceOnly++
		case domain.ComparisonArtifactOnly:
			result.ArtifactOnly++
		case domain.ComparisonModified:
			result.Modified++
		case domain.ComparisonTypeChange:
			result.TypeChanges++
		case domain.ComparisonUnverifiable:
			result.Unverifiable++
		}
	}
	return result
}

func normalizeEntries(entries []domain.FileEntry) ([]domain.FileEntry, error) {
	result := make([]domain.FileEntry, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		var err error
		entry.Path, err = validateEntry(entry)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[entry.Path]; exists {
			return nil, fmt.Errorf("duplicate inventory path %q", entry.Path)
		}
		seen[entry.Path] = struct{}{}
		result = append(result, entry)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

func validateEntry(entry domain.FileEntry) (string, error) {
	if err := validateCanonicalPath(entry.Path); err != nil {
		return "", fmt.Errorf("invalid inventory path %q: %w", entry.Path, err)
	}
	switch entry.Kind {
	case domain.FileRegular, domain.FileDirectory, domain.FileSymlink, domain.FileSpecial:
	default:
		return "", fmt.Errorf("inventory path %q has unsupported file kind %q", entry.Path, entry.Kind)
	}
	if entry.Size < 0 {
		return "", fmt.Errorf("inventory path %q has negative size", entry.Path)
	}
	if entry.SHA256 != "" {
		if len(entry.SHA256) != 64 {
			return "", fmt.Errorf("inventory path %q has invalid SHA-256 length", entry.Path)
		}
		if _, err := hex.DecodeString(entry.SHA256); err != nil {
			return "", fmt.Errorf("inventory path %q has invalid SHA-256: %w", entry.Path, err)
		}
	}
	return entry.Path, nil
}

func validateCanonicalPath(value string) error {
	if value == "" || strings.ContainsAny(value, "\\\x00\r\n") {
		return fmt.Errorf("path is empty or contains unsupported characters")
	}
	if strings.HasPrefix(value, "/") || path.IsAbs(value) || filepath.VolumeName(value) != "" {
		return fmt.Errorf("path is absolute")
	}
	if value == "." || value == ".." || strings.HasPrefix(value, "../") {
		return fmt.Errorf("path escapes its root")
	}
	if path.Clean(value) != value {
		return fmt.Errorf("path is not canonical")
	}
	return nil
}

func compareHashes(source, artifact string) domain.ComparisonKind {
	if source == "" || artifact == "" {
		return domain.ComparisonUnverifiable
	}
	if strings.EqualFold(source, artifact) {
		return domain.ComparisonIdentical
	}
	return domain.ComparisonModified
}

func entriesByPath(entries []domain.FileEntry) map[string]domain.FileEntry {
	result := make(map[string]domain.FileEntry, len(entries))
	for _, entry := range entries {
		result[entry.Path] = entry
	}
	return result
}

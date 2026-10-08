// Package compare performs deterministic comparisons of validated inventories.
package compare

import (
	"fmt"
	"strings"

	"github.com/releasecheck/releasecheck/internal/acquire"
	"github.com/releasecheck/releasecheck/internal/domain"
)

// Inventory returns a copy whose paths have a required archive-root prefix
// removed. The prefix is a logical path component such as npm's "package".
func Inventory(source acquire.ArchiveInventory, rootPrefix string) ([]domain.FileEntry, error) {
	rootPrefix = strings.Trim(strings.ReplaceAll(rootPrefix, "\\", "/"), "/")
	result := make([]domain.FileEntry, 0, len(source.Entries))
	seen := make(map[string]struct{}, len(source.Entries))
	for _, entry := range source.Entries {
		path := strings.ReplaceAll(entry.Path, "\\", "/")
		if path == rootPrefix {
			continue
		}
		prefix := rootPrefix + "/"
		if rootPrefix != "" && !strings.HasPrefix(path, prefix) {
			return nil, fmt.Errorf("archive entry %q is outside root prefix %q", entry.Path, rootPrefix)
		}
		if rootPrefix == "" {
			entry.Path = path
		} else {
			entry.Path = strings.TrimPrefix(path, prefix)
		}
		if entry.Path == "" {
			continue
		}
		if _, exists := seen[entry.Path]; exists {
			return nil, fmt.Errorf("normalized archive path %q is duplicated", entry.Path)
		}
		seen[entry.Path] = struct{}{}
		result = append(result, entry)
	}
	return result, nil
}

// Compare returns deterministic path-ordered relationships between source and
// artifact inventories. It does not infer intent from differences.
func Compare(source, artifact []domain.FileEntry) []domain.FileComparison {
	sourceByPath := entriesByPath(source)
	artifactByPath := entriesByPath(artifact)
	paths := make(map[string]struct{}, len(sourceByPath)+len(artifactByPath))
	for path := range sourceByPath {
		paths[path] = struct{}{}
	}
	for path := range artifactByPath {
		paths[path] = struct{}{}
	}

	result := make([]domain.FileComparison, 0, len(paths))
	for path := range paths {
		sourceEntry, inSource := sourceByPath[path]
		artifactEntry, inArtifact := artifactByPath[path]
		comparison := domain.FileComparison{Path: path}
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
		case sourceEntry.Kind != domain.FileRegular || sourceEntry.SHA256 == artifactEntry.SHA256:
			comparison.Kind = domain.ComparisonIdentical
		default:
			comparison.Kind = domain.ComparisonModified
		}
		result = append(result, comparison)
	}
	return domain.SortComparisons(result)
}

func entriesByPath(entries []domain.FileEntry) map[string]domain.FileEntry {
	result := make(map[string]domain.FileEntry, len(entries))
	for _, entry := range entries {
		result[entry.Path] = entry
	}
	return result
}

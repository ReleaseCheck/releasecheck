// Package security produces deterministic security-relevant observations from
// acquired release structure and explicitly supplied package metadata.
package security

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/releasecheck/releasecheck/internal/acquire"
	"github.com/releasecheck/releasecheck/internal/domain"
)

const (
	defaultLargeFileBytes    int64 = 10 << 20
	defaultLargeArchiveBytes int64 = 100 << 20
	maxManifestBytes         int   = 16 << 20
)

// Options bounds security observations. These thresholds create review
// signals; they do not replace acquisition limits or prove abuse.
type Options struct {
	LargeFileBytes    int64
	LargeArchiveBytes int64
}

func (o Options) withDefaults() Options {
	if o.LargeFileBytes == 0 {
		o.LargeFileBytes = defaultLargeFileBytes
	}
	if o.LargeArchiveBytes == 0 {
		o.LargeArchiveBytes = defaultLargeArchiveBytes
	}
	return o
}

func (o Options) validate() error {
	if o.LargeFileBytes < 1 || o.LargeArchiveBytes < 1 {
		return fmt.Errorf("security thresholds must be positive")
	}
	return nil
}

// AnalyzeInventory reports structural observations from an already validated
// archive inventory. It never follows links or reads package code.
func AnalyzeInventory(inventory acquire.ArchiveInventory, options Options) ([]domain.SecurityObservation, error) {
	options = options.withDefaults()
	if err := options.validate(); err != nil {
		return nil, err
	}
	observations := make([]domain.SecurityObservation, 0)
	if inventory.CompressedBytes >= options.LargeArchiveBytes || inventory.ExpandedBytes >= options.LargeArchiveBytes {
		observations = append(observations, domain.SecurityObservation{
			ID:          "archive.large",
			State:       domain.EvidenceKnown,
			Severity:    domain.SecurityWarning,
			Subject:     "artifact.archive",
			Value:       fmt.Sprintf("compressed=%d expanded=%d", inventory.CompressedBytes, inventory.ExpandedBytes),
			Description: "Archive size crosses the configured review threshold; this is a resource-risk signal, not evidence of maliciousness.",
		})
	}
	for _, entry := range inventory.Entries {
		switch entry.Kind {
		case domain.FileSymlink:
			observations = append(observations, domain.SecurityObservation{
				ID:          "archive.symlink",
				State:       domain.EvidenceKnown,
				Severity:    domain.SecurityWarning,
				Subject:     entry.Path,
				Value:       entry.Target,
				Description: "Archive contains a symlink; ReleaseCheck records its target and never follows it.",
			})
			if unsafeLinkTarget(entry.Target) {
				observations = append(observations, domain.SecurityObservation{
					ID:          "archive.symlink_target_unsafe",
					State:       domain.EvidenceKnown,
					Severity:    domain.SecurityWarning,
					Subject:     entry.Path,
					Value:       entry.Target,
					Description: "Symlink target is absolute or escapes its containing path; the target was not followed.",
				})
			}
		case domain.FileSpecial:
			observations = append(observations, domain.SecurityObservation{
				ID:          "archive.special_entry",
				State:       domain.EvidenceKnown,
				Severity:    domain.SecurityWarning,
				Subject:     entry.Path,
				Description: "Archive contains a special filesystem entry; it was observed without materializing or executing it.",
			})
		}
		if entry.Size >= options.LargeFileBytes {
			observations = append(observations, domain.SecurityObservation{
				ID:          "archive.large_file",
				State:       domain.EvidenceKnown,
				Severity:    domain.SecurityWarning,
				Subject:     entry.Path,
				Value:       fmt.Sprintf("%d", entry.Size),
				Description: "Archive member crosses the configured review threshold; size alone does not establish intent.",
			})
		}
	}
	return domain.SortSecurityObservations(observations), nil
}

// AnalyzeNPMManifest examines package.json bytes supplied by a safe archive
// reader. It never executes script values. Malformed JSON is represented as an
// invalid observation so callers can report the limitation deterministically.
func AnalyzeNPMManifest(data []byte) []domain.SecurityObservation {
	if len(data) > maxManifestBytes {
		return []domain.SecurityObservation{invalidMetadata("npm.manifest.oversized", "package.json", "package.json exceeds the metadata limit")}
	}
	var manifest struct {
		Scripts map[string]json.RawMessage `json:"scripts"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return []domain.SecurityObservation{invalidMetadata("npm.manifest.malformed", "package.json", err.Error())}
	}
	observations := make([]domain.SecurityObservation, 0)
	for name, raw := range manifest.Scripts {
		var command string
		if err := json.Unmarshal(raw, &command); err != nil {
			observations = append(observations, invalidMetadata("npm.script.malformed", "package.json.scripts."+name, "script value is not a string"))
			continue
		}
		if isInstallLifecycleScript(name) {
			observations = append(observations, domain.SecurityObservation{
				ID:          "npm.install_script",
				State:       domain.EvidenceKnown,
				Severity:    domain.SecurityWarning,
				Subject:     "package.json.scripts." + name,
				Value:       command,
				Description: "npm metadata declares a lifecycle script. ReleaseCheck records it but never runs npm or the command.",
			})
		}
	}
	return domain.SortSecurityObservations(observations)
}

// AnalyzePythonMetadata records Python build/install metadata by filename and
// content shape. It does not parse or execute setup.py or a build backend.
func AnalyzePythonMetadata(filename string, data []byte) []domain.SecurityObservation {
	if len(data) > maxManifestBytes {
		return []domain.SecurityObservation{invalidMetadata("python.metadata.oversized", filename, "metadata exceeds the metadata limit")}
	}
	base := strings.ToLower(path.Base(filename))
	if base == "setup.py" {
		return []domain.SecurityObservation{{
			ID:          "python.setup_py_present",
			State:       domain.EvidenceKnown,
			Severity:    domain.SecurityWarning,
			Subject:     filename,
			Description: "setup.py may contain executable build/install logic. ReleaseCheck records its presence and never executes it.",
		}}
	}
	if base == "pyproject.toml" && len(strings.TrimSpace(string(data))) == 0 {
		return []domain.SecurityObservation{invalidMetadata("python.pyproject.malformed", filename, "pyproject.toml is empty")}
	}
	return nil
}

func invalidMetadata(id, subject, description string) domain.SecurityObservation {
	return domain.SecurityObservation{ID: id, State: domain.EvidenceInvalid, Severity: domain.SecurityWarning, Subject: subject, Description: description}
}

func isInstallLifecycleScript(name string) bool {
	switch strings.ToLower(name) {
	case "preinstall", "install", "postinstall", "prepare", "prepublish", "prepublishonly":
		return true
	default:
		return false
	}
}

func unsafeLinkTarget(target string) bool {
	if target == "" || strings.ContainsAny(target, "\x00\r\n") || strings.HasPrefix(target, "/") || strings.HasPrefix(target, "\\") {
		return true
	}
	cleaned := path.Clean(strings.ReplaceAll(target, "\\", "/"))
	return cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(path.Base(target), ":")
}

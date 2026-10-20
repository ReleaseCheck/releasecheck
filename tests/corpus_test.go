package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

type corpus struct {
	SchemaVersion string   `json:"schema_version"`
	ObservedAt    string   `json:"observed_at"`
	Records       []record `json:"records"`
}

type record struct {
	Ecosystem        string     `json:"ecosystem"`
	Package          string     `json:"package"`
	Version          string     `json:"version"`
	MetadataURL      string     `json:"metadata_url"`
	ArtifactURL      string     `json:"artifact_url"`
	ArtifactSHA256   string     `json:"artifact_sha256"`
	SourceRepository string     `json:"source_repository"`
	Artifacts        []artifact `json:"artifacts"`
}

type artifact struct {
	ArtifactURL    string `json:"artifact_url"`
	ArtifactSHA256 string `json:"artifact_sha256"`
}

func TestRealReleaseCorpusIsPinnedAndWellFormed(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "fixtures", "corpus", "real-releases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got corpus
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("corpus JSON: %v", err)
	}
	if got.SchemaVersion != "1.0" || got.ObservedAt == "" || len(got.Records) != 2 {
		t.Fatalf("unexpected corpus header: %#v", got)
	}
	for _, item := range got.Records {
		if item.Package == "" || item.Version == "" || item.SourceRepository == "" {
			t.Fatalf("incomplete record: %#v", item)
		}
		for _, raw := range []string{item.MetadataURL, item.SourceRepository} {
			u, err := url.Parse(raw)
			if err != nil || u.Scheme != "https" || u.Host == "" {
				t.Fatalf("non-HTTPS URL in corpus: %q", raw)
			}
		}
		if item.ArtifactURL != "" {
			assertDigestURL(t, item.ArtifactURL, item.ArtifactSHA256)
		}
		for _, artifact := range item.Artifacts {
			assertDigestURL(t, artifact.ArtifactURL, artifact.ArtifactSHA256)
		}
	}
}

func assertDigestURL(t *testing.T, rawURL, digest string) {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		t.Fatalf("non-HTTPS artifact URL: %q", rawURL)
	}
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(digest) {
		t.Fatalf("invalid SHA-256 digest %q", digest)
	}
	if len(hex.EncodeToString(make([]byte, sha256.Size))) != len(digest) {
		t.Fatalf("unexpected SHA-256 length for %q", digest)
	}
}

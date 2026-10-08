package npm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/releasecheck/releasecheck/internal/acquire"
	"github.com/releasecheck/releasecheck/internal/domain"
)

func TestVerifyMatchingNPMRelease(t *testing.T) {
	artifact := makeTarGzip(t, map[string][]byte{
		"package/index.js":     []byte("module.exports = 1;\n"),
		"package/package.json": []byte(`{"name":"demo","version":"1.0.0"}`),
	})
	source := makeTarGzip(t, map[string][]byte{
		"acme-demo-abc1234/index.js":     []byte("module.exports = 1;\n"),
		"acme-demo-abc1234/package.json": []byte(`{"name":"demo","version":"1.0.0"}`),
	})
	server := npmFixtureServer(t, artifact, source, false)
	defer server.Close()

	client := newFixtureClient(t, server)
	result, err := client.Verify(context.Background(), domain.ReleaseRequest{Registry: domain.RegistryNPM, Name: "demo", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	defer result.Cleanup()
	if result.Metadata.Identity.Version != "1.0.0" || result.Metadata.Source == nil || result.Metadata.Git == nil {
		t.Fatalf("metadata was not resolved: %+v", result.Metadata)
	}
	if len(result.Comparisons) != 2 {
		t.Fatalf("unexpected comparison count: %d", len(result.Comparisons))
	}
	for _, item := range result.Comparisons {
		if item.Kind != domain.ComparisonIdentical {
			t.Fatalf("unexpected comparison: %+v", item)
		}
	}
}

func TestResolveSelectsLatestAndHandlesMissingRepository(t *testing.T) {
	artifact := makeTarGzip(t, map[string][]byte{"package/index.js": []byte("safe")})
	server := npmFixtureServer(t, artifact, nil, true)
	defer server.Close()
	client := newFixtureClient(t, server)

	result, err := client.Verify(context.Background(), domain.ReleaseRequest{Registry: domain.RegistryNPM, Name: "missing"})
	if err != nil {
		t.Fatalf("verify without repository failed: %v", err)
	}
	defer result.Cleanup()
	if result.Metadata.Identity.Version != "1.0.0" || result.SourceArtifact != nil || len(result.Limitations) == 0 {
		t.Fatalf("missing repository was not reported honestly: %+v", result)
	}
}

func TestResolveRejectsMalformedMetadata(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, "{")
	}))
	defer server.Close()
	client := newFixtureClient(t, server)
	_, err := client.Resolve(context.Background(), domain.ReleaseRequest{Registry: domain.RegistryNPM, Name: "bad"})
	if err == nil || !strings.Contains(err.Error(), "metadata") {
		t.Fatalf("expected metadata error, got %v", err)
	}
}

func TestVerifyRejectsIntegrityMismatch(t *testing.T) {
	artifact := makeTarGzip(t, map[string][]byte{"package/index.js": []byte("safe")})
	server := npmFixtureServer(t, artifact, nil, false)
	defer server.Close()
	client := newFixtureClient(t, server)
	_, err := client.Verify(context.Background(), domain.ReleaseRequest{Registry: domain.RegistryNPM, Name: "mismatch", Version: "1.0.0"})
	if err == nil || !strings.Contains(err.Error(), "integrity") {
		t.Fatalf("expected integrity error, got %v", err)
	}
}

func TestVerifyReportsUnavailableSource(t *testing.T) {
	artifact := makeTarGzip(t, map[string][]byte{"package/index.js": []byte("safe")})
	server := npmFixtureServer(t, artifact, nil, false)
	defer server.Close()
	client := newFixtureClient(t, server)
	_, err := client.Verify(context.Background(), domain.ReleaseRequest{Registry: domain.RegistryNPM, Name: "missing-source", Version: "1.0.0"})
	if err == nil || !strings.Contains(err.Error(), "source") {
		t.Fatalf("expected source error, got %v", err)
	}
}

func TestRepositoryNormalizationAndScopedNames(t *testing.T) {
	source, reference, err := parseRepository(json.RawMessage(`{"type":"git","url":"git+https://github.com/acme/demo.git#v1.0.0","directory":"src"}`), "")
	if err != nil || source.URL != "https://github.com/acme/demo" || source.Directory != "src" || reference.Ref != "v1.0.0" || reference.Immutable {
		t.Fatalf("unexpected repository normalization: source=%+v reference=%+v err=%v", source, reference, err)
	}
	if got := escapePackageName("@scope/name"); got != "@scope%2Fname" {
		t.Fatalf("unexpected scoped package path: %q", got)
	}
}

func newFixtureClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	limits := acquire.Limits{MaxDownloadBytes: 1 << 20, MaxExpandedBytes: 1 << 20, MaxFileBytes: 1 << 20, MaxEntries: 100, MaxRedirects: 3, HTTPTimeout: time.Second}
	downloader, err := acquire.NewDownloader(acquire.DownloadOptions{HTTPClient: server.Client(), Limits: limits})
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(Options{BaseURL: server.URL, GitHubAPIBaseURL: server.URL, HTTPClient: server.Client(), Downloader: downloader})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func npmFixtureServer(t *testing.T, artifact, source []byte, missingRepository bool) *httptest.Server {
	t.Helper()
	artifactIntegrity := sri(artifact)
	wrongIntegrity := sri([]byte("wrong artifact"))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/demo" || r.URL.Path == "/missing" || r.URL.Path == "/mismatch" || r.URL.Path == "/missing-source":
			name := strings.TrimPrefix(r.URL.Path, "/")
			integrity := artifactIntegrity
			if name == "mismatch" {
				integrity = wrongIntegrity
			}
			repository := any(map[string]string{"type": "git", "url": "git+https://github.com/acme/demo.git"})
			gitHead := "abc1234"
			if missingRepository || name == "missing" {
				repository = nil
				gitHead = ""
			}
			if name == "missing-source" {
				gitHead = "deadbee"
			}
			response := map[string]any{
				"name":      name,
				"dist-tags": map[string]string{"latest": "1.0.0"},
				"versions": map[string]any{"1.0.0": map[string]any{
					"name":    name,
					"version": "1.0.0",
					"dist": map[string]any{
						"tarball":   serverURL(r) + "/artifact.tgz",
						"integrity": integrity,
					},
					"repository": repository,
					"gitHead":    gitHead,
				}},
			}
			writeJSON(t, w, response)
		case r.URL.Path == "/artifact.tgz":
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(artifact)
		case strings.HasPrefix(r.URL.Path, "/repos/acme/demo/tarball/") && source != nil:
			_, _ = w.Write(source)
		default:
			http.NotFound(w, r)
		}
	})
	server := httptest.NewTLSServer(handler)
	return server
}

func serverURL(r *http.Request) string {
	return "https://" + r.Host
}

func writeJSON(t *testing.T, writer http.ResponseWriter, value any) {
	t.Helper()
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		t.Fatal(err)
	}
}

func sri(content []byte) string {
	digest := sha512.Sum512(content)
	return "sha512-" + base64.StdEncoding.EncodeToString(digest[:])
}

func makeTarGzip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	for name, content := range files {
		if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: int64(len(content))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write(content); err != nil {
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

package pypi

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/releasecheck/releasecheck/internal/acquire"
	"github.com/releasecheck/releasecheck/internal/domain"
)

func TestVerifySelectsAndComparesPyPISdist(t *testing.T) {
	sdist := makeTARGzip(t, map[string][]byte{
		"demo-1.0.0/demo/__init__.py": []byte("VALUE = 1\n"),
		"demo-1.0.0/pyproject.toml":   []byte("[build-system]\n"),
	})
	wheel := makeWheel(t, map[string][]byte{"demo/__init__.py": []byte("VALUE = 1\n")})
	server := pypiFixtureServer(t, sdist, wheel, true, false)
	defer server.Close()
	client := newFixtureClient(t, server)

	result, err := client.Verify(context.Background(), domain.ReleaseRequest{Registry: domain.RegistryPyPI, Name: "demo"}, VerifyOptions{})
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	defer result.Cleanup()
	if result.Selected.ContentType != "sdist" || len(result.Metadata.Artifacts) != 2 || result.Metadata.Git == nil {
		t.Fatalf("unexpected selected release: %+v", result)
	}
	if len(result.Comparisons) != 2 {
		t.Fatalf("unexpected comparison count: %d", len(result.Comparisons))
	}
	for _, item := range result.Comparisons {
		if item.Kind != domain.ComparisonIdentical {
			t.Fatalf("unexpected sdist comparison: %+v", item)
		}
	}
}

func TestVerifyExplicitWheelReportsLimitation(t *testing.T) {
	sdist := makeTARGzip(t, map[string][]byte{"demo-1.0.0/demo/__init__.py": []byte("VALUE = 1\n")})
	wheel := makeWheel(t, map[string][]byte{
		"demo/__init__.py":              []byte("VALUE = 1\n"),
		"demo-1.0.0.dist-info/METADATA": []byte("Metadata-Version: 2.0\n"),
	})
	server := pypiFixtureServer(t, sdist, wheel, true, false)
	defer server.Close()
	client := newFixtureClient(t, server)

	result, err := client.Verify(context.Background(), domain.ReleaseRequest{Registry: domain.RegistryPyPI, Name: "demo", Version: "1.0.0"}, VerifyOptions{ArtifactFilename: "demo-1.0.0-py3-none-any.whl"})
	if err != nil {
		t.Fatalf("wheel verify failed: %v", err)
	}
	defer result.Cleanup()
	if result.Selected.ContentType != "bdist_wheel" || len(result.Limitations) == 0 {
		t.Fatalf("wheel limitation was not reported: %+v", result)
	}
	if len(result.Comparisons) == 0 {
		t.Fatal("wheel comparison produced no evidence")
	}
}

func TestResolveUsesExplicitVersionAndRetainsAllFiles(t *testing.T) {
	sdist := makeTARGzip(t, map[string][]byte{"demo-1.0.0/demo.py": []byte("safe")})
	wheel := makeWheel(t, map[string][]byte{"demo.py": []byte("safe")})
	server := pypiFixtureServer(t, sdist, wheel, true, false)
	defer server.Close()
	client := newFixtureClient(t, server)
	result, err := client.Resolve(context.Background(), domain.ReleaseRequest{Registry: domain.RegistryPyPI, Name: "demo", Version: "1.0.0"})
	if err != nil || len(result.Artifacts) != 2 {
		t.Fatalf("unexpected resolved release: %+v, %v", result, err)
	}
	if result.Source == nil || result.Git == nil || result.Git.Ref != "abc1234" {
		t.Fatalf("source reference was not resolved: source=%+v git=%+v", result.Source, result.Git)
	}
}

func TestVerifyRejectsHashMismatchAndMissingArtifact(t *testing.T) {
	sdist := makeTARGzip(t, map[string][]byte{"demo-1.0.0/demo.py": []byte("safe")})
	wheel := makeWheel(t, map[string][]byte{"demo.py": []byte("safe")})
	server := pypiFixtureServer(t, sdist, wheel, false, true)
	defer server.Close()
	client := newFixtureClient(t, server)
	_, err := client.Verify(context.Background(), domain.ReleaseRequest{Registry: domain.RegistryPyPI, Name: "demo", Version: "1.0.0"}, VerifyOptions{ArtifactFilename: "demo-1.0.0.tar.gz"})
	if err == nil || !strings.Contains(err.Error(), "integrity") {
		t.Fatalf("expected hash mismatch, got %v", err)
	}
	_, err = client.Verify(context.Background(), domain.ReleaseRequest{Registry: domain.RegistryPyPI, Name: "demo", Version: "1.0.0"}, VerifyOptions{ArtifactFilename: "missing.whl"})
	if err == nil || !strings.Contains(err.Error(), "not in the release") {
		t.Fatalf("expected missing artifact error, got %v", err)
	}
}

func TestVerifyReportsMissingSourceReference(t *testing.T) {
	sdist := makeTARGzip(t, map[string][]byte{"demo-1.0.0/demo.py": []byte("safe")})
	wheel := makeWheel(t, map[string][]byte{"demo.py": []byte("safe")})
	server := pypiFixtureServer(t, sdist, wheel, false, false)
	defer server.Close()
	client := newFixtureClient(t, server)
	result, err := client.Verify(context.Background(), domain.ReleaseRequest{Registry: domain.RegistryPyPI, Name: "demo", Version: "1.0.0"}, VerifyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer result.Cleanup()
	if len(result.Limitations) == 0 {
		t.Fatal("missing source reference was not reported")
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

func pypiFixtureServer(t *testing.T, sdist, wheel []byte, withSource, wrongHash bool) *httptest.Server {
	t.Helper()
	sdistHash := sha256Hex(sdist)
	if wrongHash {
		sdistHash = strings.Repeat("0", 64)
	}
	wheelHash := sha256Hex(wheel)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pypi/demo/json":
			writeJSON(t, w, map[string]any{"info": map[string]any{"name": "demo", "version": "1.0.0"}})
		case "/pypi/demo/1.0.0/json":
			projectURLs := map[string]string{}
			if withSource {
				projectURLs["Source"] = "https://github.com/acme/demo#abc1234"
			}
			writeJSON(t, w, map[string]any{
				"info": map[string]any{"name": "demo", "version": "1.0.0", "project_urls": projectURLs},
				"urls": []map[string]any{
					{"filename": "demo-1.0.0.tar.gz", "url": serverURL(r) + "/files/demo-1.0.0.tar.gz", "packagetype": "sdist", "digests": map[string]string{"sha256": sdistHash}, "size": len(sdist)},
					{"filename": "demo-1.0.0-py3-none-any.whl", "url": serverURL(r) + "/files/demo-1.0.0-py3-none-any.whl", "packagetype": "bdist_wheel", "digests": map[string]string{"sha256": wheelHash}, "size": len(wheel)},
				},
			})
		case "/files/demo-1.0.0.tar.gz":
			_, _ = w.Write(sdist)
		case "/files/demo-1.0.0-py3-none-any.whl":
			_, _ = w.Write(wheel)
		case "/repos/acme/demo/tarball/abc1234":
			if withSource {
				_, _ = w.Write(sdist)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		default:
			http.NotFound(w, r)
		}
	})
	return httptest.NewTLSServer(handler)
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

func sha256Hex(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func makeTARGzip(t *testing.T, files map[string][]byte) []byte {
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

func makeWheel(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

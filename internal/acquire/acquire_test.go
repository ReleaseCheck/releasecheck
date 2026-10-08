package acquire

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDownloadVerifiesHashAndCleansUp(t *testing.T) {
	body := []byte("safe archive bytes")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	digest := sha256.Sum256(body)
	d, err := NewDownloader(DownloadOptions{HTTPClient: server.Client(), AllowPrivateNetworks: true})
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := d.Download(context.Background(), server.URL, hex.EncodeToString(digest[:]))
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}
	if artifact.Size != int64(len(body)) || artifact.SHA256 != hex.EncodeToString(digest[:]) {
		t.Fatalf("unexpected artifact: %+v", artifact)
	}
	if _, err := os.Stat(artifact.Path); err != nil {
		t.Fatalf("downloaded file missing: %v", err)
	}
	if err := artifact.Cleanup(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if _, err := os.Stat(artifact.Path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("artifact remained after cleanup: %v", err)
	}
}

func TestDownloadFollowsHTTPSRedirect(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/artifact", http.StatusFound)
			return
		}
		_, _ = io.WriteString(w, "redirected")
	}))
	defer server.Close()

	d, err := NewDownloader(DownloadOptions{HTTPClient: server.Client(), AllowPrivateNetworks: true})
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := d.Download(context.Background(), server.URL+"/redirect", "")
	if err != nil {
		t.Fatalf("redirect download failed: %v", err)
	}
	defer artifact.Cleanup()
	if artifact.Size != int64(len("redirected")) {
		t.Fatalf("unexpected redirected size: %d", artifact.Size)
	}
}

func TestDownloadRejectsTooManyRedirects(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("n") != "3" {
			n := r.URL.Query().Get("n")
			if n == "" {
				n = "0"
			}
			if n == "0" {
				n = "1"
			} else if n == "1" {
				n = "2"
			} else {
				n = "3"
			}
			http.Redirect(w, r, "/redirect?n="+n, http.StatusFound)
			return
		}
		_, _ = io.WriteString(w, "unexpected")
	}))
	defer server.Close()

	d, err := NewDownloader(DownloadOptions{HTTPClient: server.Client(), AllowPrivateNetworks: true, Limits: Limits{MaxRedirects: 2}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.Download(context.Background(), server.URL+"/redirect", "")
	if err == nil || !strings.Contains(err.Error(), "redirect limit") {
		t.Fatalf("expected redirect limit error, got %v", err)
	}
}

func TestDownloadRejectsNetworkFailure(t *testing.T) {
	d, err := NewDownloader(DownloadOptions{AllowPrivateNetworks: true, Limits: Limits{HTTPTimeout: time.Second}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.Download(context.Background(), "https://127.0.0.1:1/unavailable", "")
	if err == nil || !strings.Contains(err.Error(), "network") {
		t.Fatalf("expected network error, got %v", err)
	}
}

func TestDownloadRejectsPrivateAddressByDefault(t *testing.T) {
	d, err := NewDownloader(DownloadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.Download(context.Background(), "https://127.0.0.1/resource", "")
	if err == nil || !strings.Contains(err.Error(), "private") {
		t.Fatalf("expected private-address rejection, got %v", err)
	}
}

func TestDownloadRejectsInvalidMetadataAndOversizedInput(t *testing.T) {
	invalidURLDownloader, err := NewDownloader(DownloadOptions{Limits: Limits{MaxDownloadBytes: 4}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := invalidURLDownloader.Download(context.Background(), "http://example.test/file", ""); err == nil {
		t.Fatal("HTTP URL accepted")
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "12345")
	}))
	defer server.Close()
	d, err := NewDownloader(DownloadOptions{HTTPClient: server.Client(), AllowPrivateNetworks: true, Limits: Limits{MaxDownloadBytes: 4}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.Download(context.Background(), server.URL, "")
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected oversized input error, got %v", err)
	}
}

func TestInspectZIPInventoryAndRejectsUnsafePath(t *testing.T) {
	good := makeZIP(t, map[string]string{"package/index.js": "console.log(1)", "package/README.md": "readme"})
	file := writeTemp(t, good)
	inventory, err := InspectArchive(file, Limits{MaxDownloadBytes: 1 << 20, MaxExpandedBytes: 1 << 20, MaxFileBytes: 1 << 20})
	if err != nil {
		t.Fatalf("inspect ZIP: %v", err)
	}
	if inventory.Kind != ArchiveZip || len(inventory.Entries) != 2 || inventory.ExpandedBytes == 0 || inventory.Entries[0].SHA256 == "" {
		t.Fatalf("unexpected ZIP inventory: %+v", inventory)
	}

	unsafe := makeZIP(t, map[string]string{"../escape": "no"})
	unsafeFile := writeTemp(t, unsafe)
	if _, err := InspectArchive(unsafeFile, DefaultLimits()); err == nil || !strings.Contains(err.Error(), "unsafe archive path") {
		t.Fatalf("unsafe path was accepted: %v", err)
	}

	duplicate := makeDuplicateZIP(t)
	duplicateFile := writeTemp(t, duplicate)
	if _, err := InspectArchive(duplicateFile, DefaultLimits()); err == nil || !strings.Contains(err.Error(), "duplicate archive path") {
		t.Fatalf("duplicate path was accepted: %v", err)
	}
}

func TestInspectTARGzipInventoryAndBounds(t *testing.T) {
	archive := makeTARGzip(t, "src/main.py", []byte("print('safe')"))
	file := writeTemp(t, archive)
	inventory, err := InspectArchive(file, DefaultLimits())
	if err != nil {
		t.Fatalf("inspect TAR.GZ: %v", err)
	}
	if inventory.Kind != ArchiveTarGzip || len(inventory.Entries) != 1 || inventory.Entries[0].SHA256 == "" {
		t.Fatalf("unexpected TAR.GZ inventory: %+v", inventory)
	}

	small := Limits{MaxDownloadBytes: 1 << 20, MaxExpandedBytes: 4, MaxFileBytes: 4, MaxEntries: 2, MaxRedirects: 1, HTTPTimeout: time.Second}
	if _, err := InspectArchive(file, small); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expanded-size bound was not enforced: %v", err)
	}
}

func TestInspectArchiveRejectsMalformedAndUnsupportedInput(t *testing.T) {
	malformed := writeTemp(t, []byte("not an archive"))
	if _, err := InspectArchive(malformed, DefaultLimits()); err == nil {
		t.Fatal("malformed archive accepted")
	}
	unsupported := writeTemp(t, []byte("plain text"))
	if _, err := InspectArchive(unsupported, DefaultLimits()); err == nil {
		t.Fatal("unsupported input accepted")
	}
}

func makeZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(entry, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func makeDuplicateZIP(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, content := range []string{"first", "second"} {
		entry, err := writer.Create("package/duplicate.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(entry, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func makeTARGzip(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func writeTemp(t *testing.T, content []byte) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "archive.bin")
	if err := os.WriteFile(file, content, 0600); err != nil {
		t.Fatal(err)
	}
	return file
}

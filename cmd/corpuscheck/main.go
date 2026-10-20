// Command corpuscheck downloads and verifies the pinned corpus metadata.
// It never installs, imports, extracts, or executes the downloaded artifacts.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxArtifactBytes int64 = 100 << 20

type corpus struct {
	Records []record `json:"records"`
}
type record struct {
	Ecosystem      string     `json:"ecosystem"`
	Package        string     `json:"package"`
	ArtifactURL    string     `json:"artifact_url"`
	ArtifactName   string     `json:"artifact_filename"`
	ArtifactSize   int64      `json:"artifact_size"`
	ArtifactSHA256 string     `json:"artifact_sha256"`
	Artifacts      []artifact `json:"artifacts"`
}
type artifact struct {
	ArtifactURL    string `json:"artifact_url"`
	ArtifactName   string `json:"artifact_filename"`
	ArtifactSize   int64  `json:"artifact_size"`
	ArtifactSHA256 string `json:"artifact_sha256"`
}

func main() {
	corpusPath := flag.String("corpus", "fixtures/corpus/real-releases.json", "corpus metadata JSON path")
	outputDir := flag.String("output-dir", "", "directory for verified artifact bytes; empty uses a temporary directory")
	timeout := flag.Duration("timeout", 30*time.Second, "per-request timeout")
	flag.Parse()
	if *timeout <= 0 {
		fail(errors.New("timeout must be positive"))
	}
	data, err := os.ReadFile(*corpusPath)
	if err != nil {
		fail(fmt.Errorf("read corpus: %w", err))
	}
	var entries corpus
	if err := json.Unmarshal(data, &entries); err != nil {
		fail(fmt.Errorf("parse corpus: %w", err))
	}

	directory := *outputDir
	removeDirectory := false
	if directory == "" {
		directory, err = os.MkdirTemp("", "releasecheck-corpus-")
		if err != nil {
			fail(fmt.Errorf("create temporary output directory: %w", err))
		}
		removeDirectory = true
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		fail(fmt.Errorf("create output directory: %w", err))
	}
	if removeDirectory {
		defer os.RemoveAll(directory)
	}

	client := &http.Client{Timeout: *timeout, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "https" || req.URL.User != nil {
			return errors.New("redirect must remain HTTPS and cannot include userinfo")
		}
		if len(via) >= 5 {
			return errors.New("redirect limit exceeded")
		}
		return nil
	}}
	count := 0
	for _, entry := range entries.Records {
		items := []artifact{{ArtifactURL: entry.ArtifactURL, ArtifactName: entry.ArtifactName, ArtifactSize: entry.ArtifactSize, ArtifactSHA256: entry.ArtifactSHA256}}
		if entry.ArtifactURL == "" {
			items = entry.Artifacts
		}
		for _, item := range items {
			if item.ArtifactURL == "" {
				continue
			}
			if err := verifyArtifact(client, directory, entry, item); err != nil {
				fail(err)
			}
			count++
			fmt.Printf("verified %s %s %s\n", entry.Ecosystem, entry.Package, item.ArtifactName)
		}
	}
	fmt.Printf("verified %d artifact(s); no package code was executed\n", count)
}

func verifyArtifact(client *http.Client, directory string, entry record, item artifact) error {
	u, err := url.Parse(item.ArtifactURL)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return fmt.Errorf("%s %s: artifact URL must be HTTPS without userinfo", entry.Package, item.ArtifactURL)
	}
	name := filepath.Base(item.ArtifactName)
	if name == "." || name == string(filepath.Separator) || name != item.ArtifactName || strings.ContainsAny(item.ArtifactName, `/\\`) {
		return fmt.Errorf("%s %s: artifact filename must be a plain filename", entry.Package, item.ArtifactName)
	}
	if len(item.ArtifactSHA256) != sha256.Size*2 {
		return fmt.Errorf("%s %s: invalid SHA-256 length", entry.Package, item.ArtifactName)
	}
	if _, err := hex.DecodeString(item.ArtifactSHA256); err != nil {
		return fmt.Errorf("%s %s: invalid SHA-256 digest: %w", entry.Package, item.ArtifactName, err)
	}

	response, err := client.Get(item.ArtifactURL)
	if err != nil {
		return fmt.Errorf("download %s: %w", item.ArtifactURL, err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download %s: HTTP %s", item.ArtifactURL, response.Status)
	}
	if response.ContentLength > maxArtifactBytes {
		return fmt.Errorf("download %s: content length exceeds %d bytes", item.ArtifactURL, maxArtifactBytes)
	}

	temporary, err := os.CreateTemp(directory, ".download-*")
	if err != nil {
		return fmt.Errorf("create temporary artifact: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	hash := sha256.New()
	bytes, err := io.Copy(io.MultiWriter(temporary, hash), io.LimitReader(response.Body, maxArtifactBytes+1))
	closeErr := temporary.Close()
	if err != nil {
		return fmt.Errorf("write %s: %w", item.ArtifactName, err)
	}
	if closeErr != nil {
		return fmt.Errorf("close %s: %w", item.ArtifactName, closeErr)
	}
	if bytes > maxArtifactBytes {
		return fmt.Errorf("download %s: exceeds %d bytes", item.ArtifactURL, maxArtifactBytes)
	}
	if item.ArtifactSize > 0 && bytes != item.ArtifactSize {
		return fmt.Errorf("%s: size mismatch: got %d, expected %d", item.ArtifactName, bytes, item.ArtifactSize)
	}
	if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), item.ArtifactSHA256) {
		return fmt.Errorf("%s: SHA-256 mismatch", item.ArtifactName)
	}
	if err := os.Rename(temporaryName, filepath.Join(directory, name)); err != nil {
		return fmt.Errorf("store verified %s: %w", item.ArtifactName, err)
	}
	return nil
}

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }

// Package npm implements the npm registry verification path.
package npm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/releasecheck/releasecheck/internal/acquire"
	"github.com/releasecheck/releasecheck/internal/compare"
	"github.com/releasecheck/releasecheck/internal/domain"
	"github.com/releasecheck/releasecheck/internal/security"
)

const (
	defaultRegistryURL = "https://registry.npmjs.org"
	defaultGitHubAPI   = "https://api.github.com"
	metadataLimit      = 16 << 20
)

var commitPattern = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

// Options configures an npm client. Base URLs remain HTTPS-only; test clients
// may use an HTTPS httptest server with its certificate supplied in HTTPClient.
type Options struct {
	BaseURL          string
	GitHubAPIBaseURL string
	HTTPClient       *http.Client
	Downloader       *acquire.Downloader
	Limits           acquire.Limits
}

// Client resolves npm metadata and performs the non-executing verification
// path for one selected package version.
type Client struct {
	baseURL      string
	githubAPIURL string
	httpClient   *http.Client
	downloader   *acquire.Downloader
}

// NewClient creates an npm registry client.
func NewClient(options Options) (*Client, error) {
	baseURL := strings.TrimRight(options.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultRegistryURL
	}
	if err := acquire.ValidateHTTPSURL(baseURL); err != nil {
		return nil, domain.NewError(domain.ErrorInput, "validate npm registry URL", err)
	}
	githubAPIURL := strings.TrimRight(options.GitHubAPIBaseURL, "/")
	if githubAPIURL == "" {
		githubAPIURL = defaultGitHubAPI
	}
	if err := acquire.ValidateHTTPSURL(githubAPIURL); err != nil {
		return nil, domain.NewError(domain.ErrorInput, "validate GitHub API URL", err)
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		copy := *http.DefaultClient
		httpClient = &copy
	} else {
		copy := *httpClient
		httpClient = &copy
	}
	if httpClient.Timeout == 0 {
		httpClient.Timeout = 30 * time.Second
	}
	httpClient.CheckRedirect = func(req *http.Request, previous []*http.Request) error {
		if len(previous) >= 5 {
			return errors.New("npm metadata redirect limit exceeded")
		}
		if err := acquire.ValidateHTTPSURL(req.URL.String()); err != nil {
			return fmt.Errorf("npm metadata redirect target: %w", err)
		}
		return nil
	}
	downloader := options.Downloader
	if downloader == nil {
		var err error
		downloader, err = acquire.NewDownloader(acquire.DownloadOptions{
			HTTPClient: httpClient,
			Limits:     options.Limits,
		})
		if err != nil {
			return nil, err
		}
	}
	return &Client{baseURL: baseURL, githubAPIURL: githubAPIURL, httpClient: httpClient, downloader: downloader}, nil
}

type packument struct {
	Name     string                     `json:"name"`
	DistTags map[string]string          `json:"dist-tags"`
	Versions map[string]versionMetadata `json:"versions"`
}

type versionMetadata struct {
	Name       string          `json:"name"`
	Version    string          `json:"version"`
	Dist       distribution    `json:"dist"`
	Repository json.RawMessage `json:"repository"`
	GitHead    string          `json:"gitHead"`
}

type distribution struct {
	Tarball      string `json:"tarball"`
	Integrity    string `json:"integrity"`
	Shasum       string `json:"shasum"`
	FileCount    int    `json:"fileCount"`
	UnpackedSize int64  `json:"unpackedSize"`
}

// Resolve fetches and normalizes npm packument metadata without downloading or
// executing the package artifact.
func (c *Client) Resolve(ctx context.Context, request domain.ReleaseRequest) (domain.ReleaseMetadata, error) {
	if request.Registry != domain.RegistryNPM {
		return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorInput, "resolve npm release", errors.New("request registry must be npm"))
	}
	if strings.TrimSpace(request.Name) == "" {
		return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorInput, "resolve npm release", errors.New("package name is required"))
	}
	metadataURL := c.baseURL + "/" + escapePackageName(request.Name)
	var document packument
	if err := c.getJSON(ctx, metadataURL, &document); err != nil {
		return domain.ReleaseMetadata{}, err
	}
	version := strings.TrimSpace(request.Version)
	if version == "" {
		version = document.DistTags["latest"]
	}
	selected, ok := document.Versions[version]
	if !ok {
		return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorMetadata, "select npm version", fmt.Errorf("version %q is not present in registry metadata", version))
	}
	if selected.Dist.Tarball == "" {
		return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorMetadata, "read npm distribution", errors.New("dist.tarball is missing"))
	}
	identity := domain.PackageIdentity{Registry: domain.RegistryNPM, Name: request.Name, Version: version}
	artifact := domain.Artifact{
		Identity:      identity,
		Filename:      filenameFromURL(selected.Dist.Tarball),
		URL:           selected.Dist.Tarball,
		ContentType:   "application/gzip",
		ExpectedFiles: selected.Dist.FileCount,
	}
	if err := artifact.Validate(); err != nil {
		return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorMetadata, "validate npm artifact", err)
	}
	result := domain.ReleaseMetadata{
		Identity:  identity,
		Artifacts: []domain.Artifact{artifact},
		Evidence: []domain.Evidence{
			{ID: "npm.version.selected", State: domain.EvidenceKnown, Subject: "package.version", Value: version},
			{ID: "npm.artifact.url", State: domain.EvidenceKnown, Subject: "artifact.url", Value: selected.Dist.Tarball},
		},
	}
	if selected.Dist.Integrity != "" {
		result.Evidence = append(result.Evidence, domain.Evidence{ID: "npm.integrity.claimed", State: domain.EvidenceKnown, Subject: "artifact.integrity", Value: selected.Dist.Integrity})
	} else if selected.Dist.Shasum != "" {
		result.Evidence = append(result.Evidence, domain.Evidence{ID: "npm.integrity.claimed", State: domain.EvidenceKnown, Subject: "artifact.shasum", Value: selected.Dist.Shasum, Description: "Legacy SHA-1 metadata is recorded but is not treated as a strong digest."})
	} else {
		result.Evidence = append(result.Evidence, domain.Evidence{ID: "npm.integrity.unavailable", State: domain.EvidenceUnavailable, Subject: "artifact.integrity"})
	}

	source, repositoryRef, sourceErr := parseRepository(selected.Repository, selected.GitHead)
	if sourceErr != nil {
		result.Evidence = append(result.Evidence, domain.Evidence{ID: "npm.repository.invalid", State: domain.EvidenceInvalid, Subject: "source.repository", Description: sourceErr.Error()})
		return result, nil
	}
	if source == nil {
		result.Evidence = append(result.Evidence, domain.Evidence{ID: "npm.repository.unavailable", State: domain.EvidenceUnavailable, Subject: "source.repository"})
		return result, nil
	}
	result.Source = source
	result.Evidence = append(result.Evidence, domain.Evidence{ID: "npm.repository.claimed", State: domain.EvidenceKnown, Subject: "source.repository", Value: source.URL})
	if repositoryRef != nil {
		result.Git = repositoryRef
		result.Evidence = append(result.Evidence, domain.Evidence{ID: "npm.git.reference.claimed", State: domain.EvidenceKnown, Subject: "source.git_reference", Value: repositoryRef.Ref})
	} else {
		result.Evidence = append(result.Evidence, domain.Evidence{ID: "npm.git.reference.unavailable", State: domain.EvidenceUnavailable, Subject: "source.git_reference"})
	}
	return result, nil
}

// VerificationResult contains deterministic npm acquisition and comparison
// output. It is not a final user-facing report or a trust verdict.
type VerificationResult struct {
	Metadata          domain.ReleaseMetadata
	Artifact          *acquire.DownloadedArtifact
	ArtifactInventory acquire.ArchiveInventory
	SourceArtifact    *acquire.DownloadedArtifact
	SourceInventory   acquire.ArchiveInventory
	Comparisons       []domain.FileComparison
	Limitations       []string
	Security          []domain.SecurityObservation
}

// Verify resolves metadata, downloads and inventories the npm artifact, then
// retrieves a GitHub source snapshot when the claimed repository/ref is usable.
func (c *Client) Verify(ctx context.Context, request domain.ReleaseRequest) (VerificationResult, error) {
	metadata, err := c.Resolve(ctx, request)
	if err != nil {
		return VerificationResult{}, err
	}
	result := VerificationResult{Metadata: metadata}
	if len(metadata.Artifacts) != 1 {
		return result, domain.NewError(domain.ErrorMetadata, "select npm artifact", errors.New("npm release did not resolve to exactly one artifact"))
	}
	artifact := metadata.Artifacts[0]
	downloaded, err := c.downloader.Download(ctx, artifact.URL, "")
	if err != nil {
		return result, err
	}
	result.Artifact = &downloaded
	result.Metadata.Artifacts[0].SHA256 = downloaded.SHA256
	if err := verifyNPMIntegrity(downloaded.Path, metadata.Evidence); err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, err
	}
	limits := acquire.DefaultLimits()
	artifactInventory, err := acquire.InspectArchive(downloaded.Path, limits)
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, err
	}
	result.ArtifactInventory = artifactInventory
	if manifest, readErr := acquire.ReadArchiveFile(downloaded.Path, "package.json", limits); readErr == nil {
		result.Security = append(result.Security, security.AnalyzeNPMManifest(manifest)...)
	} else {
		result.Limitations = append(result.Limitations, "package.json was not available for manifest security analysis")
	}
	if metadata.Source == nil || metadata.Git == nil {
		result.Limitations = append(result.Limitations, "source comparison is unavailable because repository or git reference metadata is missing")
		return result, nil
	}
	owner, repository, ok := githubRepository(metadata.Source.URL)
	if !ok {
		result.Limitations = append(result.Limitations, "source comparison is unavailable because the repository host is not a supported GitHub source")
		return result, nil
	}
	sourceURL := c.githubAPIURL + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repository) + "/tarball/" + url.PathEscape(metadata.Git.Ref)
	sourceArchive, err := c.downloader.Download(ctx, sourceURL, "")
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorSource, "download source snapshot", err)
	}
	result.SourceArtifact = &sourceArchive
	sourceInventory, err := acquire.InspectArchive(sourceArchive.Path, limits)
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorSource, "inspect source snapshot", err)
	}
	result.SourceInventory = sourceInventory
	artifactEntries, err := compare.Inventory(artifactInventory, "package")
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorComparison, "normalize npm artifact inventory", err)
	}
	sourceRoot, err := archiveRoot(sourceInventory)
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorComparison, "identify source archive root", err)
	}
	sourceEntries, err := compare.Inventory(sourceInventory, sourceRoot)
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorComparison, "normalize source inventory", err)
	}
	result.Comparisons, err = compare.Compare(sourceEntries, artifactEntries)
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorComparison, "compare npm inventories", err)
	}
	return result, nil
}

// Cleanup removes any downloaded artifact and source snapshot.
func (r *VerificationResult) Cleanup() error {
	if r == nil {
		return nil
	}
	var first error
	if r.Artifact != nil {
		if err := r.Artifact.Cleanup(); err != nil && first == nil {
			first = err
		}
	}
	if r.SourceArtifact != nil {
		if err := r.SourceArtifact.Cleanup(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (c *Client) getJSON(ctx context.Context, rawURL string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return domain.NewError(domain.ErrorInput, "create npm metadata request", err)
	}
	request.Header.Set("Accept", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return domain.NewError(domain.ErrorNetwork, "download npm metadata", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return domain.NewError(domain.ErrorMetadata, "read npm metadata", fmt.Errorf("unexpected HTTP status %s", response.Status))
	}
	limited := io.LimitReader(response.Body, metadataLimit+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return domain.NewError(domain.ErrorNetwork, "read npm metadata", err)
	}
	if len(data) > metadataLimit {
		return domain.NewError(domain.ErrorMetadata, "bound npm metadata", errors.New("metadata exceeds size limit"))
	}
	if err := json.Unmarshal(data, target); err != nil {
		return domain.NewError(domain.ErrorMetadata, "decode npm metadata", err)
	}
	return nil
}

func escapePackageName(name string) string {
	return strings.ReplaceAll(url.PathEscape(name), "/", "%2F")
}

func filenameFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || path.Base(parsed.Path) == "." || path.Base(parsed.Path) == "/" {
		return "package.tgz"
	}
	return path.Base(parsed.Path)
}

func parseRepository(raw json.RawMessage, gitHead string) (*domain.SourceReference, *domain.GitReference, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil, nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		var object struct {
			URL       string `json:"url"`
			Directory string `json:"directory"`
		}
		if err := json.Unmarshal(raw, &object); err != nil || object.URL == "" {
			return nil, nil, errors.New("repository metadata is neither a string nor a usable object")
		}
		value = object.URL
		return normalizedSource(value, object.Directory, gitHead)
	}
	return normalizedSource(value, "", gitHead)
}

func normalizedSource(rawURL, directory, gitHead string) (*domain.SourceReference, *domain.GitReference, error) {
	value := strings.TrimSpace(rawURL)
	value = strings.TrimPrefix(value, "git+")
	if strings.HasPrefix(value, "git://") {
		value = "https://" + strings.TrimPrefix(value, "git://")
	}
	if strings.HasPrefix(value, "git@github.com:") {
		value = "https://github.com/" + strings.TrimPrefix(value, "git@github.com:")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return nil, nil, errors.New("repository URL is malformed")
	}
	claimedRef := parsed.Fragment
	parsed.Fragment = ""
	parsed.RawFragment = ""
	parsed.Path = strings.TrimSuffix(parsed.Path, ".git")
	if parsed.Scheme != "https" {
		return nil, nil, errors.New("repository URL is not HTTPS")
	}
	source := &domain.SourceReference{URL: parsed.String(), Directory: directory}
	if err := source.Validate(); err != nil {
		return nil, nil, err
	}
	if gitHead != "" {
		claimedRef = gitHead
	}
	if claimedRef == "" {
		return source, nil, nil
	}
	if strings.ContainsAny(claimedRef, "\x00\r\n") {
		return source, nil, errors.New("git reference contains invalid characters")
	}
	reference := &domain.GitReference{Ref: claimedRef, Source: source.URL, Immutable: commitPattern.MatchString(claimedRef)}
	if reference.Immutable {
		reference.Commit = claimedRef
	}
	return source, reference, nil
}

func githubRepository(rawURL string) (string, string, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil || !strings.EqualFold(parsed.Host, "github.com") {
		return "", "", false
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], strings.TrimSuffix(parts[1], ".git"), true
}

func archiveRoot(inventory acquire.ArchiveInventory) (string, error) {
	for _, entry := range inventory.Entries {
		parts := strings.SplitN(entry.Path, "/", 2)
		if parts[0] == "" {
			continue
		}
		return parts[0], nil
	}
	return "", errors.New("source archive contains no entries")
}

func verifyNPMIntegrity(filename string, evidence []domain.Evidence) error {
	for _, item := range evidence {
		if item.ID == "npm.integrity.claimed" && strings.HasPrefix(item.Value, "sha") {
			return acquire.VerifyIntegrity(filename, item.Value)
		}
	}
	return nil
}

// Package pypi implements the PyPI release verification path.
package pypi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/releasecheck/releasecheck/internal/acquire"
	"github.com/releasecheck/releasecheck/internal/compare"
	"github.com/releasecheck/releasecheck/internal/domain"
)

const (
	defaultIndexURL  = "https://pypi.org"
	defaultGitHubAPI = "https://api.github.com"
	metadataLimit    = 16 << 20
)

// Options configures a PyPI client. Base URLs remain HTTPS-only; tests may use
// HTTPS httptest servers with their certificate supplied in HTTPClient.
type Options struct {
	BaseURL          string
	GitHubAPIBaseURL string
	HTTPClient       *http.Client
	Downloader       *acquire.Downloader
	Limits           acquire.Limits
}

// VerifyOptions selects one file from a PyPI release. An empty filename
// selects the sdist when available, then the lexically first release file.
type VerifyOptions struct {
	ArtifactFilename string
}

// Client resolves PyPI release metadata and performs non-executing checks.
type Client struct {
	baseURL      string
	githubAPIURL string
	httpClient   *http.Client
	downloader   *acquire.Downloader
}

type projectResponse struct {
	Info releaseInfo `json:"info"`
}

type releaseResponse struct {
	Info releaseInfo   `json:"info"`
	URLs []releaseFile `json:"urls"`
}

type releaseInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	HomePage    string            `json:"home_page"`
	ProjectURLs map[string]string `json:"project_urls"`
}

type releaseFile struct {
	Filename    string            `json:"filename"`
	URL         string            `json:"url"`
	Packagetype string            `json:"packagetype"`
	Digests     map[string]string `json:"digests"`
	Size        int64             `json:"size"`
	Yanked      bool              `json:"yanked"`
}

// NewClient creates a PyPI client.
func NewClient(options Options) (*Client, error) {
	baseURL := strings.TrimRight(options.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultIndexURL
	}
	if err := acquire.ValidateHTTPSURL(baseURL); err != nil {
		return nil, domain.NewError(domain.ErrorInput, "validate PyPI index URL", err)
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
			return errors.New("PyPI metadata redirect limit exceeded")
		}
		if err := acquire.ValidateHTTPSURL(req.URL.String()); err != nil {
			return fmt.Errorf("PyPI metadata redirect target: %w", err)
		}
		return nil
	}
	downloader := options.Downloader
	if downloader == nil {
		var err error
		downloader, err = acquire.NewDownloader(acquire.DownloadOptions{HTTPClient: httpClient, Limits: options.Limits})
		if err != nil {
			return nil, err
		}
	}
	return &Client{baseURL: baseURL, githubAPIURL: githubAPIURL, httpClient: httpClient, downloader: downloader}, nil
}

// Resolve reads PyPI's project/release JSON APIs and returns every file in the
// selected release, plus a claimed source URL when one is discoverable.
func (c *Client) Resolve(ctx context.Context, request domain.ReleaseRequest) (domain.ReleaseMetadata, error) {
	if request.Registry != domain.RegistryPyPI {
		return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorInput, "resolve PyPI release", errors.New("request registry must be pypi"))
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorInput, "resolve PyPI release", errors.New("project name is required"))
	}
	version := strings.TrimSpace(request.Version)
	if version == "" {
		var project projectResponse
		if err := c.getJSON(ctx, c.baseURL+"/pypi/"+url.PathEscape(name)+"/json", &project); err != nil {
			return domain.ReleaseMetadata{}, err
		}
		version = project.Info.Version
	}
	if version == "" {
		return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorMetadata, "select PyPI version", errors.New("release version is missing"))
	}
	var release releaseResponse
	releaseURL := c.baseURL + "/pypi/" + url.PathEscape(name) + "/" + url.PathEscape(version) + "/json"
	if err := c.getJSON(ctx, releaseURL, &release); err != nil {
		return domain.ReleaseMetadata{}, err
	}
	identity := domain.PackageIdentity{Registry: domain.RegistryPyPI, Name: name, Version: version}
	result := domain.ReleaseMetadata{Identity: identity, Evidence: []domain.Evidence{
		{ID: "pypi.version.selected", State: domain.EvidenceKnown, Subject: "package.version", Value: version},
	}}
	if len(release.URLs) == 0 {
		return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorMetadata, "read PyPI release files", errors.New("release contains no files"))
	}
	for _, file := range release.URLs {
		if file.Filename == "" || file.URL == "" {
			return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorMetadata, "read PyPI release file", errors.New("release file is missing filename or URL"))
		}
		digest := file.Digests["sha256"]
		if len(digest) != 64 {
			return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorMetadata, "read PyPI release hash", fmt.Errorf("file %q has no valid SHA-256 digest", file.Filename))
		}
		artifact := domain.Artifact{
			Identity:       identity,
			Filename:       file.Filename,
			URL:            file.URL,
			ExpectedSHA256: strings.ToLower(digest),
			Size:           file.Size,
			ContentType:    file.Packagetype,
		}
		if err := artifact.Validate(); err != nil {
			return domain.ReleaseMetadata{}, domain.NewError(domain.ErrorMetadata, "validate PyPI release file", err)
		}
		result.Artifacts = append(result.Artifacts, artifact)
		state := domain.EvidenceKnown
		description := "SHA-256 digest supplied by PyPI release metadata."
		if file.Yanked {
			description += " The file is marked yanked."
		}
		result.Evidence = append(result.Evidence, domain.Evidence{ID: "pypi.artifact.hash." + file.Filename, State: state, Subject: "artifact.sha256", Value: digest, Description: description})
	}
	source, sourceRef, err := sourceFromInfo(release.Info)
	if err != nil {
		result.Evidence = append(result.Evidence, domain.Evidence{ID: "pypi.source.invalid", State: domain.EvidenceInvalid, Subject: "source.repository", Description: err.Error()})
	} else if source != nil {
		result.Source = source
		result.Evidence = append(result.Evidence, domain.Evidence{ID: "pypi.source.claimed", State: domain.EvidenceKnown, Subject: "source.repository", Value: source.URL})
		if sourceRef != nil {
			result.Git = sourceRef
			result.Evidence = append(result.Evidence, domain.Evidence{ID: "pypi.git.reference.claimed", State: domain.EvidenceKnown, Subject: "source.git_reference", Value: sourceRef.Ref})
		} else {
			result.Evidence = append(result.Evidence, domain.Evidence{ID: "pypi.git.reference.unavailable", State: domain.EvidenceUnavailable, Subject: "source.git_reference"})
		}
	} else {
		result.Evidence = append(result.Evidence, domain.Evidence{ID: "pypi.source.unavailable", State: domain.EvidenceUnavailable, Subject: "source.repository"})
	}
	sort.Slice(result.Artifacts, func(i, j int) bool { return result.Artifacts[i].Filename < result.Artifacts[j].Filename })
	return result, nil
}

// VerificationResult contains deterministic PyPI acquisition and comparison
// output. It is not a final user-facing report or a trust verdict.
type VerificationResult struct {
	Metadata          domain.ReleaseMetadata
	Selected          domain.Artifact
	Artifact          *acquire.DownloadedArtifact
	ArtifactInventory acquire.ArchiveInventory
	SourceArtifact    *acquire.DownloadedArtifact
	SourceInventory   acquire.ArchiveInventory
	Comparisons       []domain.FileComparison
	Limitations       []string
}

// Verify selects one release file, verifies its PyPI SHA-256, and compares it
// with a source snapshot when the source claim is a supported GitHub URL.
func (c *Client) Verify(ctx context.Context, request domain.ReleaseRequest, options VerifyOptions) (VerificationResult, error) {
	metadata, err := c.Resolve(ctx, request)
	if err != nil {
		return VerificationResult{}, err
	}
	selected, err := selectArtifact(metadata.Artifacts, options.ArtifactFilename)
	if err != nil {
		return VerificationResult{Metadata: metadata}, err
	}
	result := VerificationResult{Metadata: metadata, Selected: selected}
	downloaded, err := c.downloader.Download(ctx, selected.URL, selected.ExpectedSHA256)
	if err != nil {
		return result, err
	}
	result.Artifact = &downloaded
	for index := range result.Metadata.Artifacts {
		if result.Metadata.Artifacts[index].Filename == selected.Filename {
			result.Metadata.Artifacts[index].SHA256 = downloaded.SHA256
		}
	}
	if selected.Size > 0 && downloaded.Size != selected.Size {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorIntegrity, "verify PyPI artifact size", fmt.Errorf("got %d bytes, expected %d", downloaded.Size, selected.Size))
	}
	result.ArtifactInventory, err = acquire.InspectArchive(downloaded.Path, acquire.DefaultLimits())
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, err
	}
	if result.Metadata.Source == nil || result.Metadata.Git == nil {
		result.Limitations = append(result.Limitations, "source comparison is unavailable because PyPI project metadata has no usable source URL and git reference")
		return result, nil
	}
	owner, repository, ok := githubRepository(result.Metadata.Source.URL)
	if !ok {
		result.Limitations = append(result.Limitations, "source comparison is unavailable because the source host is not a supported GitHub repository")
		return result, nil
	}
	sourceURL := c.githubAPIURL + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repository) + "/tarball/" + url.PathEscape(result.Metadata.Git.Ref)
	sourceArchive, err := c.downloader.Download(ctx, sourceURL, "")
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorSource, "download PyPI source snapshot", err)
	}
	result.SourceArtifact = &sourceArchive
	result.SourceInventory, err = acquire.InspectArchive(sourceArchive.Path, acquire.DefaultLimits())
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorSource, "inspect PyPI source snapshot", err)
	}
	sourceRoot, err := archiveRoot(result.SourceInventory)
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorComparison, "identify PyPI source archive root", err)
	}
	sourceEntries, err := compare.Inventory(result.SourceInventory, sourceRoot)
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorComparison, "normalize PyPI source inventory", err)
	}
	artifactRoot := ""
	if selected.ContentType == "sdist" {
		artifactRoot, err = archiveRoot(result.ArtifactInventory)
		if err != nil {
			_ = result.Cleanup()
			return VerificationResult{}, domain.NewError(domain.ErrorComparison, "identify PyPI sdist archive root", err)
		}
	}
	artifactEntries, err := compare.Inventory(result.ArtifactInventory, artifactRoot)
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorComparison, "normalize PyPI artifact inventory", err)
	}
	result.Comparisons, err = compare.Compare(sourceEntries, artifactEntries)
	if err != nil {
		_ = result.Cleanup()
		return VerificationResult{}, domain.NewError(domain.ErrorComparison, "compare PyPI inventories", err)
	}
	if selected.ContentType == "bdist_wheel" {
		result.Limitations = append(result.Limitations, "wheel contents may include generated metadata or compiled distribution files and are not a byte-for-byte source equivalence claim")
	}
	return result, nil
}

// Cleanup removes downloaded release and source files.
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

func selectArtifact(artifacts []domain.Artifact, filename string) (domain.Artifact, error) {
	if filename != "" {
		for _, artifact := range artifacts {
			if artifact.Filename == filename {
				return artifact, nil
			}
		}
		return domain.Artifact{}, domain.NewError(domain.ErrorInput, "select PyPI artifact", fmt.Errorf("file %q is not in the release", filename))
	}
	for _, artifact := range artifacts {
		if artifact.ContentType == "sdist" {
			return artifact, nil
		}
	}
	if len(artifacts) == 0 {
		return domain.Artifact{}, domain.NewError(domain.ErrorMetadata, "select PyPI artifact", errors.New("release has no artifacts"))
	}
	return artifacts[0], nil
}

func (c *Client) getJSON(ctx context.Context, rawURL string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return domain.NewError(domain.ErrorInput, "create PyPI metadata request", err)
	}
	request.Header.Set("Accept", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return domain.NewError(domain.ErrorNetwork, "download PyPI metadata", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return domain.NewError(domain.ErrorMetadata, "read PyPI metadata", fmt.Errorf("unexpected HTTP status %s", response.Status))
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, metadataLimit+1))
	if err != nil {
		return domain.NewError(domain.ErrorNetwork, "read PyPI metadata", err)
	}
	if len(data) > metadataLimit {
		return domain.NewError(domain.ErrorMetadata, "bound PyPI metadata", errors.New("metadata exceeds size limit"))
	}
	if err := json.Unmarshal(data, target); err != nil {
		return domain.NewError(domain.ErrorMetadata, "decode PyPI metadata", err)
	}
	return nil
}

func sourceFromInfo(info releaseInfo) (*domain.SourceReference, *domain.GitReference, error) {
	keys := make([]string, 0, len(info.ProjectURLs))
	for key := range info.ProjectURLs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "source") || strings.Contains(lower, "repository") || strings.Contains(lower, "github") || strings.Contains(lower, "code") {
			return normalizeSourceURL(info.ProjectURLs[key])
		}
	}
	if info.HomePage != "" {
		if _, _, ok := githubRepository(info.HomePage); ok {
			return normalizeSourceURL(info.HomePage)
		}
	}
	return nil, nil, nil
}

func normalizeSourceURL(raw string) (*domain.SourceReference, *domain.GitReference, error) {
	value := strings.TrimSpace(strings.TrimPrefix(raw, "git+"))
	if strings.HasPrefix(value, "git://") {
		value = "https://" + strings.TrimPrefix(value, "git://")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.Scheme != "https" {
		return nil, nil, errors.New("source URL is not an absolute HTTPS URL")
	}
	ref := parsed.Fragment
	parsed.Fragment = ""
	parsed.RawFragment = ""
	parsed.Path = strings.TrimSuffix(parsed.Path, ".git")
	result := &domain.SourceReference{URL: parsed.String()}
	if err := result.Validate(); err != nil {
		return nil, nil, err
	}
	if ref == "" {
		return result, nil, nil
	}
	if strings.ContainsAny(ref, "\x00\r\n") {
		return result, nil, errors.New("source git reference contains invalid characters")
	}
	return result, &domain.GitReference{Ref: ref, Source: result.URL, Immutable: isCommit(ref), Commit: commitValue(ref)}, nil
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
		if parts[0] != "" {
			return parts[0], nil
		}
	}
	return "", errors.New("archive contains no entries")
}

func isCommit(value string) bool {
	if len(value) < 7 || len(value) > 64 {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}

func commitValue(value string) string {
	if isCommit(value) {
		return value
	}
	return ""
}

// Package acquire downloads and inspects untrusted release archives without
// executing package code or extracting files to disk.
package acquire

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/releasecheck/releasecheck/internal/domain"
)

const (
	defaultMaxDownloadBytes = 100 << 20
	defaultMaxEntries       = 10_000
	defaultMaxExpandedBytes = 500 << 20
	defaultMaxFileBytes     = 100 << 20
	defaultMaxRedirects     = 5
	defaultHTTPTimeout      = 30 * time.Second
)

// Limits bound network and archive processing resources. Zero values are
// replaced by the documented defaults by DefaultLimits or NewDownloader.
type Limits struct {
	MaxDownloadBytes int64
	MaxEntries       int
	MaxExpandedBytes int64
	MaxFileBytes     int64
	MaxRedirects     int
	HTTPTimeout      time.Duration
}

// DefaultLimits returns conservative defaults suitable for a CLI invocation.
func DefaultLimits() Limits {
	return Limits{
		MaxDownloadBytes: defaultMaxDownloadBytes,
		MaxEntries:       defaultMaxEntries,
		MaxExpandedBytes: defaultMaxExpandedBytes,
		MaxFileBytes:     defaultMaxFileBytes,
		MaxRedirects:     defaultMaxRedirects,
		HTTPTimeout:      defaultHTTPTimeout,
	}
}

func (l Limits) withDefaults() Limits {
	d := DefaultLimits()
	if l.MaxDownloadBytes == 0 {
		l.MaxDownloadBytes = d.MaxDownloadBytes
	}
	if l.MaxEntries == 0 {
		l.MaxEntries = d.MaxEntries
	}
	if l.MaxExpandedBytes == 0 {
		l.MaxExpandedBytes = d.MaxExpandedBytes
	}
	if l.MaxFileBytes == 0 {
		l.MaxFileBytes = d.MaxFileBytes
	}
	if l.MaxRedirects == 0 {
		l.MaxRedirects = d.MaxRedirects
	}
	if l.HTTPTimeout == 0 {
		l.HTTPTimeout = d.HTTPTimeout
	}
	return l
}

func (l Limits) validate() error {
	if l.MaxDownloadBytes < 1 || l.MaxEntries < 1 || l.MaxExpandedBytes < 1 || l.MaxFileBytes < 1 || l.MaxRedirects < 0 || l.HTTPTimeout <= 0 {
		return errors.New("all acquisition limits must be positive, except redirects which may be zero")
	}
	if l.MaxFileBytes > l.MaxExpandedBytes {
		return errors.New("maximum file size cannot exceed maximum expanded archive size")
	}
	return nil
}

// DownloadOptions configures a Downloader. HTTPClient is copied before its
// timeout and redirect policy are applied; callers retain ownership of the
// original client.
type DownloadOptions struct {
	HTTPClient *http.Client
	TempDir    string
	Limits     Limits
	// AllowPrivateNetworks is intended only for controlled local fixtures or
	// explicitly configured private mirrors. The default is false.
	AllowPrivateNetworks bool
}

// Downloader retrieves bytes into a private temporary file and calculates a
// SHA-256 digest while streaming. It never interprets downloaded bytes as
// executable input.
type Downloader struct {
	client               *http.Client
	tempDir              string
	limits               Limits
	allowPrivateNetworks bool
}

// NewDownloader creates a bounded HTTPS-only downloader.
func NewDownloader(options DownloadOptions) (*Downloader, error) {
	limits := options.Limits.withDefaults()
	if err := limits.validate(); err != nil {
		return nil, domain.NewError(domain.ErrorInput, "validate acquisition limits", err)
	}

	client := http.DefaultClient
	if options.HTTPClient != nil {
		copy := *options.HTTPClient
		client = &copy
	} else {
		copy := *http.DefaultClient
		client = &copy
	}
	client.Timeout = limits.HTTPTimeout
	client.CheckRedirect = func(req *http.Request, previous []*http.Request) error {
		if len(previous) >= limits.MaxRedirects {
			return fmt.Errorf("redirect limit exceeded: %d", limits.MaxRedirects)
		}
		if err := ValidateHTTPSURL(req.URL.String()); err != nil {
			return fmt.Errorf("redirect target: %w", err)
		}
		if err := validateNetworkURL(req.Context(), req.URL, options.AllowPrivateNetworks); err != nil {
			return fmt.Errorf("redirect target network policy: %w", err)
		}
		return nil
	}

	return &Downloader{client: client, tempDir: options.TempDir, limits: limits, allowPrivateNetworks: options.AllowPrivateNetworks}, nil
}

// DownloadedArtifact is a downloaded temporary file. Call Cleanup when the
// caller no longer needs it. The file is created with restrictive permissions.
type DownloadedArtifact struct {
	Path   string
	Size   int64
	SHA256 string
}

// Cleanup removes the temporary artifact. Calling it more than once is safe.
func (a *DownloadedArtifact) Cleanup() error {
	if a == nil || a.Path == "" {
		return nil
	}
	err := os.Remove(a.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err == nil {
		a.Path = ""
	}
	return err
}

// Download retrieves one HTTPS resource, checks its HTTP status and optional
// expected SHA-256, and enforces the configured byte bound.
func (d *Downloader) Download(ctx context.Context, rawURL, expectedSHA256 string) (DownloadedArtifact, error) {
	var result DownloadedArtifact
	if err := ValidateHTTPSURL(rawURL); err != nil {
		return result, domain.NewError(domain.ErrorInput, "validate download URL", err)
	}
	parsedURL, _ := url.Parse(rawURL)
	if err := validateNetworkURL(ctx, parsedURL, d.allowPrivateNetworks); err != nil {
		return result, domain.NewError(domain.ErrorNetwork, "validate download network", err)
	}
	expected, err := normalizeSHA256(expectedSHA256)
	if err != nil {
		return result, domain.NewError(domain.ErrorInput, "validate expected SHA-256", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return result, domain.NewError(domain.ErrorInput, "create download request", err)
	}
	req.Header.Set("Accept", "application/octet-stream")
	response, err := d.client.Do(req)
	if err != nil {
		return result, domain.NewError(domain.ErrorNetwork, "download artifact", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return result, domain.NewError(domain.ErrorNetwork, "download artifact", fmt.Errorf("unexpected HTTP status %s", response.Status))
	}
	if response.ContentLength > d.limits.MaxDownloadBytes {
		return result, domain.NewError(domain.ErrorInput, "check content length", fmt.Errorf("content length %d exceeds limit %d", response.ContentLength, d.limits.MaxDownloadBytes))
	}

	file, err := os.CreateTemp(d.tempDir, "releasecheck-artifact-*")
	if err != nil {
		return result, domain.NewError(domain.ErrorInternal, "create temporary artifact", err)
	}
	result.Path = file.Name()
	removeOnError := true
	defer func() {
		if removeOnError {
			_ = os.Remove(result.Path)
		}
	}()

	hasher := sha256.New()
	writer := io.MultiWriter(file, hasher)
	read := io.LimitReader(response.Body, d.limits.MaxDownloadBytes+1)
	count, err := io.Copy(writer, read)
	if err != nil {
		_ = file.Close()
		return DownloadedArtifact{}, domain.NewError(domain.ErrorNetwork, "read artifact", err)
	}
	if count > d.limits.MaxDownloadBytes {
		_ = file.Close()
		return DownloadedArtifact{}, domain.NewError(domain.ErrorInput, "bound artifact size", fmt.Errorf("artifact exceeds limit %d", d.limits.MaxDownloadBytes))
	}
	if err := file.Close(); err != nil {
		return DownloadedArtifact{}, domain.NewError(domain.ErrorInternal, "close temporary artifact", err)
	}

	digest := hex.EncodeToString(hasher.Sum(nil))
	if expected != "" && !strings.EqualFold(expected, digest) {
		return DownloadedArtifact{}, domain.NewError(domain.ErrorIntegrity, "verify artifact SHA-256", fmt.Errorf("got %s, expected %s", digest, expected))
	}
	result.Size = count
	result.SHA256 = digest
	removeOnError = false
	return result, nil
}

// ValidateHTTPSURL rejects non-HTTPS or non-absolute URLs before any request
// is made. Userinfo is rejected to avoid confusing credential-bearing URLs.
func ValidateHTTPSURL(rawURL string) error {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return errors.New("must be an absolute HTTPS URL without userinfo")
	}
	return nil
}

func validateNetworkURL(ctx context.Context, parsed *url.URL, allowPrivate bool) error {
	if allowPrivate {
		return nil
	}
	host := parsed.Hostname()
	if host == "" {
		return errors.New("URL has no hostname")
	}
	if ip := net.ParseIP(host); ip != nil {
		if restrictedAddress(ip) {
			return fmt.Errorf("private, loopback, link-local, multicast, or unspecified address is not allowed: %s", host)
		}
		return nil
	}
	addresses, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("resolve hostname %q: %w", host, err)
	}
	if len(addresses) == 0 {
		return fmt.Errorf("hostname %q has no addresses", host)
	}
	for _, address := range addresses {
		if restrictedAddress(address) {
			return fmt.Errorf("hostname %q resolves to a restricted address", host)
		}
	}
	return nil
}

func restrictedAddress(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()
}

func normalizeSHA256(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if len(value) != sha256.Size*2 {
		return "", errors.New("SHA-256 must contain 64 hexadecimal characters")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return "", errors.New("SHA-256 must be hexadecimal")
	}
	return strings.ToLower(value), nil
}

// VerifyIntegrity verifies a Subresource Integrity value such as
// "sha512-<base64 digest>" against a downloaded file. The supported
// algorithms are deliberately limited to SHA-256 and SHA-512.
func VerifyIntegrity(filename, integrity string) error {
	for _, token := range strings.Fields(integrity) {
		parts := strings.SplitN(token, "-", 2)
		if len(parts) != 2 {
			continue
		}
		var digest []byte
		file, err := os.Open(filename)
		if err != nil {
			return domain.NewError(domain.ErrorIntegrity, "open file for integrity", err)
		}
		switch parts[0] {
		case "sha256":
			hasher := sha256.New()
			_, err = io.Copy(hasher, file)
			digest = hasher.Sum(nil)
		case "sha512":
			hasher := sha512.New()
			_, err = io.Copy(hasher, file)
			digest = hasher.Sum(nil)
		default:
			_ = file.Close()
			continue
		}
		closeErr := file.Close()
		if err != nil {
			return domain.NewError(domain.ErrorIntegrity, "hash file for integrity", err)
		}
		if closeErr != nil {
			return domain.NewError(domain.ErrorIntegrity, "close file for integrity", closeErr)
		}
		expected, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return domain.NewError(domain.ErrorIntegrity, "decode integrity", err)
		}
		if bytes.Equal(digest, expected) {
			return nil
		}
		return domain.NewError(domain.ErrorIntegrity, "verify integrity", fmt.Errorf("digest mismatch for %s", parts[0]))
	}
	return domain.NewError(domain.ErrorIntegrity, "verify integrity", errors.New("no supported integrity value found"))
}

// ArchiveKind identifies the supported archive container.
type ArchiveKind string

const (
	ArchiveZip     ArchiveKind = "zip"
	ArchiveTarGzip ArchiveKind = "tar.gz"
)

// ArchiveInventory is a bounded logical inventory. No archive member is
// extracted or executed.
type ArchiveInventory struct {
	Kind            ArchiveKind        `json:"kind"`
	Entries         []domain.FileEntry `json:"entries"`
	CompressedBytes int64              `json:"compressed_bytes"`
	ExpandedBytes   int64              `json:"expanded_bytes"`
}

// InspectArchive validates and inventories a ZIP or gzip-compressed TAR file.
func InspectArchive(filename string, limits Limits) (ArchiveInventory, error) {
	limits = limits.withDefaults()
	if err := limits.validate(); err != nil {
		return ArchiveInventory{}, domain.NewError(domain.ErrorInput, "validate archive limits", err)
	}
	file, err := os.Open(filename)
	if err != nil {
		return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "open archive", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "stat archive", err)
	}
	if info.Size() > limits.MaxDownloadBytes {
		return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "bound archive size", fmt.Errorf("archive size %d exceeds limit %d", info.Size(), limits.MaxDownloadBytes))
	}

	kind, err := detectArchiveKind(file)
	if err != nil {
		return ArchiveInventory{}, err
	}
	if kind == ArchiveZip {
		return inspectZIP(file, info.Size(), limits)
	}
	if kind == ArchiveTarGzip {
		return inspectTarGzip(file, info.Size(), limits)
	}
	return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "identify archive format", errors.New("unsupported archive format; expected ZIP or gzip-compressed TAR"))
}

// ReadArchiveFile returns one bounded regular-file member from a validated
// archive. It never writes the member to disk and never interprets it as code.
// The target may be an exact archive path or a basename path such as
// "package.json"; duplicate archive paths are rejected by InspectArchive.
func ReadArchiveFile(filename, target string, limits Limits) ([]byte, error) {
	limits = limits.withDefaults()
	if err := limits.validate(); err != nil {
		return nil, domain.NewError(domain.ErrorInput, "validate archive limits", err)
	}
	if strings.TrimSpace(target) == "" {
		return nil, domain.NewError(domain.ErrorInput, "read archive file", errors.New("archive target is required"))
	}
	if err := validateArchivePath(target); err != nil {
		return nil, domain.NewError(domain.ErrorInput, "validate archive target", err)
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, domain.NewError(domain.ErrorArchive, "open archive", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, domain.NewError(domain.ErrorArchive, "stat archive", err)
	}
	kind, err := detectArchiveKind(file)
	if err != nil {
		return nil, err
	}
	match := func(name string) bool {
		cleaned := cleanArchivePath(name)
		wanted := cleanArchivePath(target)
		return cleaned == wanted || strings.HasSuffix(cleaned, "/"+wanted)
	}
	if kind == ArchiveZip {
		reader, err := zip.NewReader(file, info.Size())
		if err != nil {
			return nil, domain.NewError(domain.ErrorArchive, "read ZIP archive", err)
		}
		for _, entry := range reader.File {
			if !match(entry.Name) || zipEntryKind(entry) != domain.FileRegular {
				continue
			}
			if entry.UncompressedSize64 > uint64(limits.MaxFileBytes) {
				return nil, domain.NewError(domain.ErrorArchive, "bound archive metadata", errors.New("archive member exceeds metadata limit"))
			}
			opened, err := entry.Open()
			if err != nil {
				return nil, domain.NewError(domain.ErrorArchive, "read ZIP metadata", err)
			}
			data, readErr := io.ReadAll(io.LimitReader(opened, int64(entry.UncompressedSize64)+1))
			closeErr := opened.Close()
			if readErr != nil || closeErr != nil {
				if readErr != nil {
					return nil, domain.NewError(domain.ErrorArchive, "read ZIP metadata", readErr)
				}
				return nil, domain.NewError(domain.ErrorArchive, "close ZIP metadata", closeErr)
			}
			if int64(len(data)) > limits.MaxFileBytes {
				return nil, domain.NewError(domain.ErrorArchive, "bound archive metadata", errors.New("archive member exceeds metadata limit"))
			}
			return data, nil
		}
		return nil, domain.NewError(domain.ErrorArchive, "read archive file", os.ErrNotExist)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, domain.NewError(domain.ErrorArchive, "rewind archive", err)
	}
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, domain.NewError(domain.ErrorArchive, "read gzip archive", err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, domain.NewError(domain.ErrorArchive, "read TAR archive", err)
		}
		if !match(header.Name) || (header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA) {
			continue
		}
		if header.Size < 0 || header.Size > limits.MaxFileBytes {
			return nil, domain.NewError(domain.ErrorArchive, "bound archive metadata", errors.New("archive member exceeds metadata limit"))
		}
		data, err := io.ReadAll(io.LimitReader(tarReader, header.Size+1))
		if err != nil {
			return nil, domain.NewError(domain.ErrorArchive, "read TAR metadata", err)
		}
		if int64(len(data)) > header.Size {
			return nil, domain.NewError(domain.ErrorArchive, "bound archive metadata", errors.New("archive member exceeds declared size"))
		}
		return data, nil
	}
	return nil, domain.NewError(domain.ErrorArchive, "read archive file", os.ErrNotExist)
}

func detectArchiveKind(file *os.File) (ArchiveKind, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", domain.NewError(domain.ErrorArchive, "seek archive header", err)
	}
	var header [4]byte
	if _, err := io.ReadFull(file, header[:]); err != nil {
		return "", domain.NewError(domain.ErrorArchive, "read archive header", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", domain.NewError(domain.ErrorArchive, "rewind archive", err)
	}
	if header[0] == 'P' && header[1] == 'K' && (header[2] == 3 || header[2] == 5 || header[2] == 7) {
		return ArchiveZip, nil
	}
	if header[0] == 0x1f && header[1] == 0x8b {
		return ArchiveTarGzip, nil
	}
	return "", domain.NewError(domain.ErrorArchive, "identify archive format", errors.New("unsupported archive header; expected ZIP or gzip-compressed TAR"))
}

func inspectZIP(file *os.File, compressed int64, limits Limits) (ArchiveInventory, error) {
	reader, err := zip.NewReader(file, compressed)
	if err != nil {
		return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "read ZIP archive", err)
	}
	inventory := ArchiveInventory{Kind: ArchiveZip, CompressedBytes: compressed}
	for _, entry := range reader.File {
		kind := zipEntryKind(entry)
		if err := validateEntryBounds(&inventory, entry.Name, uint64(entry.UncompressedSize64), limits); err != nil {
			return ArchiveInventory{}, err
		}
		var digest string
		target := ""
		if kind == domain.FileRegular {
			opened, err := entry.Open()
			if err != nil {
				return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "read ZIP entry", err)
			}
			digest, err = hashArchiveEntry(opened, int64(entry.UncompressedSize64))
			closeErr := opened.Close()
			if err != nil {
				return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "hash ZIP entry", err)
			}
			if closeErr != nil {
				return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "close ZIP entry", closeErr)
			}
		} else if kind == domain.FileSymlink {
			opened, err := entry.Open()
			if err != nil {
				return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "read ZIP link", err)
			}
			target, err = readArchiveTarget(opened, limits.MaxFileBytes)
			closeErr := opened.Close()
			if err != nil {
				return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "read ZIP link target", err)
			}
			if closeErr != nil {
				return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "close ZIP link", closeErr)
			}
		}
		if err := addEntry(&inventory, entry.Name, kind, uint64(entry.UncompressedSize64), digest, target, limits); err != nil {
			return ArchiveInventory{}, err
		}
	}
	return inventory, nil
}

func zipEntryKind(entry *zip.File) domain.FileKind {
	if entry.FileInfo().IsDir() {
		return domain.FileDirectory
	}
	mode := entry.Mode()
	if mode&os.ModeSymlink != 0 {
		return domain.FileSymlink
	}
	if !mode.IsRegular() {
		return domain.FileSpecial
	}
	return domain.FileRegular
}

func inspectTarGzip(file *os.File, compressed int64, limits Limits) (ArchiveInventory, error) {
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "read gzip archive", err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	inventory := ArchiveInventory{Kind: ArchiveTarGzip, CompressedBytes: compressed}
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "read TAR archive", err)
		}
		kind := domain.FileRegular
		switch header.Typeflag {
		case tar.TypeDir:
			kind = domain.FileDirectory
		case tar.TypeSymlink, tar.TypeLink:
			kind = domain.FileSymlink
		case tar.TypeReg, tar.TypeRegA:
			kind = domain.FileRegular
		default:
			kind = domain.FileSpecial
		}
		if err := validateEntryBounds(&inventory, header.Name, uint64(maxInt64(header.Size)), limits); err != nil {
			return ArchiveInventory{}, err
		}
		digest := ""
		target := ""
		if kind == domain.FileRegular {
			digest, err = hashArchiveEntry(tarReader, header.Size)
			if err != nil {
				return ArchiveInventory{}, domain.NewError(domain.ErrorArchive, "hash TAR entry", err)
			}
		} else if kind == domain.FileSymlink {
			target = header.Linkname
		}
		if err := addEntry(&inventory, header.Name, kind, uint64(maxInt64(header.Size)), digest, target, limits); err != nil {
			return ArchiveInventory{}, err
		}
	}
	return inventory, nil
}

func addEntry(inventory *ArchiveInventory, rawName string, kind domain.FileKind, size uint64, digest, target string, limits Limits) error {
	if err := validateEntryBounds(inventory, rawName, size, limits); err != nil {
		return err
	}
	cleanedPath := cleanArchivePath(rawName)
	inventory.Entries = append(inventory.Entries, domain.FileEntry{Path: cleanedPath, Kind: kind, SHA256: digest, Size: int64(size), Target: target})
	inventory.ExpandedBytes += int64(size)
	return nil
}

func readArchiveTarget(reader io.Reader, maxBytes int64) (string, error) {
	value, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(value)) > maxBytes {
		return "", fmt.Errorf("link target exceeds limit %d", maxBytes)
	}
	if strings.ContainsRune(string(value), '\x00') {
		return "", errors.New("link target contains NUL")
	}
	return string(value), nil
}

func validateEntryBounds(inventory *ArchiveInventory, rawName string, size uint64, limits Limits) error {
	if err := validateArchivePath(rawName); err != nil {
		return domain.NewError(domain.ErrorArchive, "validate archive path", err)
	}
	if len(inventory.Entries) >= limits.MaxEntries {
		return domain.NewError(domain.ErrorArchive, "bound archive entries", fmt.Errorf("entry count exceeds limit %d", limits.MaxEntries))
	}
	cleanedPath := cleanArchivePath(rawName)
	for _, existing := range inventory.Entries {
		if existing.Path == cleanedPath {
			return domain.NewError(domain.ErrorArchive, "validate archive paths", fmt.Errorf("duplicate archive path %q", rawName))
		}
	}
	if size > uint64(limits.MaxFileBytes) {
		return domain.NewError(domain.ErrorArchive, "bound archive file", fmt.Errorf("entry %q exceeds file limit %d", rawName, limits.MaxFileBytes))
	}
	if size > uint64(limits.MaxExpandedBytes-inventory.ExpandedBytes) {
		return domain.NewError(domain.ErrorArchive, "bound archive expansion", fmt.Errorf("expanded bytes exceed limit %d", limits.MaxExpandedBytes))
	}
	return nil
}

func hashArchiveEntry(reader io.Reader, expectedSize int64) (string, error) {
	if expectedSize < 0 {
		return "", errors.New("archive entry size is negative")
	}
	hasher := sha256.New()
	count, err := io.CopyN(hasher, reader, expectedSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if count != expectedSize {
		return "", fmt.Errorf("archive entry size mismatch: read %d, expected %d", count, expectedSize)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func validateArchivePath(rawName string) error {
	if rawName == "" || strings.ContainsRune(rawName, '\x00') {
		return errors.New("archive path is empty or contains NUL")
	}
	cleaned := cleanArchivePath(rawName)
	if strings.HasPrefix(strings.ReplaceAll(rawName, "\\", "/"), "/") || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, ":") {
		return fmt.Errorf("unsafe archive path %q", rawName)
	}
	return nil
}

func cleanArchivePath(rawName string) string {
	return path.Clean(strings.ReplaceAll(rawName, "\\", "/"))
}

func maxInt64(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

# Release Process

This document describes the v0.1 release path. A release is made only from a tag matching `vMAJOR.MINOR.PATCH`.

## Local preparation

Run the repository checks from the repository root:

```text
go test ./...
go vet ./...
go build ./cmd/releasecheck
```

Inspect `git diff`, `git diff --check`, the generated help output, and the release notes before tagging. The working tree must contain only intended changes. Do not tag a release with known failing checks or unreviewed generated files.

## Release workflow

1. Update `CHANGELOG.md` and `VERSION` for the release.
2. Run the local checks and review the resulting diff.
3. Create and push a tag such as `v0.1.0`.
4. GitHub Actions runs `.github/workflows/release.yml`.
5. The workflow builds Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 archives.
6. The workflow publishes the archives and `SHA256SUMS` to the GitHub release for that tag.

The workflow uses only the repository tag, checked-out source, and the Go version declared by `go.mod`. It does not execute package installation, package code, or arbitrary build hooks.

## Reproducibility boundary

ReleaseCheck uses `-trimpath`, `-buildvcs=false`, `CGO_ENABLED=0`, normalized archive ownership, normalized archive ordering, normalized file timestamps, and SHA-256 checksums. These choices make v0.1 artifacts reproducible enough to compare and audit, but the project does not claim bit-for-bit reproducibility until independent builds have been measured and documented.

The release script injects the source commit and release version; `releasecheck --version` exposes both. Build inputs include the Go toolchain, so independent verification should use the Go version selected from `go.mod`.

## Security and permissions

CI has read-only repository permissions. The release job is the only job with `contents: write`, and it runs only for version-shaped tags. GitHub Actions are pinned to full commit SHAs. Release signing, package-manager publishing, and provenance generation are outside this phase and must not be implied by a GitHub release.

## Rollback

Do not move a published tag silently. If an artifact is wrong, document the problem, revoke or delete the GitHub release according to repository policy, and publish a new patch version with a clear changelog entry.

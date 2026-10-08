# ReleaseCheck

![CI](https://github.com/ReleaseCheck/releasecheck/actions/workflows/ci.yml/badge.svg?branch=main) ![Go 1.25](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go) ![Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue)

ReleaseCheck checks how a published npm or PyPI package relates to its claimed source release, then reports deterministic evidence and limitations.

[Core](https://github.com/ReleaseCheck/releasecheck) · [GitHub Action](https://github.com/ReleaseCheck/releasecheck-action) · [User documentation](https://github.com/ReleaseCheck/releasecheck-docs) · [Security policy](SECURITY.md) · [Roadmap](ROADMAP.md)

**Project status:** pre-v0.1.0 · Phase 15 complete · Phase 16 independent re-audit required.

## Contents

[How it works](#problem) · [Quick start](#installation) · [Usage](#usage) · [Reports](#what-releasecheck-verifies) · [Security boundaries](#security-boundary) · [Limitations](#current-limitations) · [Contributing](#contributing)

## Problem

Package consumers can often download an artifact and discover a repository URL, but the relationship between the bytes received and the claimed source release is not always obvious. ReleaseCheck collects that relationship as inspectable evidence:

```mermaid
flowchart TD
    R[Public registry] --> A[Published artifact]
    A --> H[Artifact hash and inventory]
    A --> M[Release metadata]
    M --> S[Claimed source repository]
    S --> G[Git tag or commit]
    G --> T[Source snapshot]
    H --> C[Deterministic comparison]
    T --> C
    M --> P[Provenance and attestation evidence]
    C --> O[Machine-readable and human report]
    P --> O
```

Differences are evidence, not automatic proof of maliciousness. Build systems may legitimately transform source into distributable files. Reports distinguish facts, inferences, warnings, limitations, and verdicts.

## v0.1 scope

The current release candidate scope is limited to npm and PyPI. ReleaseCheck retrieves registry metadata and artifacts, resolves source references where possible, compares archive contents deterministically, records security-relevant archive and metadata observations, consumes supported provenance evidence, and emits human-readable, JSON, and SARIF reports with meaningful exit statuses.

It does not install packages, run package code, run lifecycle scripts, run Python build backends, execute arbitrary builds, scan for vulnerabilities, generate SBOMs, or implement Sigstore or SLSA.

## Project status

The initial build and Phase 15 audit remediation are complete. The hosted cross-platform test matrix passes. Phase 16 independent re-audit is required before Phase 17 release-workflow validation and the v0.1 release decision. No public v0.1.0 release exists. See [FINAL_AUDIT.md](FINAL_AUDIT.md) for findings and [ROADMAP.md](ROADMAP.md) for the current phase register.

## Repository Map

ReleaseCheck is organized under the [ReleaseCheck GitHub organization](https://github.com/ReleaseCheck):

| Repository | Responsibility |
| --- | --- |
| [releasecheck](https://github.com/ReleaseCheck/releasecheck) | Core Go CLI, verification engine, tests, fixtures, technical source of truth, and release engineering |
| [releasecheck-action](https://github.com/ReleaseCheck/releasecheck-action) | GitHub Actions integration that invokes released core binaries |
| [releasecheck-docs](https://github.com/ReleaseCheck/releasecheck-docs) | Curated public user documentation and learning material |

The core repository remains the authoritative home of implementation, architecture, security decisions, roadmap, and contributor engineering guidance. The Action does not reimplement verification, and the docs repository does not duplicate the core source tree. Additional repositories are discovery-gated; repository count is not a project success metric.

## Installation

There is no package-manager distribution yet. Build from a checked-out release or commit with Go 1.25 or newer:

```text
go build -trimpath -o releasecheck ./cmd/releasecheck
```

The Phase 12 release workflow publishes archives for Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 when a version tag is created. See [docs/RELEASE.md](docs/RELEASE.md).

## Usage

Build the command locally:

```text
go build ./cmd/releasecheck
```

Run the offline test suite and static checks:

```text
go test ./...
go vet ./...
gofmt -l .
```

Verify a selected release from npm or PyPI:

```text
releasecheck verify npm [flags] NAME [VERSION]
releasecheck verify pypi [flags] NAME [VERSION]
```

Useful flags are `--json`, `--sarif`, `--output PATH`, `--artifact FILENAME` for PyPI, and `--timeout DURATION`. The command never installs packages, runs lifecycle scripts, executes Python build logic, or invokes package managers. Exit code `0` means `MATCH`; `1` means `REVIEW` or `INCOMPLETE`; `2` means usage error; and `3` means operational or rendering error.

For output contracts and examples, see [docs/REPORTS.md](docs/REPORTS.md) and [docs/USAGE.md](docs/USAGE.md).

## What ReleaseCheck verifies

ReleaseCheck collects registry metadata, the selected artifact, artifact hashes and file inventory, claimed source information, a source snapshot where supported, deterministic file comparisons, structural security observations, and available provenance evidence. It reports the chain as evidence rather than reducing it to an unsupported security guarantee.

`MATCH` means the available comparison evidence is identical and no review signal or limitation prevented that result. `REVIEW` means differences or security/provenance observations require inspection. `INCOMPLETE` means important evidence could not be obtained; it does not silently pass CI.

## What it does not verify

ReleaseCheck does not prove that code is benign, detect all malware, reproduce arbitrary builds, verify every signature or certificate chain, establish maintainer identity, or prove that a source repository is honest. It does not install packages, execute package code, run build backends, scan vulnerabilities, generate SBOMs, or implement Sigstore or SLSA.

Wheels, bundled JavaScript, compiled extensions, generated files, and other transformed distributions may legitimately differ from source. A difference is evidence for review, not an automatic maliciousness verdict.

## CI usage

The CLI is suitable for CI because its exit status is stable and its JSON/SARIF output is machine-readable:

```text
releasecheck verify npm --json --output releasecheck.json package-name
```

Use `--sarif` for systems that ingest SARIF. Keep reports and package inputs treated as untrusted data, and do not pass package metadata into shell commands. See [docs/CI.md](docs/CI.md) for a complete example.

## Current limitations

The v0.1 implementation supports npm and PyPI. Source retrieval is currently limited to supported GitHub source URLs, bounded manifest observations cover supported `package.json`, `setup.py`, and `pyproject.toml` members, and provenance evidence is parsed but not cryptographically verified. Artifact downloads reject restricted network destinations by default. The internal packages are not a compatibility SDK. See [docs/DESIGN.md](docs/DESIGN.md) for the complete limitation record.

## Relationship to existing tools

ReleaseCheck overlaps with registry integrity, provenance, reproducibility, and supply-chain tools. It deliberately does not replace them. Its current focus is a small, explicit evidence chain across npm and PyPI, with safe non-executing inspection and common report semantics. See [docs/COMPETITIVE_LANDSCAPE.md](docs/COMPETITIVE_LANDSCAPE.md) for the researched boundaries and source links.

## Security boundary

ReleaseCheck processes untrusted package artifacts. It must never execute package code or installation/build hooks merely to inspect a release. Archive extraction and parsing is bounded and path-safe. See [SECURITY.md](SECURITY.md) and the threat model in [docs/DESIGN.md](docs/DESIGN.md).

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) before making changes. The repository is intentionally documentation-led: implementation work must satisfy the phase acceptance criteria and update the durable project records.

## License

ReleaseCheck is licensed under the Apache License 2.0. See [LICENSE](LICENSE).

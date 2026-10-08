# ReleaseCheck

ReleaseCheck is planned as a deterministic Go CLI and SDK for examining whether a package published to a public registry corresponds to its claimed source repository and release reference.

The repository is being built phase by phase. The current codebase contains the domain model, safe artifact acquisition, internal npm and PyPI verification paths, deterministic comparison and security observations, provenance evidence parsing, internal report renderers, an offline fixture matrix, and the initial CLI. A public SDK remains future roadmap work.

## Problem

Package consumers can often download an artifact and discover a repository URL, but the relationship between the bytes received and the claimed source release is not always obvious. ReleaseCheck will collect that relationship as inspectable evidence:

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

Differences are evidence, not automatic proof of maliciousness. Build systems may legitimately transform source into distributable files. Reports will distinguish facts, inferences, warnings, limitations, and verdicts.

## v0.1 scope

The initial release is limited to npm and PyPI. It will retrieve registry metadata and artifacts, resolve source references where possible, compare archive contents deterministically, record security-relevant archive and metadata observations, consume supported provenance evidence, and emit human-readable, JSON, and SARIF reports with meaningful exit statuses.

It will not install packages, run package code, run lifecycle scripts, run Python build backends, execute arbitrary builds, scan for vulnerabilities, generate SBOMs, or implement Sigstore or SLSA.

## Project status

Phases 0 through 11 are complete: discovery, repository foundation, domain contracts, safe artifact acquisition, fixture-backed npm and PyPI verification paths, deterministic comparison hardening, structural security analysis, provenance evidence parsing, stable report rendering, offline fixture coverage, and the initial CLI. Phase 12, CI and release engineering, is next. See [ROADMAP.md](ROADMAP.md), [PROJECT_CONTEXT.md](PROJECT_CONTEXT.md), and [docs/DESIGN.md](docs/DESIGN.md).

## Usage

Build the command locally:

```text
go build ./cmd/releasecheck
```

Verify a selected release from npm or PyPI:

```text
releasecheck verify npm [flags] NAME [VERSION]
releasecheck verify pypi [flags] NAME [VERSION]
```

Useful flags are `--json`, `--sarif`, `--output PATH`, `--artifact FILENAME` for PyPI, and `--timeout DURATION`. The command never installs packages, runs lifecycle scripts, executes Python build logic, or invokes package managers. Exit code `0` means `MATCH`; `1` means `REVIEW` or `INCOMPLETE`; `2` means usage error; and `3` means operational or rendering error.

## Security boundary

ReleaseCheck processes untrusted package artifacts. It must never execute package code or installation/build hooks merely to inspect a release. Archive extraction and parsing will be bounded and path-safe. See [SECURITY.md](SECURITY.md) and the threat model in [docs/DESIGN.md](docs/DESIGN.md).

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) before making changes. The repository is intentionally documentation-led: implementation work must satisfy the phase acceptance criteria and update the durable project records.

## License

ReleaseCheck is licensed under the Apache License 2.0. See [LICENSE](LICENSE).

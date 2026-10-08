# ReleaseCheck Security and Release Audit

Audit date: 2026-10-08

Audit scope: current `main` after the Phase 15 remediation changes. This document supersedes the stale Phase 14 checkpoint wording; it is an audit update, not a claim that v0.1 is released.

## Executive result

**Recommendation: CONDITIONAL. Do not declare v0.1 complete yet.**

The non-execution boundary remains credible, and the remediation adds bounded manifest analysis, default-deny restricted-network checks for artifact downloads, full-length commit recognition, and a repository line-ending policy for cross-platform formatting. Hosted CI run `37852987328` passed all declared jobs for commit `8e0d05c`; an independent re-audit and release-workflow validation must still review these changes.

## Verified strengths

1. Package code is not executed. ReleaseCheck does not invoke package managers, lifecycle scripts, setup.py, build backends, shell commands, or metadata-derived commands.
2. Downloads use HTTPS, timeouts, bounded response sizes, streaming SHA-256, restrictive temporary files, cleanup, and bounded archive inspection.
3. Artifact download redirects are revalidated and restricted destination addresses are rejected by default. Loopback, private, link-local, multicast, and unspecified IP results are denied unless a controlled test/private-mirror opt-in is explicitly supplied.
4. ZIP and gzip/TAR inspection is non-extracting. Paths, duplicates, entry counts, file sizes, expanded size, symlinks, and special entries are handled explicitly.
5. npm and PyPI paths now read bounded `package.json`, `setup.py`, and `pyproject.toml` members in memory and include supported observations in the CLI report. No metadata is executed.
6. Seven- or otherwise abbreviated hexadecimal references are no longer treated as immutable commits; only full 40-character Git object IDs receive immutable classification.
7. Comparison, provenance parsing, reports, exit codes, fixtures, and deterministic ordering remain covered by offline tests.
8. CI action references are pinned to full commit SHAs and workflows use least privilege.

## Finding status

### F-001: Hosted CI and release workflow evidence

Status: **RESOLVED for the remediation commit.**

The previous hosted run failed the Windows formatting check because checkout line endings were not declared. `.gitattributes` now declares LF for Go and workflow source files. GitHub Actions run `37852987328` passed Ubuntu, macOS, Windows, and race jobs for commit `8e0d05c`. A first release workflow must still be reviewed before claiming release publication works.

### F-002: Source archive binding

Status: **PARTIALLY MITIGATED; ACCEPTED LIMITATION for this candidate.**

Only full 40-character references are classified as immutable. Mutable tags and abbreviated references remain explicitly represented as mutable or limited evidence. The GitHub archive response is still not independently verified against a Git object/tree proof. ReleaseCheck must not describe this as cryptographic source-to-artifact proof.

### F-003: Manifest security analysis

Status: **RESOLVED for the supported v0.1 observations.**

Manifest bytes are read from accepted archive members with bounded in-memory reads. npm lifecycle scripts and Python setup metadata are included in the result path. Detailed TOML build-backend interpretation, malware detection, and package execution remain out of scope.

### F-004: Registry-controlled network destinations

Status: **RESOLVED for artifact downloads; limited for local-service isolation.**

The default artifact downloader rejects restricted destinations and rechecks redirect targets. ReleaseCheck remains a local developer tool, not a sandbox or network-isolated service. Metadata/API endpoints are configured HTTPS services and must not be replaced with arbitrary untrusted endpoints in a hosted deployment.

### F-005: Static-analysis and environment evidence

Status: **ACCEPTED ENVIRONMENT LIMITATION; independent review still required.**

The workstation does not provide every security analysis tool. `go vet` and the offline test suite pass locally, and the hosted race job passed in run `37852987328`. `golangci-lint`, `gosec`, and `govulncheck` are not claimed as passed; adding them remains a separate reviewed CI decision.

## Verification performed locally

- `go test -p 1 ./...`: passed
- `go vet ./...`: passed
- `gofmt -l .`: no output
- `git diff --check`: passed
- restricted-network rejection test: passed
- archive manifest wiring tests and existing security fixtures: passed

## Remaining limitations

- No cryptographic verification of DSSE, Sigstore, transparency logs, or registry trust roots.
- No cryptographic proof that a GitHub archive response is the exact tree of the claimed commit.
- No malware, vulnerability, SBOM, or arbitrary-build analysis.
- No stable public SDK compatibility promise.
- Local/private mirror use requires explicit controlled configuration and is not a hosted isolation boundary.
- A reviewed release workflow and independent re-audit remain required.

## Release gate

Do not publish or describe `v0.1.0` as complete until:

1. the current hosted CI matrix passes or each failure has an explicit accepted owner and rationale;
2. the independent Codex re-audit reviews the remediation commit;
3. the release workflow produces and validates intended artifacts;
4. README, docs, roadmap, project context, and this audit agree on the same status;
5. the final recommendation is changed by evidence rather than deadline pressure.

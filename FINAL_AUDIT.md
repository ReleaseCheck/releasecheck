# ReleaseCheck Final Audit

Audit date: 2026-10-08

Auditor: independent Codex review of the repository state at the Phase 14 checkpoint

Audited commit before fixes: `db1344a Complete phase 13 contributor readiness`

Audit fixes are recorded in the follow-up Phase 14 checkpoint commit.

## Scope

This audit covered:

- architecture and domain contracts;
- artifact acquisition, redirects, limits, temporary files, and archive handling;
- npm and PyPI metadata, artifact, and source resolution;
- deterministic comparison and report ordering;
- security observations and non-execution boundaries;
- provenance parsing and artifact binding;
- CLI behavior, output files, exit codes, and report schemas;
- fixtures, tests, vet, formatting, CI, release workflows, documentation, claims, license, and dead-code/TODO searches.

Repository state, tests, and documented decisions were treated as authoritative.

## Verified strengths

1. Package code is not executed. There is no package-manager invocation, shell execution, `eval`, `setup.py` execution, build-backend execution, or metadata-derived command path in the inspected implementation.
2. Artifact downloads require HTTPS, use request timeouts, enforce response-size bounds, hash while streaming, use restrictive temporary files, and clean up on failure.
3. ZIP and gzip/TAR inspection is non-extracting. Archive paths, duplicate normalized paths, entry counts, file sizes, expanded size, symlink metadata, and special entries are handled explicitly.
4. Comparison is path-based and hash-based. Missing hashes are `unverifiable`; differences are not labelled malicious; output is normalized before reporting.
5. npm and PyPI flows verify registry-provided integrity where supported and preserve missing or ambiguous source evidence as limitations.
6. Provenance parsing validates structure and artifact subject binding without pretending that DSSE, Sigstore, transparency logs, or registry trust roots were verified.
7. Human, JSON schema `1.0`, and SARIF `2.1.0` renderers share the same evidence and verdict logic. `MATCH` is not presented as a safety guarantee.
8. The default fixture suite is offline and includes malformed, unsafe-path, symlink, size-boundary, metadata, comparison, and provenance cases.
9. CI is least-privilege and action references are pinned to full commit SHAs. Release automation is tag-gated and documents its reproducibility boundary.
10. README, security policy, contributor guidance, report semantics, adapter guidance, and limitations avoid fabricated adoption, novelty, or security-certification claims.

## Findings

### F-001: Hosted CI has not executed

Severity: Release blocker

Status: Unresolved environmental prerequisite

Evidence: `.github/workflows/ci.yml`, `.github/workflows/release.yml`, and repository state show no configured remote. The local equivalent normal suite, vet, formatting, build, cross-target compilation, and repeated-build hash checks have been run, but GitHub-hosted Windows/Linux/macOS CI and the Ubuntu race job have not executed.

Impact: The project cannot honestly claim that the committed workflows pass on all target runners or that release publication works in GitHub's environment.

Required action: Configure the intended remote, push the checkpoint, run CI, inspect every matrix job, and perform a dry-run or reviewed first release workflow before declaring v0.1 complete.

### F-002: Source archive is not cryptographically bound to the claimed commit

Severity: Medium

Status: Unresolved documented limitation

Evidence: `internal/npm/npm.go:239-245` and `internal/pypi/pypi.go:250-256` construct a GitHub API tarball URL from the claimed reference and compare its contents. The implementation does not independently verify a returned archive's tree against the claimed commit or record a trusted source-archive binding.

Impact: A successful comparison establishes what the retrieved GitHub response contained, not cryptographic proof that those bytes came from the intended commit. Movable tags and API/archive regeneration remain relevant limitations.

Required action: Keep the limitation prominent. A future stronger mode should resolve the commit through a trusted Git object/tree API or equivalent evidence and record the binding; it must not silently upgrade the current result.

### F-003: Manifest security analysis is not wired into live verification

Severity: Medium

Status: Unresolved scope limitation

Evidence: `internal/security/security.go` provides `AnalyzeNPMManifest` and `AnalyzePythonMetadata`, but `internal/cli/cli.go:130-175` only analyzes the archive inventory. `docs/DESIGN.md` and `PROJECT_CONTEXT.md` correctly disclose that registry adapters do not yet provide manifest bytes.

Impact: The CLI currently cannot report package.json lifecycle scripts or setup.py presence for every live artifact, even though those are important release-integrity signals.

Required action: Either wire safe, bounded manifest-byte extraction into the adapter/report pipeline with fixtures, or keep these signals explicitly described as library-level capabilities rather than complete CLI coverage. Do not claim comprehensive install-metadata detection in v0.1.

### F-004: Arbitrary HTTPS artifact URLs remain a network trust boundary

Severity: Medium

Status: Unresolved residual risk

Evidence: `internal/acquire/acquire.go:165-188` requires HTTPS and validates redirects, but does not restrict DNS resolution or reject loopback, private, link-local, or other internal address ranges. Registry metadata controls the artifact URL consumed by the downloader.

Impact: A compromised or untrusted registry response could cause a local invocation to request an internal HTTPS endpoint. HTTPS alone authenticates the endpoint only if its certificate is trusted; it is not an SSRF defense.

Required action: Decide and document the network policy before a hosted or server-side use case. A future hardened mode should use a reviewed transport/DNS policy and preserve explicit opt-in support for private mirrors. The current CLI should be treated as a local developer tool, not a network-isolated service.

### F-005: Hosted security/static-analysis tooling is not locally available

Severity: Low

Status: Unresolved environment limitation

Evidence: `golangci-lint`, `gosec`, and `govulncheck` are unavailable locally. Windows race runs also fail in the installed MSYS2 GCC linker or temporary test-binary cleanup, although normal tests and vet pass.

Impact: This workstation cannot independently reproduce all static-analysis and race evidence expected by the roadmap.

Required action: Run the configured Ubuntu race job and add any deliberately selected static-analysis tool only after reviewing its dependency and action-pinning implications. Do not report local race/static-analysis success from this machine.

## Fixes made during this audit

- Provenance digest selection now sorts subject algorithms before choosing a matching digest, removing map-iteration nondeterminism.
- Domain, evidence, security, comparison, and provenance sorting now includes tie-breaker fields, making duplicate-key report output deterministic.
- CLI report-file writes now apply mode `0600` to existing files as well as newly created files, matching the documented permission boundary.
- Regression tests cover deterministic digest selection, tied report ordering, and report-file replacement behavior.

## Verification evidence

Passed locally after the fixes:

- `go vet ./...`
- `go build ./cmd/releasecheck`
- `gofmt -l .`
- `git diff --check`
- Markdown relative-link check
- Focused changed-package tests and the previously established offline fixture tests reached passing test results. Aggregate Windows runs remain environment-limited: the host can deny launching or removing a temporary `*.test.exe` even after individual package tests report `ok`.

Environment-limited:

- `go test -race ./...`: local MSYS2 GCC cannot launch `collect2.exe` for race linking.
- Some normal aggregate runs: Windows temporary test-binary cleanup can return `Access is denied` after package tests pass.
- GitHub-hosted CI and release workflow: no configured remote.

## Unresolved limitations

- No cryptographic verification of DSSE/Sigstore/PyPI/npm provenance trust chains.
- No cryptographic source-archive-to-commit binding beyond the requested GitHub API reference.
- No comprehensive live CLI manifest analysis for package.json, setup.py, or pyproject.toml.
- GitHub-only source retrieval in the initial verification paths.
- No public compatibility SDK.
- No private-mirror configuration in the CLI.
- No malware, vulnerability, SBOM, or arbitrary-build analysis.
- Local SSRF defenses are not complete for registry-controlled HTTPS artifact URLs.
- Hosted matrix and release execution remain unverified.

## Release recommendation

**CONDITIONAL / DO NOT DECLARE V0.1 COMPLETE YET.**

The implementation has a credible secure non-execution foundation and the core deterministic evidence/reporting paths are well tested. The remaining findings are material to a security-sensitive release: hosted CI has not run, source identity is not cryptographically bound, live manifest analysis is incomplete, and the network trust boundary needs a deliberate SSRF policy. Phase 14 should remain `AUDIT REQUIRED` until hosted CI is executed and the maintainer explicitly accepts or remediates F-002 through F-005.

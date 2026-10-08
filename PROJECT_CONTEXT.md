# ReleaseCheck Project Context

This file is the durable memory of the project. Repository state is authoritative. Do not trust stale chat context over the filesystem, git state, tests, or documented decisions.

## Identity

- Name: ReleaseCheck
- Planned language: Go
- Module: `github.com/releasecheck/releasecheck` (hosting namespace remains to be confirmed before publishing)
- Current state: post-Phase-14 audit, `AUDIT REQUIRED`, pre-v0.1 release
- Current release target: defensible `v0.1.0`, not yet released
- Primary ecosystems: npm and PyPI
- Deadline: 12:00 PM WAT, October 9, 2026

## Problem and product definition

ReleaseCheck is a deterministic release-integrity analysis CLI for npm and PyPI. It collects evidence about whether a package published to a public registry corresponds to its claimed source repository and release. It compares downloaded distribution contents with a source snapshot when the source identity and reference can be resolved. It reports what is known, inferred, unavailable, different, or limited. A stable public SDK is future work, not a current product promise.

The product is not a SaaS dashboard, chatbot, AI product, generic package scanner, SBOM generator, vulnerability scanner, package installer, build-reproduction engine, generic GitHub bot, AI verdict engine, SLSA implementation, Sigstore replacement, repository health checker, or dashboard-first SaaS.

## Engineering principles

1. Deterministic technical truth.
2. Security before breadth.
3. Reproducibility and offline fixtures.
4. Narrow scope and honest reporting.
5. Excellent documentation and contributor readiness.
6. Strong automated tests and CI.
7. Long-term maintainability and real developer utility.
8. AI may summarize deterministic evidence later, but never decides correctness or trustworthiness.

## Frozen v0.1 scope

In scope: npm and PyPI metadata and artifact acquisition; SHA-256 and registry-provided hash evidence; source repository and git reference resolution where possible; safe archive inspection; deterministic file-set and hash comparison; security-relevant structural observations; supported provenance/attestation evidence extraction; human, JSON, and SARIF reports; stable exit statuses; offline fixtures and CI.

Out of scope: crates.io, Go modules, generic vulnerability scanning, SBOM generation, arbitrary builds, package installation, package execution, malware detection, a hosted service, GitHub App, AI decisions, implementation of Sigstore/SLSA, and a large adapter framework.

## Current implementation

The repository currently contains a Go CLI with npm and PyPI verification paths; bounded HTTPS acquisition; SHA-256 and supported npm integrity handling; safe ZIP/TAR.GZ inventory; source and artifact file comparison; deterministic structural security observations; provenance structure parsing and artifact binding; human, JSON, and SARIF reports; stable exit codes; offline fixtures; contributor documentation; CI and release workflows; and an independent final-audit report. The code is real and tested, but the public release gate is not complete.

The current CLI is:

```text
releasecheck verify npm [flags] NAME [VERSION]
releasecheck verify pypi [flags] NAME [VERSION]
```

## Current milestone

The project is **post-Phase-14 audit / audit-required / pre-v0.1 release**. Phase 14 found a credible non-execution foundation plus unresolved release and trust-boundary decisions. Phase 15 audit remediation is the next executable phase. Future work is organized in ROADMAP.md as v0.1 phases 15-19, v1 phases 20-29, v2 phases 30-39, and discovery-gated v3+ reserved bands.

## Architecture and security decisions

- Go standard library first; dependencies require a documented reason.
- Small registry adapter interfaces, with npm and PyPI implementations behind them.
- Evidence is immutable data assembled by stages; reports consume evidence rather than recomputing security conclusions.
- Artifact and source retrieval are separate from comparison.
- A commit is stronger than a movable branch or tag; reports preserve the exact resolution and limitation.
- Wheels and other built distributions are not byte-for-byte source comparisons by default.
- No package manager is invoked on untrusted input.
- JSON schema and exit semantics are versioned before CLI polish.
- Phase 2 domain model: one internal data-oriented package with small `RegistryAdapter` and `Cache` interfaces; standard `context.Context`; validation only for cross-ecosystem invariants.
- Phase 3 acquisition: `internal/acquire` is HTTPS-only, bounded, SHA-256-aware, temporary-file based, and archive-inspection-only; it never extracts or executes package content.
- Phase 4 npm path: the adapter consumes packument `dist`, `repository`, and `gitHead` metadata, verifies npm SRI when present, supports GitHub source snapshots, and reports unsupported/missing source data as limitations.
- Phase 5 PyPI path: the adapter uses project/release JSON, retains all release files, verifies PyPI SHA-256 digests, prefers sdists, reports wheel limitations, and retrieves source only from an explicit Git reference supplied in source metadata.
- Phase 6 comparison engine: inventory paths are canonical slash-separated relative paths; only a declared archive root may be removed. Duplicate, unsafe, unsupported, or malformed entries are rejected. Regular-file identity requires two valid matching SHA-256 hashes; missing hashes are `unverifiable`. Symlink targets are compared as metadata without following links, and all comparison results are path-ordered.
- Phase 7 security analysis: `internal/security` emits stable observations for symlinks, unsafe link targets, special entries, configured archive/file size thresholds, npm lifecycle scripts, `setup.py`, and malformed or oversized metadata. Findings are typed as known or invalid with info/warning severity; the analyzer never executes content and never labels malware or intent.
- Phase 8 provenance: `internal/provenance` parses PyPI PEP 740 provenance objects and npm DSSE/bundle-shaped evidence, binds single in-toto subjects to filename and available digests, extracts supported predicate/source fields, and reports absent, unavailable, invalid, and insufficient states. It does not verify DSSE signatures, Sigstore roots, Rekor, Fulcio, TUF, or registry trust.
- Phase 9 reporting: `internal/report` normalizes evidence and renders stable JSON schema 1.0, human output, and SARIF 2.1.0. Verdict precedence is `INCOMPLETE` for missing source/comparisons, `REVIEW` for differences, warnings, invalid/insufficient provenance, and `MATCH` only for complete identical comparison evidence without warnings. `INCOMPLETE` maps to exit code 1.
- Phase 10 test suite: `tests/fixture_matrix_test.go` covers matching, artifact-only, source-only, modified, malformed, install metadata, missing repository, unavailable source, unsafe path, symlink, size boundary, and provenance variation cases. Checked-in metadata/provenance samples are inert; archive bytes are generated deterministically with Go's standard library. The default suite is offline and uses local test servers only.
- Phase 11 CLI: `releasecheck verify npm [flags] NAME [VERSION]` and `releasecheck verify pypi [flags] NAME [VERSION]` wire the existing adapters to human, JSON, and SARIF output. `--artifact` selects a PyPI file, `--timeout` bounds operations, `--output` writes a report with restrictive permissions, and exit codes follow the report contract. No public SDK is exposed yet because the internal adapter/result lifecycle is not stable enough to promise compatibility.
- Phase 12 release engineering: CI tests Ubuntu, Windows, and macOS with read-only permissions, runs an Ubuntu race job, pins third-party actions to immutable release commit SHAs, and builds the CLI for five target pairs from version tags. Release archives use `-trimpath`, `-buildvcs=false`, disabled CGO, normalized archive metadata, and SHA-256 checksums. Reproducibility is explicitly described as measured enough for v0.1 preparation, not a bit-for-bit guarantee.
- Phase 13 documentation: README, usage, report-contract, CI, registry-adapter, and scoped good-first-issue guidance now describe the implemented CLI and its boundaries. Contributor and feature-request templates require deterministic tests, security impact, evidence, scope, and limitations. Documentation intentionally makes no adoption, novelty, safety, or hosted-CI claims.
- License: Apache-2.0 for a permissive license with an explicit patent grant and familiar enterprise/open-source reuse terms; MIT remains a possible future reconsideration only through a documented project decision.
- Untrusted bytes are downloaded over HTTPS with timeouts and bounded storage. Archive paths, links, expansion, file counts, and sizes are validated. Symlinks and special entries are observed safely and are never followed during comparison. Package code, install hooks, setup.py, build backends, and lifecycle scripts are never executed.

## Research conclusions

Registry-native mechanisms already provide important pieces: npm metadata, integrity values, signatures, and provenance; PyPI release hashes, index metadata, and PEP 740 attestations; GitHub immutable commit archives; and SLSA/Sigstore standards. vltpkg/reproduce is the closest overlap for npm reproducibility because it can clone, check out, install dependencies, and run pack strategies. ReleaseCheck therefore must not claim novelty or superior reproducibility. Its defensible v0.1 focus is an explicit cross-ecosystem evidence chain and safe, non-executing comparison/reporting model.

## Repository conventions

- Read this file, ROADMAP.md, docs/DESIGN.md, `git status`, the repository tree, and relevant tests at the start of every session.
- Use ASCII by default.
- Prefer standard library and small interfaces.
- Keep live registry tests separate from deterministic offline tests.
- Update ROADMAP.md and this file at every phase checkpoint.
- Never commit or push unless explicitly requested.

## Agent and model strategy

- Antigravity with Claude Sonnet 4.6: default implementation, documentation, tests, debugging, refactoring, and routine architecture.
- Gemini 3.1 Pro inside Antigravity: broad multi-file reasoning or alternative planning when useful.
- Codex, strongest available model: independent review of security architecture, archive parsing, provenance verification, subtle supply-chain questions, large refactors, and final audit.
- Copilot: secondary implementation, documentation, or debugging support.
- Do not default to OpenCode.

## Phase status

- Phase 0: COMPLETE for the bootstrap checkpoint. Evidence: this file, ROADMAP.md, docs/COMPETITIVE_LANDSCAPE.md, docs/DESIGN.md.
- Phase 1: COMPLETE. Foundation files, CI skeleton, Git initialization, and local gofmt/test/vet/diff checks are evidenced in the repository.
- Phase 2: COMPLETE. Evidence: internal/domain/model.go, internal/domain/model_test.go, docs/DESIGN.md, and passing test/race/vet/format checks.
- Phase 3: COMPLETE. Evidence: internal/acquire/acquire.go, internal/acquire/acquire_test.go, docs/DESIGN.md, and passing test/race/vet/format checks.
- Phase 4: COMPLETE. Evidence: internal/npm/npm.go, internal/npm/npm_test.go, internal/compare, docs/DESIGN.md, and passing deterministic test/vet/format checks. Race testing is environment-limited because the installed MSYS2 GCC cannot launch `collect2.exe`.
- Phase 5: COMPLETE. Evidence: internal/pypi/pypi.go, internal/pypi/pypi_test.go, docs/DESIGN.md, and passing deterministic test/vet/format checks. Race testing remains environment-limited because the installed MSYS2 GCC cannot launch `collect2.exe`.
- Phase 6: COMPLETE. Evidence: internal/compare/compare.go, internal/compare/compare_test.go, updated archive symlink-target inventory handling, and passing serial test/race/vet/format/diff checks.
- Phase 7: COMPLETE. Evidence: internal/security/security.go, internal/security/security_test.go, the domain security-observation model, and passing serial test/race/vet/format/diff checks.
- Phase 8: COMPLETE. Evidence: internal/provenance/provenance.go, internal/provenance/provenance_test.go, docs/DESIGN.md, and passing serial test/race/vet/format/diff checks.
- Phase 9: COMPLETE. Evidence: internal/report/report.go, internal/report/report_test.go, docs/DESIGN.md, and passing serial test/race/vet/format/diff checks.
- Phase 10: COMPLETE. Evidence: fixtures/README.md, fixtures/metadata, fixtures/provenance, tests/fixture_matrix_test.go, and passing offline test/race/vet/format/diff checks.
- Phase 11: COMPLETE. Evidence: cmd/releasecheck/main.go, internal/cli/cli.go, internal/cli/cli_test.go, README usage, and passing repository test/build/smoke/vet/format/diff checks plus focused CLI/fixture race checks. Aggregate race execution is currently blocked by Windows denying access to a temporary pre-existing `compare.test.exe`.
- Phase 12: COMPLETE. Evidence: pinned cross-platform CI, race job, tag-driven release workflow, scripts/build-release.sh, VERSION, docs/RELEASE.md, and the local Phase 12 checkpoint checks.
- Phase 13: COMPLETE. Evidence: README.md, docs/USAGE.md, docs/REPORTS.md, docs/CI.md, docs/REGISTRY_ADAPTERS.md, docs/GOOD_FIRST_ISSUES.md, updated SECURITY.md/CONTRIBUTING.md, feature-request template, and the documentation checkpoint review.
- Phase 14: AUDIT REQUIRED. Evidence: FINAL_AUDIT.md and audit regression fixes. The conditional recommendation requires hosted CI execution and explicit maintainer decisions for the remaining source-binding, manifest-analysis, network-boundary, and tooling limitations.

## Definition of done

v0.1 is done only after all roadmap acceptance criteria pass, independent Codex audit findings are resolved or explicitly accepted, offline tests pass, CI passes on Windows/Linux/macOS targets, reports and exit codes are documented, and the repository contains no misleading claims or unreviewed security TODOs.

## Important commands

```text
go test ./...
go test -race ./...
go vet ./...
gofmt -w .
git status --short --branch
git diff --check
```

Commands may evolve with the implementation. Never run package managers or builds against untrusted artifacts as part of ReleaseCheck.

## Known limitations at Phase 14 audit checkpoint

- No public SDK implementation exists. Comparison, security analysis, provenance parsing, and report rendering are wired to the initial CLI; broader API compatibility remains deferred.
- Manifest analysis currently accepts explicitly supplied safe bytes; registry adapters do not yet wire package.json, setup.py, or pyproject.toml contents into the analysis pipeline.
- `pyproject.toml` is intentionally not parsed with a custom TOML implementation in this phase; empty metadata is detected, while detailed build-backend evidence remains future work.
- Provenance parsing is not cryptographic verification. npm and PyPI adapters do not yet fetch and wire registry provenance URLs into `ReleaseMetadata`; the parser is ready for later integration. Current-source research on npm and PyPI provenance was checked 2026-10-08 and recorded in docs/DESIGN.md.
- npm verification is implemented as an internal path behind the CLI, and source retrieval is currently limited to GitHub.
- PyPI verification is implemented as an internal path behind the CLI; source references are unavailable for releases whose project metadata does not provide an explicit ref.
- The fixture matrix does not contact live npm or PyPI services. Live-registry compatibility checks, if added later, must be separate, opt-in, and non-authoritative. CLI commands intentionally use the public registry defaults and currently have no mirror/base-URL flags.
- No public SDK is promised in v0.1 yet; callers should not depend on `internal/` packages as a compatibility API.
- Aggregate `go test -race -p 1 ./...` currently reaches all packages but can exit when Windows denies access to the temporary pre-existing `compare.test.exe`; the changed CLI and fixture-matrix packages pass focused race runs.
- The module path is provisional until the hosting namespace is confirmed.
- Remote, GitHub repository metadata, and maintainer identity are not configured.
- Research is a focused initial pass, not a complete literature or market survey.
- GitHub-hosted workflows have not run locally because the repository has no configured remote. The workflows are checked structurally and the equivalent local commands passed, but hosted CI status is not claimed.
- Release archives are reproducible enough for v0.1 preparation under the documented build inputs; independent cross-machine bit-for-bit reproducibility has not yet been established.
- The release workflow creates GitHub releases using the runner's preinstalled `gh` CLI. Release signing and package-registry publishing are intentionally not implemented.
- Documentation links and examples were reviewed locally; GitHub's rendered Markdown, issue-template UI, and a first-time external contributor walkthrough still require hosted or independent review.
- Final audit limitations and findings are recorded in FINAL_AUDIT.md. Do not declare v0.1 complete until its release recommendation is changed by evidence, not by deadline pressure.

The Phase 14 findings remain open unless FINAL_AUDIT.md and the repository provide new evidence. In particular: hosted CI has not run because no remote exists; source-to-commit binding is incomplete; manifest analysis is not fully wired into live verification; the SSRF/network policy needs a deliberate decision; provenance is parsed rather than fully cryptographically verified; no stable public SDK exists; local race/static-analysis validation is environment-limited; and there is no adoption or ecosystem proof.

## Future-agent operating rule

Determine the current phase from repository evidence, not from conversational claims. If asked whether a phase is complete, evaluate every acceptance criterion and verification command in ROADMAP.md, report COMPLETE or INCOMPLETE with evidence, and do not guess. If told to continue, resume the current phase from filesystem and git state.

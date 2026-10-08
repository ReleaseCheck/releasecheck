# ReleaseCheck Project Context

This file is the durable memory of the project. Repository state is authoritative. Do not trust stale chat context over the filesystem, git state, tests, or documented decisions.

## Identity

- Name: ReleaseCheck
- Planned language: Go
- Planned module: `github.com/releasecheck/releasecheck` (confirm before publishing)
- Current state: Phase 9 complete; Phase 10 fixture and test suite is next
- Deadline: 12:00 PM WAT, October 9, 2026

## Problem and product definition

ReleaseCheck will be a deterministic CLI and SDK that collects evidence about whether a package published to npm or PyPI corresponds to its claimed source repository and release. It compares downloaded distribution contents with a source snapshot when the source identity and reference can be resolved. It reports what is known, inferred, unavailable, or different.

The product is not a SaaS dashboard, chatbot, AI product, generic package scanner, SBOM generator, vulnerability scanner, repository health checker, Sigstore replacement, or SLSA replacement.

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
- Phases 10-14: NOT STARTED. Phase 10 fixture and test suite is next.

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

## Known limitations at Phase 9 checkpoint

- No CLI or SDK implementation exists. Comparison, security analysis, provenance parsing, and report rendering are internal layers and are not yet exposed through a user-facing command.
- Manifest analysis currently accepts explicitly supplied safe bytes; registry adapters do not yet wire package.json, setup.py, or pyproject.toml contents into the analysis pipeline.
- `pyproject.toml` is intentionally not parsed with a custom TOML implementation in this phase; empty metadata is detected, while detailed build-backend evidence remains future work.
- Provenance parsing is not cryptographic verification. npm and PyPI adapters do not yet fetch and wire registry provenance URLs into `ReleaseMetadata`; the parser is ready for later integration. Current-source research on npm and PyPI provenance was checked 2026-10-08 and recorded in docs/DESIGN.md.
- npm verification is implemented as an internal path only; no CLI/report wiring exists, and source retrieval is currently limited to GitHub.
- PyPI verification is implemented as an internal path only; source references are unavailable for releases whose project metadata does not provide an explicit ref. Report renderers are internal; CLI command wiring and output flags remain Phase 11 work.
- The module path is provisional until the hosting namespace is confirmed.
- Remote, GitHub repository metadata, and maintainer identity are not configured.
- Research is a focused initial pass, not a complete literature or market survey.

## Future-agent operating rule

Determine the current phase from repository evidence, not from conversational claims. If asked whether a phase is complete, evaluate every acceptance criterion and verification command in ROADMAP.md, report COMPLETE or INCOMPLETE with evidence, and do not guess. If told to continue, resume the current phase from filesystem and git state.

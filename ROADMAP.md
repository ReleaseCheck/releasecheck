# ReleaseCheck Roadmap

This is an executable plan. A phase is complete only when its acceptance criteria and definition of done are satisfied and the checkpoint records are updated. Allowed statuses: NOT STARTED, IN PROGRESS, BLOCKED, COMPLETE, AUDIT REQUIRED.

## Status

| Phase | Status | Evidence |
| --- | --- | --- |
| 0 Discovery and scope freeze | COMPLETE | PROJECT_CONTEXT.md; docs/COMPETITIVE_LANDSCAPE.md; docs/DESIGN.md |
| 1 Repository foundation | COMPLETE | governance files; go.mod; internal/project; CI skeleton; gofmt/test/vet/diff checks |
| 2 Architecture and domain model | COMPLETE | internal/domain; docs/DESIGN.md; test/race/vet/gofmt checks |
| 3 Safe artifact acquisition | NOT STARTED | - |
| 4 npm verification path | NOT STARTED | - |
| 5 PyPI verification path | NOT STARTED | - |
| 6 Comparison engine | NOT STARTED | - |
| 7 Security analysis | NOT STARTED | - |
| 8 Provenance and attestation evidence | NOT STARTED | - |
| 9 Reporting | NOT STARTED | - |
| 10 Fixture and test suite | NOT STARTED | - |
| 11 CLI and SDK polish | NOT STARTED | - |
| 12 CI and release engineering | NOT STARTED | - |
| 13 Documentation and contributor readiness | NOT STARTED | - |
| 14 Final audit | NOT STARTED | - |

At every session start, read PROJECT_CONTEXT.md, this file, docs/DESIGN.md, git status, the tree, and relevant tests. Before advancing, verify the previous phase from its acceptance criteria. At every checkpoint run the commands, inspect git diff and tree, update this file and PROJECT_CONTEXT.md, record decisions and limitations, and summarize. Do not commit or push automatically.

## Phase 0 - Discovery and scope freeze

Objective: establish the problem, boundaries, architecture direction, and honest positioning before implementation.

Why: supply-chain terminology and registry behavior are easy to overclaim.

Prerequisites: environment inspection and focused current-source research.

Implementation tasks: inspect Go/Git/tooling; research npm, PyPI, GitHub, provenance, SLSA, Sigstore, GUAC, Scorecard, and reproducibility tools; compare overlap; freeze v0.1.

Files expected: PROJECT_CONTEXT.md, docs/COMPETITIVE_LANDSCAPE.md, docs/DESIGN.md, ROADMAP.md.

Tests expected: documentation checks only; no product tests.

Security: record non-execution boundary, archive risks, trust boundaries, and limitations.

Documentation: current source links, dates, alternatives, rejected claims, decisions.

Acceptance criteria: problem, competitors, overlap, defensible differentiation, threat model, architecture direction, and v0.1 scope are explicit.

Definition of done: documents agree and Phase 1 is executable without chat history.

Verification: inspect files; `git diff --check`.

Do not: implement verification, claim novelty, fabricate adoption, or begin future ecosystems.

Agent/model: Antigravity with Claude Sonnet 4.6; Codex only for unresolved security ambiguity.

Checkpoint: update status/context with source dates, decisions, and limitations.

## Phase 1 - Repository foundation

Objective: create a clean, honest, contributor-ready Go repository skeleton.

Why: later agents need stable conventions and CI before implementation grows.

Prerequisites: Phase 0 complete.

Implementation tasks: confirm module path; create Go module, license, README, governance docs, changelog, docs/fixtures/tests, CI, issue and PR templates, and appropriate metadata.

Files expected: go.mod, LICENSE, README.md, CONTRIBUTING.md, SECURITY.md, CODE_OF_CONDUCT.md, SUPPORT.md, CHANGELOG.md, ROADMAP.md, PROJECT_CONTEXT.md, docs/, fixtures/, tests/, .github/.

Tests expected: `go test ./...`, `go vet ./...`, gofmt check, `git diff --check`.

Security: CI uses least privilege and does not download or execute untrusted fixtures.

Documentation: no feature claims; license rationale and bootstrap status are clear.

Acceptance criteria: clean foundation, no misleading claims, CI skeleton checks formatting/tests, and contribution/security channels are clear.

Definition of done: tree and checks are inspectable and Phase 2 has a home for code/tests.

Verification: `go test ./...`; `go vet ./...`; `git diff --check`.

Do not: implement the core engine, add unnecessary dependencies, build a web app, or invent repository metadata.

Agent/model: Antigravity with Claude Sonnet 4.6; Codex for licensing or CI security review if uncertain.

Checkpoint: record module/license decisions and inspect diff/tree.

## Phase 2 - Architecture and domain model

Objective: represent the evidence chain with small stable Go types and interfaces.

Why: domain contracts must precede adapters and reporting.

Prerequisites: Phase 1 complete.

Implementation tasks: define package identity, registry, artifact, source and git references, comparison, evidence, provenance, report, verdict, exit code, error taxonomy, and cache types.

Files expected: internal/domain or equivalent packages, focused tests, updated DESIGN.md.

Tests expected: validation, deterministic ordering, serialization round trips, error classification.

Security: model unknown/unavailable separately from false; prevent evidence from implying stronger trust.

Documentation: types, invariants, alternatives, rejected abstractions, versioning.

Acceptance criteria: code mirrors DESIGN.md; important behavior has unit tests; no registry-specific leakage into core types.

Definition of done: adapters can be built against the model without inventing semantics.

Verification: `go test ./...`; `go test -race ./...`; `go vet ./...`.

Do not: build a plugin framework, add speculative registries, or finalize CLI syntax prematurely.

Agent/model: Antigravity/Claude; switch to Codex for independent model/security review before Phase 3.

Checkpoint: record API decisions and compatibility risks.

## Phase 3 - Safe artifact acquisition

Objective: retrieve and inspect bounded artifacts without executing content.

Why: every later conclusion depends on safe, reproducible inputs.

Prerequisites: Phase 2 model and approved limits.

Implementation tasks: HTTPS client, redirect policy, timeouts, bounded streaming downloads, content-length checks, SHA-256, safe temp files, archive validation, cleanup, malformed input handling.

Files expected: acquisition/archive packages and offline HTTP/archive tests.

Tests expected: network failure, malformed archive, oversized input, redirects, invalid metadata, cleanup.

Security: path traversal, absolute paths, links, expansion bombs, file count, disk/memory bounds; no shell.

Documentation: limits and threat assumptions.

Acceptance criteria: safely retrieve and inspect an artifact; no code or package manager executes.

Definition of done: adversarial fixtures pass and resource limits are documented.

Verification: `go test ./...`; race; vet.

Do not: extract blindly, follow symlinks, trust filenames, or invoke npm/pip/build tools.

Agent/model: Antigravity/Claude; Codex for archive parsing/security review.

Checkpoint: record limits, rejected-input behavior, residual risks.

## Phase 4 - npm verification path

Objective: complete a deterministic npm metadata-to-comparison path.

Why: npm is the first highest-value path and has rich registry metadata.

Prerequisites: Phases 2-3 complete.

Implementation tasks: packument/version selection, tarball and integrity metadata, repository normalization, gitHead/reference resolution, source snapshot retrieval, evidence/report integration.

Files expected: npm adapter, fixtures, tests, docs updates.

Tests expected: matching, transformed artifact, missing/ambiguous repository, missing ref, bad metadata, provenance present/absent.

Security: validate metadata URLs/refs; no npm CLI or lifecycle execution.

Documentation: actual fields and registry behavior, not assumptions.

Acceptance criteria: representative npm cases are deterministic and limitations are explicit.

Definition of done: offline fixtures cover normal and ambiguous paths; live tests are separate.

Verification: offline tests, vet, race.

Do not: reproduce with npm install/pack, infer maliciousness from a diff, or treat provenance as source equivalence.

Agent/model: Antigravity/Claude; Codex for source/ref and provenance audit.

Checkpoint: update npm decisions and fixture inventory.

## Phase 5 - PyPI verification path

Objective: complete a deterministic PyPI release-file path.

Why: PyPI distributions have multiple file types and different metadata semantics.

Prerequisites: Phases 2-3 complete; comparison/report contracts reusable.

Implementation tasks: JSON/index metadata, selected sdist/wheel, hashes, core metadata/project URLs, source reference resolution, commit snapshot, distribution-aware comparison.

Files expected: PyPI adapter, sdist/wheel fixtures, tests, docs updates.

Tests expected: multiple files, hash mismatch, sdist, wheel, missing URL/ref, attestations.

Security: validate filenames and archives; never run setup.py, build backend, or pip.

Documentation: explain why wheels are not direct byte-for-byte source comparisons.

Acceptance criteria: representative PyPI cases are deterministic and honest about transformations.

Definition of done: offline tests cover file selection and comparison semantics.

Verification: offline tests, race, vet.

Do not: claim every wheel maps directly to source or compare compressed bytes as source equivalence.

Agent/model: Antigravity/Claude; Codex for packaging and attestation ambiguity.

Checkpoint: record selection policy and limitations.

## Phase 6 - Comparison engine

Objective: identify identical, source-only, artifact-only, and modified files with hashes.

Why: comparison is the deterministic core.

Prerequisites: safe archive model and source/artifact inputs.

Implementation tasks: canonical inventories, SHA-256 file hashes, path policy, normalization rules, type mismatch handling, stable ordering.

Files expected: comparison package and exhaustive fixtures/tests.

Tests expected: all categories, binary files, type changes, normalization boundaries.

Security: do not normalize away meaningful files; preserve unsafe observations.

Documentation: semantics and non-equivalence cases.

Acceptance criteria: identical input yields stable evidence and no hidden filtering.

Definition of done: fixtures demonstrate every category.

Verification: tests; race; vet.

Do not: label differences malicious or silently drop generated files.

Agent/model: Antigravity/Claude; Codex for correctness audit.

Checkpoint: record normalization and verdict impact.

## Phase 7 - Security analysis

Objective: add justified deterministic security-relevant observations.

Why: structural release risks matter even when differences are legitimate.

Prerequisites: acquisition, adapters, comparison.

Implementation tasks: install metadata/script observations, unsafe paths, symlinks, special entries, size anomalies, malformed metadata.

Files expected: analysis package, adversarial fixtures/tests, SECURITY updates.

Tests expected: malicious/malformed inputs and false-positive boundaries.

Security: analysis is not malware detection and never executes scripts.

Documentation: signal definitions and limitations.

Acceptance criteria: observations are deterministic, explainable, and separate from verdict policy.

Definition of done: security tests cover supported signals.

Verification: tests, race, vet, static analysis when available.

Do not: claim malware detection, vulnerability scanning, or malicious intent.

Agent/model: Antigravity/Claude plus mandatory Codex review.

Checkpoint: record threat coverage and blind spots.

## Phase 8 - Provenance and attestation evidence

Objective: consume supported registry provenance without recreating its infrastructure.

Why: attestations provide identity/build evidence but do not prove equality or benignness.

Prerequisites: stable evidence model and npm/PyPI paths.

Implementation tasks: detect present/absent/unavailable/invalid/insufficient evidence; bind to artifact digest; record identity/source/build fields where supported.

Files expected: provenance package, fixtures/tests, docs updates.

Tests expected: all evidence availability and validity variations.

Security: trust roots and cryptographic boundaries explicit; no custom Sigstore/SLSA.

Documentation: what each mechanism establishes and does not establish.

Acceptance criteria: provenance never upgrades ambiguity to trust solely because an attestation exists.

Definition of done: deterministic evidence is stable and limitations visible.

Verification: tests, vet, race, independent review.

Do not: recreate Sigstore, SLSA, npm, or PyPI infrastructure.

Agent/model: Antigravity/Claude; strongest available Codex for verification semantics.

Checkpoint: record trust-root and dependency decisions.

## Phase 9 - Reporting

Objective: produce stable human, JSON, and SARIF output.

Why: evidence matters only when humans and CI can interpret it consistently.

Prerequisites: evidence/verdict models stable.

Implementation tasks: schema version, stable ordering, human summary, JSON, SARIF, verdict mapping, exit codes.

Files expected: report packages, schemas/docs, golden tests.

Tests expected: golden outputs, repeated-run stability, invalid report inputs.

Security: avoid secrets/control-sequence injection; preserve fact/observation/warning/limitation distinctions.

Documentation: schema fields, verdicts, exit codes.

Acceptance criteria: same input produces stable machine-readable evidence.

Definition of done: consumers can rely on versioned output contracts.

Verification: tests; vet; race; diff checks.

Do not: hide uncertainty behind a single score or AI summary.

Agent/model: Antigravity/Claude; Codex for schema/exit-code review.

Checkpoint: record compatibility policy.

## Phase 10 - Fixture and test suite

Objective: make core behavior reproducible offline.

Why: live registries change and cannot be the test oracle.

Prerequisites: implemented paths and report contracts.

Implementation tasks: fixtures for matching, extra/missing/modified files, malformed package, suspicious metadata, missing repo, missing tag, unsafe path, symlink, size boundary, provenance variations.

Files expected: fixtures/, tests/, golden reports.

Tests expected: all twelve required scenarios; live tests separate and opt-in.

Security: fixtures inert and never executed.

Documentation: fixture provenance and test policy.

Acceptance criteria: offline suite exercises core behavior and live tests are clearly marked.

Definition of done: deterministic tests pass on supported OSes.

Verification: `go test ./...`; race; vet; repeated runs.

Do not: rely on current public packages for core correctness.

Agent/model: Antigravity/Claude; Codex for adversarial test review.

Checkpoint: record coverage and gaps.

## Phase 11 - CLI and SDK polish

Objective: finalize ergonomic commands, flags, help, errors, exit codes, and justified public API.

Why: utility depends on a small, understandable interface.

Prerequisites: reporting and tests stable.

Implementation tasks: research CLI conventions; finalize commands/flags; document local files and CI; expose SDK only if API maturity warrants it.

Files expected: cmd/, internal wiring, public package if justified, CLI tests, README updates.

Tests expected: command acceptance, JSON/SARIF, invalid input, exit status.

Security: no shell execution or package manager invocation.

Documentation: install, usage, CI, limitations.

Acceptance criteria: a developer can use the CLI from README without chat assistance.

Definition of done: interface stable enough for v0.1.

Verification: CLI tests and all checks.

Do not: create unnecessary public SDK surface or a dashboard.

Agent/model: Antigravity/Claude; Codex for command handling review.

Checkpoint: record compatibility and UX decisions.

## Phase 12 - CI and release engineering

Objective: establish reliable checks and reproducible-enough v0.1 artifacts.

Why: security tooling must be testable and distributable across targets.

Prerequisites: CLI and suite stable.

Implementation tasks: gofmt, vet, tests, race, static analysis, Windows/Linux/macOS builds, GitHub Actions, release workflow, semver, v0.1.0 preparation.

Files expected: workflows, build/release configuration, version docs.

Tests expected: CI matrix and artifact checks.

Security: pin actions, least privilege, no secret leakage, review release permissions.

Documentation: build and release process.

Acceptance criteria: CI passes and release artifacts are reproducible enough for v0.1.

Definition of done: maintainer can cut a release from documented steps.

Verification: local checks and CI.

Do not: claim bit-for-bit reproducibility without measuring it.

Agent/model: Antigravity/Claude; Codex for workflow/release audit.

Checkpoint: record tool availability and exceptions.

## Phase 13 - Documentation and contributor readiness

Objective: make the project independently understandable and contributable.

Why: durable open-source utility depends on clear boundaries and contribution paths.

Prerequisites: feature and release behavior stable.

Implementation tasks: finalize README, architecture, security, limitations, examples, CI, JSON/SARIF docs, issue templates, PR template, good-first-issue candidates, adapter guide.

Files expected: docs and .github updates.

Tests expected: link checks where available; examples kept current.

Security: docs must not suggest unsafe commands or guarantees.

Documentation: what it verifies, does not verify, differs, and how to contribute.

Acceptance criteria: a competent developer can understand and contribute without author contact.

Definition of done: docs are accurate against code and current research.

Verification: doc review, tests, CI.

Do not: add marketing claims or fabricated validation.

Agent/model: Antigravity/Claude; Codex for final security wording.

Checkpoint: record doc gaps and reader review.

## Phase 14 - Final audit

Objective: independently audit the v0.1 release candidate.

Why: independent review is required before declaring security-sensitive tooling complete.

Prerequisites: all prior phases complete; release candidate available.

Implementation tasks: audit architecture, security, archives, source resolution, provenance, comparison, CLI, exit codes, schemas, tests, docs, CI, artifacts, license, dead code, TODOs, and claims.

Files expected: FINAL_AUDIT.md and addressed fixes.

Tests expected: full suite, race, vet, static analysis, cross-platform CI.

Security: findings require severity, evidence, remediation, and residual risk.

Documentation: strengths, findings, fixes, unresolved limitations, release recommendation.

Acceptance criteria: audit recommends release or explains why not; blocking findings are resolved.

Definition of done: only then mark v0.1 complete.

Verification: all checks and independent Codex review.

Do not: waive findings because the deadline is near; never cut core verification, security, tests, docs, CI, or reproducibility.

Agent/model: strongest available Codex; Antigravity/Claude implements fixes.

Checkpoint: update roadmap/context and create FINAL_AUDIT.md.

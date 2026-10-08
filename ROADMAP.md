# ReleaseCheck Roadmap

This is an executable plan. A phase is complete only when its acceptance criteria and definition of done are satisfied and the checkpoint records are updated. Allowed statuses: NOT STARTED, IN PROGRESS, BLOCKED, AUDIT REQUIRED, COMPLETE, DEFERRED, CANCELLED.

## Current Version State

| Item | Current truth |
| --- | --- |
| Product version | Pre-v0.1.0; no public release has been made |
| Current milestone | Foundation complete; post-Phase-14 audit; v0.1 release gate not passed |
| Current active phase | Phase 14 - Final audit (`AUDIT REQUIRED`) |
| Next executable phase | Phase 15 - Audit remediation (`NOT STARTED`) |
| Release decision | Conditional; see FINAL_AUDIT.md |
| Repository evidence | Audit checkpoint is committed; current working-tree changes must always be checked with `git status` and `git diff` |

## Canonical Phase Register

This is the single status overview for the whole roadmap. The detailed sections below define the work; this register defines the current status and evidence position. Future-phase evidence entries identify the governing plan, not completed implementation.

| Version / band | Phase | Status | Evidence / current truth | Next gate |
| --- | ---: | --- | --- | --- |
| Foundation | 0 | COMPLETE | Discovery docs, competitive landscape, design, and context | Historical |
| Foundation | 1 | COMPLETE | Governance, Go module, CI skeleton, repository foundation | Historical |
| Foundation | 2 | COMPLETE | `internal/domain`, design contracts, domain tests | Historical |
| Foundation | 3 | COMPLETE | `internal/acquire`, bounded download/archive tests | Historical |
| Foundation | 4 | COMPLETE | `internal/npm`, npm fixtures, comparison baseline | Historical |
| Foundation | 5 | COMPLETE | `internal/pypi`, sdist/wheel fixtures, hash/source tests | Historical |
| Foundation | 6 | COMPLETE | `internal/compare`, path/hash/symlink comparison tests | Historical |
| Foundation | 7 | COMPLETE | `internal/security`, adversarial structural observations | Historical |
| Foundation | 8 | COMPLETE | `internal/provenance`, artifact-binding fixtures/tests | Historical |
| Foundation | 9 | COMPLETE | `internal/report`, JSON/human/SARIF and verdict tests | Historical |
| Foundation | 10 | COMPLETE | Offline fixture matrix and twelve required scenarios | Historical |
| Foundation | 11 | COMPLETE | CLI, CLI tests, build/smoke checks, documented SDK deferral | Historical |
| Foundation | 12 | COMPLETE | Pinned CI/release workflows, build script, local release checks | Hosted validation still required by v0.1 gate |
| Foundation | 13 | COMPLETE | README, contributor docs, report/CI/adapter guides, templates | Historical |
| Foundation | 14 | AUDIT REQUIRED | `FINAL_AUDIT.md`; fixes applied; residual findings remain | Phase 15 remediation and evidence |
| Organization | M0 | COMPLETE | Three public repositories created, pushed, and verified; core `1ff03f9`, Action `2fcc5ec`, docs `5bf6e08` | Maintain one-way ownership boundaries |
| v0.1.0 candidate | 15 | NOT STARTED | Audit-remediation plan in Phase 15 section | Classify or resolve F-001 to F-005 |
| v0.1.0 candidate | 16 | NOT STARTED | Independent re-audit plan in Phase 16 section | Phase 15 evidence |
| v0.1.0 candidate | 17 | NOT STARTED | Hosted CI/release-validation plan in Phase 17 section | Public remote and Phase 16 review |
| v0.1.0 candidate | 18 | NOT STARTED | Release-preparation plan in Phase 18 section | Hosted validation and go decision |
| v0.1.0 | 19 | NOT STARTED | Release/stabilization plan in Phase 19 section | Actual release and stabilization evidence |
| v1.0 | 20 | NOT STARTED | v1 discovery plan; scope must be evidence-driven | Real v0.1 stabilization evidence |
| v1.0 | 21 | DEFERRED | Public SDK candidate; no stable API promise today | Phase 20 proves API need |
| v1.0 | 22 | NOT STARTED | Stronger source identity/commit-binding candidate | v1 scope freeze |
| v1.0 | 23 | NOT STARTED | Cryptographic provenance-verification candidate | v1 scope and standards review |
| v1.0 | 24 | NOT STARTED | Hardened network/SSRF policy candidate | v1 scope and threat model |
| v1.0 | 25 | NOT STARTED | GitHub/CI integration candidate | Demonstrated workflow need |
| v1.0 | 26 | NOT STARTED | Additional source-host candidate | Demonstrated host need |
| v1.0 | 27 | NOT STARTED | Advanced package/release semantics candidate | Real transformation evidence |
| v1.0 | 28 | NOT STARTED | Contributor/ecosystem scale candidate | Measured collaboration needs |
| v1.0 | 29 | NOT STARTED | v1 audit/release gate | All v1 acceptance evidence |
| v2.0 | 30 | DEFERRED | v2 discovery; exact scope intentionally unknown | Real v1 usage and incidents |
| v2.0 | 31 | DEFERRED | Additional registries are candidates only | Discovery and demand evidence |
| v2.0 | 32 | DEFERRED | Broader provenance interoperability candidate | Demonstrated standards need |
| v2.0 | 33 | DEFERRED | Evidence-policy engine candidate | Stable evidence contracts |
| v2.0 | 34 | DEFERRED | Persistent/cache analysis candidate | Privacy and retention decision |
| v2.0 | 35 | DEFERRED | Large-scale CI integration candidate | Real organization workflows |
| v2.0 | 36 | DEFERRED | Performance/scale candidate | Measurements, not assumptions |
| v2.0 | 37 | DEFERRED | Advanced investigation candidate | Evidence-authority review |
| v2.0 | 38 | DEFERRED | Ecosystem/integration expansion candidate | Demonstrated integration need |
| v2.0 | 39 | NOT STARTED | v2 audit/release gate | All v2 acceptance evidence |
| v3.0 reserved | 40-49 | DEFERRED | Reserved discovery-gated planning band | v3 discovery |
| v4.0 reserved | 50-59 | DEFERRED | Reserved discovery-gated planning band | v4 discovery |
| v5.0+ reserved | Future | DEFERRED | No fixed feature promises; evidence-led discovery required | Future discovery |

## Current Repository Architecture

The initial organization architecture is established as three repositories. The migration milestone was completed on 2026-10-08 after all three public repositories existed, contained the intended history/content, and local `main` matched remote `main`.

```text
CURRENT
├── ReleaseCheck/releasecheck         core implementation and technical truth
├── ReleaseCheck/releasecheck-action  GitHub Actions integration
└── ReleaseCheck/releasecheck-docs    public user documentation

FUTURE - NOT CREATED
├── candidate repository 4            planned/discovery-gated
├── candidate repository 5            planned/discovery-gated
└── candidate repository 6            planned/discovery-gated
```

The rule is simple: create a repository only when it has a genuinely independent responsibility. The core is not split into npm and PyPI repositories. Future candidates may include RFC/design proposals, ecosystem integrations, or community/experimental tooling, but no names or commitments are assigned until discovery establishes a need.

### Organization milestone M0 - initial three-repository architecture

Status: COMPLETE.

Objective: establish the initial `ReleaseCheck` organization layout without duplicating the verification engine or weakening the core repository as the technical source of truth.

Acceptance criteria:

- `ReleaseCheck/releasecheck` is the core repository and retains its existing history.
- `ReleaseCheck/releasecheck-action` contains only the GitHub Actions integration and its tests/docs.
- `ReleaseCheck/releasecheck-docs` contains only curated user documentation and its contribution/security metadata.
- all three repositories are public, have accurate descriptions, and default to `main`.
- each local repository has a clean working tree after commit and its local `main` is verified against remote `main`.
- no future repository is created and no v0.1 implementation scope is expanded.

Evidence: `https://github.com/ReleaseCheck/releasecheck` at `1ff03f92cb394103e793e84b0fa7d90faf2a2830`, `https://github.com/ReleaseCheck/releasecheck-action` at `2fcc5ec21b4b180af4583273f9535e2c835b94b5`, and `https://github.com/ReleaseCheck/releasecheck-docs` at `5bf6e08195e1904e9a214f9dc3c38078a7c8bc53`. All are public with default branch `main`; local and remote commit IDs matched at checkpoint.

At every session start, read PROJECT_CONTEXT.md, this file, docs/DESIGN.md, git status, the tree, and relevant tests. Before advancing, verify the previous phase from its acceptance criteria. At every checkpoint run the commands, inspect git diff and tree, update this file and PROJECT_CONTEXT.md, record decisions and limitations, and summarize. Do not commit or push automatically.

## Versioned Product Roadmap

Phases 0-14 are the historical foundation and current audit state. Future phases are organized around release milestones:

```text
FOUNDATION / INITIAL BUILD       Phases 0-14
              |
              v
v0.1.0 RELEASE CANDIDATE         Phases 15-19
              |
              v
v1.0 PRODUCT                    Phases 20-29
              |
              v
v2.0 PLATFORM EXPANSION         Phases 30-39
              |
              v
v3+ FUTURE EVOLUTION            Phase bands reserved and discovery-gated
```

The relationship is: `phase -> capability -> acceptance criteria -> evidence -> milestone -> release`. A roadmap item may be deferred or cancelled when evidence shows that it does not strengthen the central product problem. Roadmap size is not a success metric.

## Current Reality vs Future

| Area | Current reality | v0.1 | v1 | v2+ |
| --- | --- | --- | --- | --- |
| npm | Implemented internal path and CLI command | Stabilize and validate | Deeper release semantics | Broader ecosystems |
| PyPI | Implemented internal path and CLI command | Stabilize and validate | Deeper release semantics | Broader ecosystems |
| Source/artifact comparison | Implemented deterministic inventory/hash comparison | Harden known limitations | Explain legitimate transformations better | Broader analysis |
| Provenance | Evidence parsing and artifact binding | Honest limitations | Stronger verification where justified | Broader interoperability |
| SDK | Internal packages only | Deferred unless API is mature | Stable public Go SDK | Integrations |
| CI | Workflows exist; hosted proof pending | Validate hosted workflows | First-class CI integration | Scale usage |
| Source hosts | Limited supported GitHub path | Document limits | Expand behind clean abstraction | Evaluate further hosts |
| Web/GUI | Not implemented and not core | No | Evaluate only with evidence | Only if justified |

## Release Gate Model

Every major release must pass the same gate:

```text
Scope frozen
    -> implementation complete
    -> tests passing
    -> security review complete
    -> documentation complete
    -> cross-platform validation complete
    -> release artifacts validated
    -> known limitations documented
    -> explicit release decision
```

No phase or release may be marked complete merely because files exist or code compiles. A release decision must name unresolved risks and the evidence supporting acceptance.

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

Checkpoint: COMPLETE on 2026-10-08. Added a pinned cross-platform CI matrix, Ubuntu race job, tag-driven release workflow, reproducible-enough build script, version metadata, and release documentation. Local tests, vet, formatting, whitespace, CLI build, cross-target build, and repeated-build hash checks passed. After the repository migration, hosted CI ran: Linux, macOS, and race passed; the Windows formatting job failed and remains an open release-engineering issue.

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

Checkpoint: COMPLETE on 2026-10-08. Updated the README with installation, verification scope, limitations, CI, and existing-tool boundaries; added usage, report, CI, registry-adapter, and scoped good-first-issue guidance; added a feature-request template; corrected stale security and CLI wording; and completed link, command, formatting, test, vet, and whitespace review. GitHub-hosted rendering and first-time contributor review remain external verification tasks, not invented as completed.

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

Checkpoint: AUDIT REQUIRED on 2026-10-08. FINAL_AUDIT.md records verified strengths, four material residual findings, one environment limitation, fixes made during review, regression coverage, and a conditional release recommendation. v0.1 is not declared complete until hosted CI runs and the remaining findings are explicitly accepted or remediated.

## v0.1.0 Release Candidate - Phases 15-19

These phases are the executable path from the current audited repository to a defensible first public release. They are not started by this documentation pass.

### Phase 15 - Audit remediation

Milestone: v0.1.0 release candidate preparation

Status: NOT STARTED

Objective: close, mitigate, or explicitly accept every finding in FINAL_AUDIT.md.

Why: the current audit recommendation is conditional and v0.1 must not be declared complete by documentation alone.

Prerequisites: Phase 14 audit report; maintainer decision on whether each residual risk is acceptable for a local CLI.

Implementation scope:

- hosted CI validation workstream and remote setup;
- source-to-commit binding design and implementation, or an explicit v0.1 limitation decision;
- safe manifest-byte extraction and npm/Python security-observation integration, or an explicit scope reduction;
- SSRF and network policy decision covering redirects, DNS, private ranges, proxies, and private mirrors;
- provenance verification boundary decision and any justified cryptographic integration;
- practical local security tooling selection;
- any new findings discovered while remediating the audit.

Likely files/areas: internal/acquire, internal/npm, internal/pypi, internal/security, internal/provenance, internal/cli, fixtures, tests, SECURITY.md, docs/DESIGN.md, FINAL_AUDIT.md, and CI configuration.

Tests: adversarial URL and redirect tests; source-binding tests; manifest extraction and lifecycle metadata tests; provenance trust-boundary tests; regression tests for every fixed audit finding; offline fixtures for every changed behavior.

Security: preserve non-execution, bounded resources, safe archives, and honest evidence states. Do not solve SSRF by silently breaking documented private-mirror use cases; define the policy first.

Documentation: update the audit finding status, threat model, limitations, report semantics, and contributor guidance for every accepted or resolved finding.

Acceptance criteria: every F-001 through F-005 is `RESOLVED`, `ACCEPTED`, or replaced by a more precise finding with an owner and evidence plan; no contradictory claim remains.

Definition of done: remediation tests pass, security boundaries are reviewed, and FINAL_AUDIT.md contains evidence for each status.

Evidence required: test output, hosted CI runs where applicable, design decisions, changed-file references, and an updated independent-review request.

Do not: add future ecosystems, execute package builds, weaken source identity semantics, or label parsed provenance as cryptographically verified.

Agent/model: Antigravity + Claude Sonnet 4.6 for implementation; Gemini 3.1 Pro for alternative network/source-binding exploration; Codex for every security-boundary decision and remediation review.

Switch to Codex when: changing URL resolution, archive handling, provenance verification, source binding, or verdict semantics.

Checkpoint: update FINAL_AUDIT.md, ROADMAP.md, PROJECT_CONTEXT.md, tests, and limitations; inspect tree and diff; do not advance on unclassified findings.

### Phase 16 - Independent re-audit

Milestone: v0.1.0 release candidate approval

Status: NOT STARTED

Objective: have an independent reviewer re-evaluate all Phase 14 findings and remediation.

Why: the implementer should not be the sole authority for security-boundary changes.

Prerequisites: Phase 15 evidence and a clean checkpoint.

Scope: review every previous finding, changed security path, report behavior, documentation claim, and residual limitation.

Tests: full offline suite, focused adversarial tests, race/static-analysis evidence where available, and changed-path review.

Security: verify that remediation did not introduce execution, path traversal, unbounded work, trust confusion, or unsafe network behavior.

Documentation: produce an independent audit addendum or replacement audit with severity, evidence, fixes, accepted risks, and recommendation.

Acceptance criteria: every previous finding is classified; resolved findings have repository evidence; accepted findings are explicit; documentation is internally consistent.

Definition of done: the independent review recommends proceeding to hosted validation or explains the blocking conditions.

Evidence required: independent review file, test results, and a clean Git checkpoint.

Do not: treat an unreviewed code diff or passing unit test as independent review.

Agent/model: strongest available Codex; Antigravity/Claude implements only requested follow-up fixes.

Switch to Codex when: the remediation touches trust decisions or when the first review and implementation disagree.

Checkpoint: update status only after the review is written and all blocking findings are addressed.

### Phase 17 - Hosted CI and release validation

Milestone: v0.1.0 release candidate validation

Status: NOT STARTED

Objective: prove the repository works in the actual hosted environment.

Why: local Windows checks cannot substitute for the committed Linux, Windows, and macOS workflows.

Prerequisites: intended public remote; Phase 16 recommendation to proceed.

Scope: configure remote; run matrix CI; run Ubuntu race job; run supported static analysis; validate release workflow permissions, tag matching, builds, checksums, and artifact contents.

Tests: actual GitHub Actions runs on Linux, Windows, and macOS; release workflow dry run or reviewed test tag; downloaded artifact smoke tests; checksum verification; report/CLI smoke tests.

Security: confirm least-privilege permissions, immutable action pins, no secret leakage, no untrusted package execution, and correct release write scope.

Documentation: record workflow URLs/results, supported runner/tool versions, release validation procedure, and any hosted-only exceptions.

Acceptance criteria: CI passes on all declared runners; race/static-analysis jobs pass or have documented accepted limitations; release artifacts build and validate in the hosted environment.

Definition of done: hosted evidence is linked from the repository and no claim says merely that workflows exist.

Evidence required: CI run identifiers, artifact hashes, smoke-test output, and updated FINAL_AUDIT.md.

Do not: push package code to live registries, claim release success from a YAML inspection, or add a dashboard to observe CI.

Agent/model: Antigravity + Claude Sonnet 4.6 for workflow fixes; Codex for permissions and release-security review.

Switch to Codex when: release permissions, action pins, signing, or artifact provenance changes.

Checkpoint: record actual hosted results and inspect the repository tree/diff before proceeding.

### Phase 18 - v0.1.0 release preparation

Milestone: v0.1.0 public release candidate

Status: NOT STARTED

Objective: prepare a truthful, installable, documented first release.

Why: release quality includes usage clarity, support expectations, and known limitations, not only binaries.

Prerequisites: Phase 17 hosted validation and accepted audit recommendation.

Scope: version and changelog; release notes; supported platforms; installation instructions; security disclosure path; examples; sample reports; release artifact validation; tag procedure.

Tests: clean-tree release rehearsal; CLI `--version` check; archive extraction and binary smoke tests on declared targets; JSON/SARIF sample validation; checksum verification.

Security: do not claim signatures, provenance verification, or source equivalence that the release does not provide. Review release permissions and accidental debug/test artifacts.

Documentation: update README, docs/RELEASE.md, docs/USAGE.md, docs/REPORTS.md, SECURITY.md, SUPPORT.md, and CHANGELOG.md.

Acceptance criteria: a new user can install the exact release, run a documented command, understand the result, report a problem privately, and see all known limitations.

Definition of done: the release candidate is reproducible enough under the documented process and has an explicit go/no-go decision.

Evidence required: release checklist, artifact hashes, sample output, hosted validation, and maintainer review.

Do not: fabricate release metrics, announce adoption, or publish v0.1.0 before the gate passes.

Agent/model: Antigravity + Claude Sonnet 4.6; Codex for final release/security wording.

Switch to Codex when: release artifacts, version metadata, or security claims change.

Checkpoint: record the candidate tag decision without pushing automatically.

### Phase 19 - v0.1.0 release and stabilization

Milestone: v0.1.0

Status: NOT STARTED

Objective: publish the first release only after the preceding gates pass, then capture real stabilization evidence.

Why: the first public release creates facts that cannot be safely invented beforehand.

Prerequisites: Phase 18 go decision.

Scope: actual tag/release; issue triage; reproducibility observations; bugs; contributor feedback; documentation friction; CI issues; unexpected registry behavior; capability requests.

Tests: post-release installation and CLI smoke tests from published artifacts; report-schema checks; regression fixtures for every confirmed bug.

Security: monitor disclosures through SECURITY.md; do not treat early usage as security validation; preserve a clear distinction between reports and trust guarantees.

Documentation: publish release notes, stabilization checkpoint, known issues, and changes to the roadmap based on evidence.

Acceptance criteria: v0.1.0 is actually published; post-release checks are recorded; stabilization issues are triaged; no false adoption or security claims are added.

Definition of done: the stabilization checkpoint identifies what is working, what is not, and whether v1 discovery should begin.

Evidence required: published release URL, artifact hashes, post-release smoke output, issue/feedback summary, and updated CHANGELOG.md.

Do not: treat release publication as proof of maturity or use unverified user anecdotes as quantitative validation.

Agent/model: Antigravity + Claude Sonnet 4.6 for maintenance; Codex for security regressions and release incidents.

Switch to Codex when: a report changes trust semantics or a package/registry security incident is investigated.

Checkpoint: create a dated stabilization record before marking Phase 19 complete.

## v1.0 Product - Phases 20-29

v1 phases are candidate product work after real v0.1 evidence. Phase 20 must freeze the actual v1 scope; later candidate phases may be deferred or cancelled.

### Phase 20 - v1 discovery and evidence review

Status: NOT STARTED. Objective: use real v0.1 usage, bugs, contributor feedback, source-mapping failures, package edge cases, and CI needs to freeze v1 scope. Acceptance requires a dated scope decision, evidence summary, and rejected-feature list. Do not infer demand from roadmap votes or fabricate usage. Agent: Antigravity/Claude for synthesis; Codex for security-incident interpretation. Evidence: stabilization checkpoint and scope-freeze document.

### Phase 21 - Stable public Go SDK

Status: DEFERRED until Phase 20 demonstrates a stable API need. Objective: expose deliberately designed public packages, compatibility guarantees, examples, and API policy. Tests require public API examples and compatibility checks. Security requires no accidental exposure of unsafe internals. Do not export `internal/` packages merely to call the project an SDK. Agent: Antigravity/Claude; Codex for API/security review. Evidence: API design, examples, compatibility policy, tests.

### Phase 22 - Stronger source identity and commit binding

Status: NOT STARTED. Objective: strengthen the relationship between claimed repository, exact commit/ref, retrieved source, and reported evidence. Acceptance requires tested binding semantics and explicit proof limits. Do not call a mutable tag immutable. Agent: Antigravity/Claude; Codex required. Evidence: design decision, adversarial fixtures, report examples, independent review.

### Phase 23 - Full provenance verification layer

Status: NOT STARTED. Objective: evaluate and implement justified cryptographic verification for attestations, DSSE, Sigstore-related evidence, transparency logs, and registry trust roots. Acceptance requires supported standards, trust roots, failure semantics, and fixtures. Do not recreate or replace Sigstore/SLSA infrastructure. Agent: Antigravity/Claude for integration; Codex required for cryptographic review. Evidence: verified test vectors, trust policy, documentation, independent audit.

### Phase 24 - Hardened network trust model

Status: NOT STARTED. Objective: define network policy for SSRF, redirects, DNS/rebinding, private ranges, proxies, mirrors, timeouts, and failures. Acceptance requires a documented policy and adversarial tests. Do not silently block legitimate private mirrors or claim isolation without a tested transport boundary. Agent: Antigravity/Claude; Codex required. Evidence: threat model, transport tests, policy documentation.

### Phase 25 - GitHub and CI integration

Status: NOT STARTED. Objective: add first-class automation only where it strengthens evidence review, such as an Action, annotations, SARIF workflows, or policy failure modes. Acceptance requires a minimal supported integration, security review, and examples. Do not build a dashboard-first SaaS or generic GitHub bot. Agent: Antigravity/Claude; Codex for permissions. Evidence: hosted integration tests and documentation.

### Phase 26 - Source-host expansion

Status: NOT STARTED. Objective: reduce unnecessary dependence on one source host behind a clean abstraction. Acceptance requires at least one justified host, offline fixtures, consistent identity semantics, and no provider-specific leakage into domain code. Do not add hosts without evidence of need. Agent: Antigravity/Claude; Codex for URL/trust review. Evidence: adapter tests, design update, contributor guide.

### Phase 27 - Advanced package and release semantics

Status: NOT STARTED. Objective: explain legitimate generated, compiled, bundled, ignored, wheel, documentation, and source-map differences more accurately. Acceptance requires documented normalization/provenance semantics and fixtures proving meaningful evidence is not erased. Do not normalize away suspicious files or make source/artifact equality the only valid outcome. Agent: Antigravity/Claude; Codex for comparison/security review. Evidence: semantic fixtures and report examples.

### Phase 28 - Ecosystem and contributor scale

Status: NOT STARTED. Objective: improve adapter development, fixture conventions, issue triage, compatibility guarantees, performance, and maintainer workflows only as real usage demands. Acceptance requires measured pain points and contributor-tested workflows. Do not optimize for imaginary scale. Agent: Antigravity/Claude; Codex for security-sensitive contributor tooling. Evidence: contributor feedback, benchmarks where relevant, updated guides.

### Phase 29 - v1.0 audit and release

Status: NOT STARTED. Objective: audit and release a mature core product. Acceptance requires all v1 criteria, security and architecture review, cross-platform validation, reproducibility review, documentation, artifact validation, and explicit limitations. Do not mark complete because code compiles. Agent: strongest Codex for audit; Antigravity/Claude for fixes. Evidence: FINAL_AUDIT.md replacement/addendum, hosted CI, release artifacts, release recommendation.

## v2.0 Platform Expansion - Phases 30-39

This band is **PLANNED / DISCOVERY-GATED**. The exact v2 scope is intentionally not fixed.

### Phase 30 - v2 discovery

Status: DEFERRED until v1 usage exists. Reassess real users, contributors, security incidents, edge cases, performance, and CI patterns, then freeze v2 priorities. Evidence: discovery report and accepted scope. Do not infer demand without evidence.

### Phase 31 - Additional registries

Status: DEFERRED / DISCOVERY-GATED. Candidates include crates.io, Go modules, RubyGems, NuGet, and others. Each candidate requires evidence of need, adapter fit, fixture strategy, and security review. Do not implement future ecosystems in v0.1.

### Phase 32 - Broader provenance interoperability

Status: DEFERRED / DISCOVERY-GATED. Expand standards interoperability only where demonstrated demand and verified trust semantics justify it. Do not claim universal provenance support.

### Phase 33 - Advanced evidence policy engine

Status: DEFERRED / DISCOVERY-GATED. Explore organization-defined policies over deterministic evidence without becoming generic compliance SaaS. Acceptance requires stable evidence contracts and clear policy semantics.

### Phase 34 - Persistent and cached analysis

Status: DEFERRED / DISCOVERY-GATED. Investigate content-addressed caching, reproducible evidence storage, and offline re-analysis. Do not automatically create a hosted service or retain sensitive package data.

### Phase 35 - Large-scale CI integration

Status: DEFERRED / DISCOVERY-GATED. Address organization-wide patterns only if real workflows require them. Security, permissions, tenancy, and data-retention consequences require separate review.

### Phase 36 - Performance and scale

Status: DEFERRED / DISCOVERY-GATED. Measure large packages, high-volume verification, concurrency, caching, memory, and repeatability before optimizing. Do not add concurrency merely for appearance.

### Phase 37 - Advanced investigation workflows

Status: DEFERRED / DISCOVERY-GATED. Explore deeper deterministic investigation while preserving evidence as authority. AI, if ever used, may summarize evidence but may not decide correctness or trustworthiness.

### Phase 38 - Ecosystem and integration expansion

Status: DEFERRED / DISCOVERY-GATED. Evaluate integrations with release, security, and developer tooling only when they reinforce the central problem. Do not turn ReleaseCheck into a generic platform.

### Phase 39 - v2 audit and release

Status: NOT STARTED. Require the same audit-first release gate as v1: scope, implementation, tests, security review, documentation, hosted validation, artifacts, limitations, and explicit recommendation.

## v3+ Future Horizon

Future major versions use reserved planning bands rather than fixed feature promises:

```text
v3.0  Phases 40-49
v4.0  Phases 50-59
v5.0+ Future reserved bands
```

Each band begins with discovery. Features may be added, removed, deferred, or cancelled. Every proposal must answer: does it improve deterministic understanding of the relationship between a published artifact, its claimed source, its release identity, and associated evidence? If not, it belongs out of scope or in a separate project. Future milestones are not adoption targets and roadmap length is not success.

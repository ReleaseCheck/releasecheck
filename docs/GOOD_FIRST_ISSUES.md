# Contributor Issue Catalog

These are ready-to-file issue specifications, not claims that external GitHub issues already exist. The maintainer should create only the items that fit current capacity and the active roadmap.

## 1. Hosted release rehearsal evidence

- Problem: The tag-triggered release path has not yet produced hosted artifact evidence.
- Scope: Run and document `.github/workflows/release-rehearsal.yml`; do not create a tag or GitHub Release.
- Files: `.github/workflows/release-rehearsal.yml`, `scripts/rehearse-release.sh`, `docs/RELEASE.md`.
- Acceptance: five archives, SHA256SUMS, archive contents, Linux `--version`, and no release state are recorded.
- Tests/output: successful workflow URL and artifact/checksum log.
- Security: workflow has `contents: read` and never publishes.
- Difficulty: small; dependency: none.

## 2. Action release fixture E2E

- Problem: Shell validation covers checksum parsing, but a hosted Action run against a public core release is still unavailable before v0.1.
- Scope: After the first authorized core release, add a workflow that invokes the Action by immutable commit and explicit core version.
- Files: `releasecheck-action/.github/workflows/e2e.yml`, Action test fixtures, core release docs.
- Acceptance: MATCH, REVIEW, invalid checksum, missing archive, and nonzero verifier exit are asserted.
- Tests/output: workflow summaries and report artifacts with expected exit handling.
- Security: no mutable branch or silent version fallback.
- Difficulty: medium; dependency: authorized public core release.

## 3. Archive adversarial matrix

- Problem: traversal, duplicate-entry, symlink, and oversized archive cases should be easier to extend.
- Scope: Add table-driven ZIP/TAR fixtures with exact error categories and no extraction.
- Files: `tests/fixture_matrix_test.go`, `internal/acquire`, `internal/security`.
- Acceptance: every unsafe case fails closed with a stable category.
- Tests/output: offline `go test` assertions.
- Security: preserve bounded reads and path validation.
- Difficulty: medium; dependency: none.

## 4. Provenance evidence fixtures

- Problem: Registry provenance is parsed structurally but the fixture set is still small.
- Scope: Add absent, unavailable, malformed, artifact-mismatch, and insufficient npm/PyPI evidence.
- Files: `fixtures/provenance`, `internal/provenance/*_test.go`, `docs/REPORTS.md`.
- Acceptance: status, explanation, binding fields, and verdict effects are exact and deterministic.
- Tests/output: offline JSON and SARIF snapshots.
- Security: do not imply cryptographic trust verification.
- Difficulty: medium; dependency: current report contract.

## 5. Offline verification mode

- Problem: CI users may need to analyze saved metadata and archives without registry access.
- Scope: Design and implement an explicit offline input contract; do not reinterpret live URLs implicitly.
- Files: `internal/cli`, `internal/domain`, `docs/USAGE.md`, fixtures.
- Acceptance: missing inputs fail clearly; equivalent inputs produce identical JSON.
- Tests/output: offline CLI smoke and exit-code tests.
- Security: no arbitrary local path or archive execution.
- Difficulty: large; dependency: Phase 20 scope decision.

## 6. Redirect-policy regression coverage

- Problem: Redirect and connected-peer policy is security-sensitive and needs continued regression coverage.
- Scope: Add deterministic redirect chains, mixed DNS, HTTPS-only, and rejected destination cases.
- Files: `internal/acquire/*_test.go`, `docs/DESIGN.md`.
- Acceptance: redirects are revalidated and restricted destinations fail closed.
- Tests/output: local test-server cases with stable error categories.
- Security: this is transport-boundary work and requires security review.
- Difficulty: medium; dependency: none.

## 7. Structured reason-code documentation

- Problem: Automation needs stable explanations in addition to verdict names.
- Scope: Document and extend `reason_code`, missing/conflicting evidence, and recommended action fields without changing verdict semantics.
- Files: `internal/domain`, `internal/report`, `docs/REPORTS.md`, public docs.
- Acceptance: each verdict has a documented reason code and deterministic JSON fields.
- Tests/output: renderer and schema regression tests.
- Security: reason codes summarize evidence; they never claim malware detection.
- Difficulty: small; dependency: current schema review.

## 8. Ecosystem compatibility matrix

- Problem: npm and PyPI behavior varies across package metadata and release-file shapes.
- Scope: Add opt-in compatibility cases from sanitized live observations; keep default tests offline.
- Files: `fixtures`, adapter tests, `docs/REGISTRY_ADAPTERS.md`.
- Acceptance: supported and unsupported cases are classified without false MATCH results.
- Tests/output: fixture matrix and documented limitations.
- Security: no live registry dependency in default CI.
- Difficulty: medium; dependency: maintainer-approved fixture provenance.

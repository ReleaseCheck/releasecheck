# Submission Brief

This brief is an evidence summary for maintainer review. It is not a funding application and does not claim approval, adoption, users, or contributors. Public core release `v0.1.1` exists, but post-release stabilization remains open.

## Problem

ReleaseCheck helps maintainers and consumers inspect how a published npm or PyPI artifact relates to its claimed source repository and release. It records registry metadata, artifact hashes and inventories, source references, git references, deterministic differences, structural security observations, provenance-shaped evidence, limitations, and a policy-neutral verdict.

## Product evidence

- Go CLI with npm and PyPI paths.
- Non-executing artifact and archive inspection.
- Bounded downloads, safe path handling, duplicate detection, symlink observations, and SHA-256/SRI handling.
- Human, deterministic JSON schema 1.0, and SARIF 2.1.0 output.
- Offline fixtures and hosted Ubuntu/Windows/macOS plus race CI.
- Official Ubuntu-oriented GitHub Action that downloads an explicit core release and verifies its checksum.

## Honest status

Phase 17 requires audit follow-up: the v0.1.1 release workflow passed, while Action E2E run `39165054145` found a real tar PAX-header compatibility defect. The fix is merged to `main`, but v0.1.1 is immutable; a patched release and passing E2E rerun remain required. ReleaseCheck may be described as publicly released and tested, not production-proven or adoption-proven.

## Security model

ReleaseCheck never installs or executes inspected packages, source code, setup.py, build backends, lifecycle scripts, or arbitrary metadata-derived commands. It is not a sandbox, malware detector, vulnerability scanner, SBOM generator, Sigstore implementation, or SLSA implementation. Differences are evidence and may be legitimate build transformations.

## Contributor path

The core repository contains architecture, security, roadmap, fixture, release, and contributor guidance. `docs/GOOD_FIRST_ISSUES.md` contains eight scoped issue specifications with files, acceptance criteria, tests, security implications, difficulty, and dependencies. External issues are deliberately not fabricated.

## Drips Wave readiness

Readiness is improved by the public cross-platform release and hosted evidence, but a final numeric score is intentionally deferred until the patched release passes the Action E2E gate. Current gaps are the open post-release E2E defect, absent external issue history, and owner-controlled program participation.

## GrantFox / Stellar readiness

Indicative readiness score: **18/100** for a Stellar-oriented submission. ReleaseCheck has no Stellar-specific implementation or demonstrated Stellar ecosystem dependency. It should not add blockchain functionality merely to improve eligibility. The exact GrantFox program, current requirements, and eligibility must be confirmed by the owner before any submission discussion.

## Maintainer approvals required

1. Prepare and explicitly authorize the patched release after the tar PAX-header fix is reviewed.
2. Run the Action E2E workflow against that exact immutable patched release.
3. Record the final stabilization result and update the phase register.
4. Decide whether to create real external contributor issues and pursue any funding program.

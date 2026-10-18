# Submission Brief

This brief is an evidence summary for maintainer review. It is not a funding application and does not claim approval, adoption, users, contributors, or a public release.

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

Phase 16 is complete. Phase 17 is in progress. No public `v0.1.0` tag or GitHub Release exists. The read-only release rehearsal and manual post-release Action E2E workflows are prepared but have not been authorized or run against a public release. ReleaseCheck must be described as pre-v0.1, not production-proven or adoption-proven.

## Security model

ReleaseCheck never installs or executes inspected packages, source code, setup.py, build backends, lifecycle scripts, or arbitrary metadata-derived commands. It is not a sandbox, malware detector, vulnerability scanner, SBOM generator, Sigstore implementation, or SLSA implementation. Differences are evidence and may be legitimate build transformations.

## Contributor path

The core repository contains architecture, security, roadmap, fixture, release, and contributor guidance. `docs/GOOD_FIRST_ISSUES.md` contains eight scoped issue specifications with files, acceptance criteria, tests, security implications, difficulty, and dependencies. External issues are deliberately not fabricated.

## Drips Wave readiness

Indicative readiness score: **58/100**, subject to current program rules and maintainer eligibility. Strengths are a focused open-source security utility, documented contributor work, offline tests, and a clear non-execution boundary. Gaps are the missing public release, missing Action E2E evidence, absent external issue history, and owner-controlled program participation.

## GrantFox / Stellar readiness

Indicative readiness score: **18/100** for a Stellar-oriented submission. ReleaseCheck has no Stellar-specific implementation or demonstrated Stellar ecosystem dependency. It should not add blockchain functionality merely to improve eligibility. The exact GrantFox program, current requirements, and eligibility must be confirmed by the owner before any submission discussion.

## Maintainer approvals required

1. Review and merge the implementation PRs after hosted checks pass.
2. Run the read-only release rehearsal.
3. Authorize and publish the first public core release only after the release gate passes.
4. Run the Action E2E workflow against that exact immutable release.
5. Decide whether to create real external contributor issues and pursue any funding program.

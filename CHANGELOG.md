# Changelog

All notable changes to ReleaseCheck will be documented here.

## Unreleased

- Bootstrapped the repository documentation, scope, architecture, security boundary, and executable roadmap.
- Established the Go module, governance files, CI skeleton, and initial package boundary.
- Added the Phase 2 domain model, registry/cache contracts, typed evidence and error categories, and deterministic unit tests.
- Added bounded HTTPS artifact download, SHA-256 verification, secure temporary-file cleanup, and non-extracting ZIP/TAR.GZ inspection with adversarial tests.
- Added the npm packument adapter, SRI verification, GitHub source snapshot path, root-aware deterministic comparison, and offline npm fixtures.
- Added the PyPI project/release JSON adapter, all-file retention, SHA-256 verification, sdist/wheel selection semantics, explicit source-reference handling, and offline PyPI fixtures.
- Hardened deterministic comparison with canonical path validation, duplicate rejection, explicit type and unverifiable states, stable ordering, summaries, and symlink-target evidence.
- Added deterministic security observations for archive structure, symlink targets, special entries, size thresholds, npm lifecycle scripts, setup.py presence, and malformed metadata without executing package content.
- Added provenance evidence parsing for PyPI PEP 740 objects and npm DSSE/bundle-shaped evidence, with artifact binding and explicit absent, unavailable, invalid, and insufficient states. Cryptographic trust verification remains outside this phase.
- Added deterministic human, JSON schema 1.0, and SARIF 2.1.0 report renderers with explicit verdict precedence and stable exit-code mapping.

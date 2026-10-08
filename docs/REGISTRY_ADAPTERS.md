# Registry Adapter Guide

This guide is for contributors proposing support for a registry beyond npm and PyPI. New ecosystems are not part of v0.1 and require a roadmap decision before implementation.

## Contract

The small shared contract is `domain.RegistryAdapter`:

```go
type RegistryAdapter interface {
    Registry() Registry
    Resolve(context.Context, ReleaseRequest) (ReleaseMetadata, error)
}
```

The adapter translates registry-specific metadata into `ReleaseMetadata`. It must preserve the distinction between observed facts, claimed source identity, resolved immutable references, unavailable evidence, and limitations. It must not hide ambiguity by inventing a repository or git reference.

## Required behavior

An adapter proposal must document:

1. Package name and version grammar.
2. Release metadata endpoint and artifact selection policy.
3. Registry-provided digests and how they are checked.
4. Source repository fields and normalization rules.
5. Git tag/commit resolution and whether the result is immutable.
6. Source archive retrieval behavior.
7. Provenance or attestation fields and what they actually establish.
8. Rate limits, timeouts, redirects, and bounded response handling.
9. Missing, ambiguous, malformed, and conflicting metadata behavior.
10. Offline fixtures for representative and adversarial cases.

## Security requirements

Adapters must use the shared bounded acquisition path or an equivalently reviewed path. They must use HTTPS, context timeouts, bounded response sizes, strict URL validation, and typed errors. They must never invoke a package manager, execute package code, run an install hook, run a build backend, or execute a command derived from registry metadata.

Repository URLs, tags, filenames, archive contents, and attestation payloads are untrusted input. Do not follow symlinks or silently normalize away paths that could be security evidence.

## Tests and documentation

Adapter tests should use local HTTPS test servers or deterministic checked-in fixtures. Live registry tests must be opt-in and supplemental. Add fixture-matrix cases for matching evidence, missing source metadata, unavailable references, malformed archives, digest mismatch, and registry-specific edge cases.

Update `docs/DESIGN.md`, `PROJECT_CONTEXT.md`, `ROADMAP.md`, and the README when an adapter changes the evidence model or user-facing behavior. Include the date and primary sources for current registry behavior.

## Review standard

The default implementation agent is Antigravity with Claude Sonnet 4.6. Use Codex for independent review of archive handling, source resolution, provenance semantics, or a large adapter refactor. A registry adapter is not complete when it merely downloads an artifact; it is complete when its evidence and limitations are reproducible and honestly reported.

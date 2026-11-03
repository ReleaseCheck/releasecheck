# Companion Repository Audit

This document records the independent review boundary for the three public ReleaseCheck repositories. It prevents the core audit from treating companion repositories as assumed evidence.

## Responsibilities

| Repository | Reviewed responsibility | Boundary confirmed |
| --- | --- | --- |
| `ReleaseCheck/releasecheck` | Core Go CLI, verification engine, reports, fixtures, security policy, and release workflow | Source of implementation truth |
| `ReleaseCheck/releasecheck-action` | Thin GitHub Actions wrapper that downloads a selected core release, verifies its checksum, and propagates the result | No second verification engine; package code is never executed |
| `ReleaseCheck/releasecheck-docs` | Curated user guides, CI examples, attestation guidance, and limitations | Links to core semantics; does not redefine them |

## Review checks

- README scope and limitations agree with the core evidence model.
- Action acquisition is platform-aware for Ubuntu, macOS, and Windows and verifies the selected release archive before execution.
- Action CI covers its validation gate and all three declared runner platforms.
- Documentation describes external attestation verification as interoperable evidence, not as native ReleaseCheck trust verification.
- Documentation does not claim package safety, source equivalence, universal provenance support, adoption, or novelty.
- Repository-local PR templates require Summary, Verification, Security, and documentation/limitation sections.

## Evidence policy

This is a repository-structure and documentation review, not an independent cryptographic audit of GitHub, npm, PyPI, Sigstore, or SLSA. Hosted CI results and exact commit IDs must be recorded at each release checkpoint. A passing companion-repository check does not upgrade a package verdict or prove that a published artifact is safe.

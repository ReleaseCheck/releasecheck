# Product Positioning

Status: v0.1.2, reviewed 2026-10-19 (WAT)

This document answers the audit concern that ReleaseCheck overlaps with existing supply-chain tooling. The overlap is real and intentional. ReleaseCheck must be positioned as a focused evidence-comparison layer, not as a replacement for registry or provenance infrastructure.

## Core question

For one selected npm or PyPI release file:

1. What artifact and version did the registry publish?
2. What source repository and git reference did the metadata claim?
3. Could the referenced source snapshot be retrieved safely?
4. How do the source and artifact file inventories and hashes relate?
5. What security-relevant structural observations and provenance evidence exist?
6. What remains unknown or insufficient for a stronger conclusion?

## Boundary matrix

| Capability | Existing adjacent tools | ReleaseCheck v0.1 |
| --- | --- | --- |
| Registry hashes and metadata | npm and PyPI APIs | Consumes and reports them for the selected release |
| Signatures and attestations | npm signatures, PyPI attestations, Sigstore | Consumes supported evidence; does not recreate trust infrastructure |
| Provenance standards | SLSA, in-toto, registry builders | Reports available structured evidence and limitations |
| Reproducible package builds | vltpkg/reproduce and ecosystem-specific builders | Does not install, build, or claim reproducibility |
| Source/artifact relationship | Often ecosystem- or workflow-specific | Compares inventories and hashes with common semantics across npm and PyPI |
| Archive safety | General archive libraries and scanners | Bounded, non-extracting inspection for release verification |
| Malware/vulnerability detection | Security scanners and vulnerability databases | Explicitly out of scope |
| Repository security posture | OpenSSF Scorecard and similar tools | Explicitly out of scope |
| Dependency graph funding/metadata | GUAC and other graph systems | Emits release evidence; does not build a global graph |

## Defensible claim

ReleaseCheck provides a deterministic, non-executing report about the relationship between a published npm/PyPI artifact and its claimed source release, when the available metadata permits that relationship to be examined.

This claim is deliberately narrower than “the package is safe,” “the package is reproducible,” “the source is authentic,” or “the artifact was built by a trusted workflow.”

## What makes the product useful despite overlap

- One report contract across npm and PyPI.
- Explicit separation of known facts, inferences, differences, warnings, unavailable evidence, and limitations.
- Safe inspection that never runs package code, installation scripts, or arbitrary builds.
- A comparison result that treats generated files and build transformations as evidence requiring interpretation, not automatic maliciousness.
- CI-friendly JSON, SARIF, and exit statuses for incomplete or review-required evidence.

## Claims not permitted

Do not claim ReleaseCheck is first, unique, complete, a replacement for Sigstore/SLSA/npm/PyPI, a malware detector, a vulnerability scanner, a reproducible-build system, or proof that matching files are safe. Do not claim adoption, users, contributors, funding, or ecosystem demand without repository evidence.

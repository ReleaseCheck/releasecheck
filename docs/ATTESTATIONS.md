# Attestation Interoperability

ReleaseCheck is compatible with attestation ecosystems without replacing them.

## Evidence roles

| System | What it can establish | What ReleaseCheck does with it |
| --- | --- | --- |
| Sigstore/Cosign | Signature, certificate identity, bundle and transparency-log evidence when verified | Consume an external verification result or parse supplied evidence; do not claim cryptographic verification unless a supported verifier performed it |
| SLSA provenance | Claims about builder, source/materials, build instructions, and artifact subjects | Report provenance claims separately from source/artifact comparison |
| GitHub artifact attestations | A cryptographically verifiable relationship between an artifact and GitHub build workflow/repository | Accept as adjacent build evidence when available; it does not establish npm/PyPI source equivalence |
| npm provenance/signatures | Registry-specific publishing and provenance evidence | Consume as registry evidence; do not recreate npm trust infrastructure |
| PyPI attestations | Release-file attestations and PEP 740 provenance | Consume as registry evidence; cryptographic trust-root verification remains outside v0.1 |

## Recommended external verification

For a GitHub-produced binary, a consumer may independently verify an attestation with:

```text
gh attestation verify ./releasecheck -R ReleaseCheck/releasecheck
```

For Sigstore artifacts, use the official Cosign verification flow appropriate to the artifact type and signature format. For SLSA provenance, verify the envelope/signature, trusted builder identity, expected build type, subject digest, and policy expectations. These commands are external verification steps; ReleaseCheck never executes package code or arbitrary build commands to obtain them.

## Current release limitation

ReleaseCheck v0.1.2 publishes signed-by-checksum GitHub release archives but does not publish GitHub artifact attestations or claim Sigstore verification. The release asset SHA-256 manifest proves only the downloaded asset's digest matches the published manifest; it is not a signature, provenance statement, or source-equivalence proof.

Primary references: [GitHub artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations), [GitHub verification](https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations), [Sigstore verification](https://docs.sigstore.dev/cosign/verifying/verify/), [SLSA v1.2 provenance](https://slsa.dev/spec/v1.2/provenance), and [SLSA artifact verification](https://slsa.dev/spec/v1.2/verifying-artifacts).

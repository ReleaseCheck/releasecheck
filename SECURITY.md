# Security Policy

ReleaseCheck is security-sensitive infrastructure that processes untrusted package artifacts. The v0.1 implementation includes a CLI, but this policy applies equally to the CLI, internal packages, tests, and future SDK surface.

## Non-execution boundary

ReleaseCheck must never execute package code or invoke package managers while inspecting a release. In particular it must not run npm lifecycle scripts, `npm install`, Python installation, `setup.py`, arbitrary build backends, post-install hooks, `eval`, or metadata-derived shell commands.

## Untrusted input

Artifacts, registry metadata, repository URLs, git references, archive names, archive contents, and attestations are untrusted input. Implementations must use HTTPS, timeouts, bounded downloads, bounded archive expansion, bounded file counts and sizes, safe temporary storage, strict path validation, and cleanup. Artifact downloads reject loopback, private, link-local, multicast, and unspecified destinations by default and recheck redirect targets. Path traversal, absolute paths, symlinks, special entries, malformed archives, and suspicious metadata must be handled explicitly and never blindly followed or extracted.

## Deterministic observations

The security-analysis package reports structural facts such as lifecycle-script declarations, `setup.py` presence, symlinks, special entries, unsafe symlink targets, malformed metadata, and configured size thresholds. These are review signals, not malware detection, exploit detection, or proof of malicious intent. A lifecycle script is recorded as metadata and is never executed. Invalid metadata remains an invalid observation instead of being treated as absent.

Provenance handling is also evidence-only. A structurally valid attestation may be reported as `insufficient` until its signature, certificate chain, transparency log, and registry trust root have been verified by an explicitly supported verifier. ReleaseCheck does not treat a provenance object as proof of source equivalence, benignness, or maliciousness.

## Reporting boundary

Evidence of a difference is not proof of maliciousness. Provenance is not proof that code is benign. ReleaseCheck is not a malware detector, vulnerability scanner, SBOM generator, Sigstore implementation, or SLSA implementation. Reports must distinguish known facts, inferences, warnings, limitations, and unavailable evidence.

## Reporting a vulnerability

Until a security contact is configured, report suspected vulnerabilities privately to the project maintainer rather than opening a public issue with exploit details. Include a minimal reproduction, affected commit or version, impact, and suggested mitigation. Do not submit package artifacts that execute on extraction or require installation.

## Secure development

Security-sensitive changes require focused tests and an independent review when they affect archive parsing, source resolution, provenance verification, trust decisions, or process execution boundaries. CI must use least privilege and must not process live untrusted packages as part of ordinary tests.

ReleaseCheck is a local evidence tool, not a network-isolated service. Private mirrors require explicit controlled configuration and security review; they are not silently trusted by the default artifact downloader.

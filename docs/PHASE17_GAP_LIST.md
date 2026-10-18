# Phase 17 Gap List

Audit date: 2026-10-18. This is an implementation gap list, not a release approval.

## Confirmed defects or missing implementation

- No public core release exists, so the Action cannot yet complete a real public-release E2E run.
- The core repository previously had no non-publishing hosted release rehearsal; `scripts/rehearse-release.sh` and the read-only manual workflow now provide that path, but they still require a hosted run.
- Reports previously exposed verdicts without stable reason, explanation, missing-evidence, conflict, and next-action fields; the v1.0 JSON contract now includes those fields and focused tests.
- The Action has deterministic shell validation but its manual public-release E2E workflow has not yet been run against a real core release.

## Missing evidence

- Hosted release-rehearsal run and artifact inspection.
- Public core release artifact and checksum verification.
- Action E2E run against that immutable public artifact.
- Cross-machine reproducibility measurement beyond the documented reproducible-enough build inputs.

## Documentation-only or governance gaps

- Branch protection/rulesets are owner-controlled and not evidenced as enabled.
- External contributor issues have not been created; `docs/GOOD_FIRST_ISSUES.md` contains ready-to-file specifications.
- Organization profile and funding configuration remain owner decisions.

## Future or out-of-scope work

- Cryptographic Sigstore/DSSE trust verification, malware detection, SBOM generation, arbitrary build reproduction, additional registries, and a public SDK remain outside v0.1.
- SARIF is implemented and documented; it is not a trust decision.

## Security conclusion

No new P0/P1 code defect was identified in this pass. The release and Action blockers are evidence/operational gates, not permission to weaken the non-execution or immutable-dependency boundaries.

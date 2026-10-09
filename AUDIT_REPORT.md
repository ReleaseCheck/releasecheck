# ReleaseCheck Independent Audit Report

Audit date: 2026-10-09 (Africa/Lagos)

Scope: independent local and GitHub audit of the ReleaseCheck organization, the three discovered repositories, the Go implementation, the GitHub Action, release workflows, documentation, repository governance, license records, and funding readiness. This report records the audit branches and does not declare a release.

## Executive verdict

ReleaseCheck is a real, tested pre-v0.1.0 Go CLI for deterministic non-executing npm and PyPI release-integrity evidence. The core repository is contributor-readable and the hosted core test matrix is green. It is not ready to publish v0.1.0 yet.

The audit confirmed two important P1 defects: the Action checksum parser rejected checksum lines emitted by the core release script, and the tag-triggered release workflow published without running its own validation gate. The original audit PR was merged, including the Action parser and release-gate changes. This follow-up PR corrects the separate checksum working-directory defect and validates the exact publish-job artifacts. A P1 network-boundary hardening change was also merged after review. No release or tag has been created.

Contributor readiness is moderate: the code, fixtures, governance files, CI, and documentation exist, but the public repositories have no protected `main`, no open contributor issues, no public release, and no verified end-to-end Action run. Funding readiness is opportunity-dependent and unconfirmed; Drips is the strongest near-term fit, while Stellar/GrantFox fit is currently weak or unknown because ReleaseCheck has no Stellar-specific capability.

## Organization inventory

Authenticated identity: `Marvelg256` via GitHub CLI. Organization: `ReleaseCheck`. Three public repositories were returned; no additional public organization repositories were found through the authenticated repository listing.

| Repository | Local path | Baseline local HEAD | Remote `main` | Baseline sync | Audit branch |
| --- | --- | --- | --- | --- | --- |
| `ReleaseCheck/releasecheck` | `C:\Users\user\Projects\ReleaseCheck\releasecheck` | `4b469e51fc84ff675e4f6f7787a082770357d9d4` | same | synchronized | `audit/2026-10-09` |
| `ReleaseCheck/releasecheck-action` | `C:\Users\user\Projects\ReleaseCheck\releasecheck-action` | `af6e934c098505ef8ce736754dd04ed1cfe47770` | same | synchronized | `audit/2026-10-09` |
| `ReleaseCheck/releasecheck-docs` | `C:\Users\user\Projects\ReleaseCheck\releasecheck-docs` | `f00884b333070ad08e1af055535b0dce750f5e74` | same | synchronized | `audit/2026-10-09` |

Baseline collection used `git fetch --prune origin`, `git status --short --branch`, `git rev-list --left-right --count HEAD...origin/main`, remote inspection, and authenticated GitHub API/CLI queries. The original `main` worktrees were clean before audit branches were created. Ignored core build/cache paths were `.gocache/`, `releasecheck.exe`, and `tmp/`; no secrets were printed or staged.

## Product and implementation assessment

The implementation matches the documented product boundary in the important respects:

- Go CLI with npm and PyPI paths, deterministic inventories, hashes, comparisons, evidence, JSON/SARIF/human reports, and stable exit codes.
- No package installation, package execution, lifecycle execution, Python build backend, arbitrary build, shell invocation, or metadata-derived command execution.
- Bounded HTTPS acquisition, SHA-256/SRI handling, temporary-file cleanup, non-extracting ZIP/TAR.GZ inspection, path and duplicate checks, symlink/special-entry observations, and offline fixtures.
- GitHub source snapshots are supported where source identity and a usable reference can be resolved; source archive identity is not cryptographically proven against the claimed Git object.
- Provenance-shaped evidence is parsed and bound where possible, but cryptographic trust-root, DSSE, Sigstore, transparency-log, and registry-signature verification is not claimed.

The public SDK remains intentionally deferred. The current internal packages are not a compatibility promise.

## Confirmed findings and corrections

### RC-AUDIT-001 - Action checksum lookup rejected generated checksum formats

Severity: P1. Repository: `releasecheck-action`.

Evidence: `scripts/run-releasecheck.sh` compared `$2` exactly with the unprefixed archive name. The core `scripts/build-release.sh` invokes `sha256sum ./*.tar.gz ./*.zip`, and GNU `sha256sum` emits binary-mode entries such as `HASH *./releasecheck-v0.1.0-linux-amd64.tar.gz`. Text-mode implementations can emit `HASH  ./name`. Neither form was reliably accepted.

Impact: a valid published release could fail closed with `archive checksum is missing`, preventing the Action from running.

Correction on `audit/2026-10-09`: normalize an optional leading `*`, accept prefixed and unprefixed names, require exactly one matching 64-character hexadecimal digest, reject malformed/ambiguous/missing entries, verify the downloaded digest, and extract only the expected binary member.

Validation: `bash tests/validate-action.sh` passed with prefixed binary-mode, prefixed text-mode, unprefixed, missing, malformed, and mismatch cases. The first hosted PR run exposed a missing executable bit on `scripts/run-releasecheck.sh`; commit `1e69ad8` corrected the mode, and hosted Action checks `37930070590` and `37930076098` passed. No public release was created.

Residual risk: the Action remains Ubuntu/Bash-oriented and has no live public core release to test end to end.

### RC-AUDIT-002 - Tag release workflow lacked a pre-publication validation gate

Severity: P1. Repository: `releasecheck/.github/workflows/release.yml`.

Evidence: the previous tag workflow checked out the tag, built archives, and immediately ran `gh release create` with `contents: write`; it did not run the repository test, vet, race, formatting, version, or artifact checks in that workflow.

Impact: a tag could publish artifacts even when the tagged commit had not passed the required validation suite or archive checks.

Correction merged from the original audit branch: added a read-only `validate` job before the publish job. It runs formatting, whitespace, vet, tests, race tests, CLI build, `VERSION`/tag consistency, archive count checks, and checksum validation. This follow-up also validates the exact archives rebuilt by the publish job immediately before `gh release create`. Only the publish job receives `contents: write`; no release has been published.

Validation: the workflow YAML was reviewed locally. The build script rehearsal reached all target builds but could not finish on this workstation because `zip` is not installed. Hosted validation remains required after review/merge.

Residual risk: GitHub repository rules do not currently require the workflow or prevent direct tag creation; owner action is required for governance.

### RC-AUDIT-003 - DNS policy was not enforced at the connection boundary

Severity: P1 hardening. Repository: `releasecheck/internal/acquire`.

Evidence: the prior implementation resolved a hostname for preflight policy and then allowed the HTTP transport to resolve/dial separately. That creates a time-of-check/time-of-use gap for changing DNS responses. The exact exploitability depends on resolver and transport behavior, but the boundary was not strong enough for the documented security intent.

Correction on `audit/2026-10-09`: under the default restricted policy, clone the transport, disable proxy routing, and validate the actual remote peer after dialing. Keep the explicit private-network opt-in for controlled fixtures/private mirrors. Add mixed-DNS and connected-private-peer tests.

Validation: focused acquisition tests passed; full tests, vet, formatting, and race tests passed with isolated local caches.

Residual risk: ReleaseCheck is still a local analysis tool, not a general network sandbox. Metadata clients and user-supplied transport policies require continued review.

### RC-AUDIT-004 - Phase 16 status mismatch

Severity: P1 documentation/governance. The phase register previously said `AUDIT REQUIRED` while the detailed Phase 16 section said `NOT STARTED`.

Correction: the register, detailed phase, README, `PROJECT_CONTEXT.md`, release documentation, Action README, and user docs now identify Phase 16 as `IN PROGRESS`. It remains incomplete until the fixes are reviewed and merged through normal pull requests.

## License assessment

All three repositories contain the same 9,977-byte Apache License 2.0 text with SHA-256 `2D305689E4E662EFCA61D03A008B4C03B8A6915CFDA5738B87A881DCA0C26093`. The text begins with `Apache License Version 2.0, January 2004` and contains the standard terms and ending. The canonical reference is https://www.apache.org/licenses/LICENSE-2.0.txt.

GitHub reports `NOASSERTION`/`Other` for the API license field in all three repositories. That is a classifier/metadata result, not evidence that the license text is invalid. No license rewrite was justified by this audit. The maintainer should inspect GitHub's rendered license detection after future repository metadata changes; do not claim the classifier is fixed until it changes.

## GitHub governance and release state

- Organization profile API: public organization exists, but description, website, location, and company fields are empty.
- `ReleaseCheck/.github` profile repository: API returned 404. No public organization profile README was found through the accessible inventory.
- All three repositories are public with default branch `main`; all have issues/projects/wiki enabled; all had zero open issues and zero open pull requests at audit time.
- `main` branch protection API returned `404 Branch not protected` for the core repository. Rulesets returned an empty list. Equivalent protection checks for the other repositories should be repeated by the organization owner if the API view changes; no protection was evidenced in the initial inventory.
- Core workflow run `37855548443` on commit `4b469e5` completed successfully. Earlier successful runs `37852987328` and `37853277292` remain recorded. Historical failures/cancellation are visible in GitHub and should not be hidden.
- No GitHub releases or tags were returned for the core repository. The Action therefore cannot complete a real public-release end-to-end run.
- No `FUNDING.yml` or verified sponsorship profile was found in the repository trees inspected.

## Test matrix

| Check | Result | Evidence |
| --- | --- | --- |
| Core `go test -p 1 -count=1 ./...` | PASS | Completed on isolated repository cache/temp paths |
| Core `go test -race -p 1 -count=1 ./...` | PASS | Completed on 2026-10-09 |
| Core `go vet ./...` | PASS | Completed on 2026-10-09 |
| Core `gofmt -l .` and `git diff --check` | PASS | No output/errors |
| Core focused network tests | PASS | Mixed DNS and connected-peer cases |
| Action `bash tests/validate-action.sh` | PASS | Git Bash; six checksum/security cases |
| Core release build rehearsal | PARTIAL | All Go target builds reached; stopped because workstation lacks `zip` |
| Hosted core CI | PASS | PR runs `37929923027` and `37929949414`; Ubuntu, Windows, macOS, and race jobs passed |
| Hosted Action CI | PASS after correction | PR/push runs `37930070590` and `37930076098` passed after executable-bit fix |
| `govulncheck`, `gosec`, `golangci-lint` | NOT RUN | Not installed; no result claimed |
| Live registry tests | NOT RUN | Offline fixtures are authoritative for default tests |
| Public Action end-to-end | BLOCKED | No public core release exists |
| Branch protection/rulesets | PARTIAL | Core explicitly unprotected/empty rulesets; organization profile/settings require owner review |

The initial Go test attempts encountered Windows cache/process-lock issues. Re-running with repository-local `GOCACHE`, `GOTMPDIR`, and serialized package execution completed successfully; this environment detail is retained rather than hidden.

## Funding and program fit

Audit date: 2026-10-09. No funding application, approval, adoption, contributor count, or award is claimed.

| Route | Verified requirement/evidence | ReleaseCheck fit | Missing/manual action |
| --- | --- | --- | --- |
| Drips Wave | Maintainers submit a repository with technical issues; selection is discretionary. Issues can be added through the Drips app or program label, and maintainers are expected to scope and review contributor work. Official sources: https://github.com/apps/drips-wave, https://docs.drips.network/wave/terms-and-rules/, https://docs.drips.network/wave/maintainers/faq/ | Moderate to strong for focused Action/release/security documentation issues, once issues are real and well-scoped | Owner must install/connect Drips Wave, confirm the current Wave program/deadline, submit the repositories, and create/maintain genuine issues. No current Wave admission was verified. |
| GrantFox / Stellar | The audit found no authoritative GrantFox source that establishes a current program identity. Official Stellar materials describe SDF grants and SCF for projects building on/contributing to Stellar or Soroban; public-good awards require ecosystem validation and direct utility. Sources: https://stellar.org/grants-and-funding, https://stellar.gitbook.io/scf-handbook/scf-awards/official-rules-for-submissions, https://stellar.gitbook.io/scf-handbook/supporting-programs/public-goods-award/official-rules | Weak/unknown today. ReleaseCheck has no demonstrated Stellar/Soroban feature and must not add blockchain functionality solely to pursue funding | Identify the exact GrantFox program/URL, confirm eligibility, and only apply if a genuine Stellar ecosystem use case and validation exist. KYC/team/legal requirements are manual. |
| GitHub Sponsors | Receiving funds requires an eligible maintainer/organization in a supported region and payment/tax onboarding. Official source: https://docs.github.com/en/sponsors/getting-started-with-github-sponsors/about-github-sponsors | Unknown until account and region are verified; the public organization has no verified Sponsors state in this audit | Owner checks the Nigeria/organization eligibility and Stripe onboarding, creates profile, then adds verified `FUNDING.yml` only after profile activation. |
| Open Collective | A financially active Collective needs a fiscal host or legal organization; Open Source Collective advertises a 10% host fee. Sources: https://opencollective.com/opensource/apply/intro, https://docs.oscollective.org/welcome-and-introduction-to-osc/fees | Moderate for transparent open-source maintenance funding | Owner chooses a host, confirms geographic/legal fit, applies, completes verification, and creates a real Collective. No profile exists from this audit. |
| Polar and similar services | These are sponsorship/monetization platforms, not grant awards; no project profile or revenue model was verified in this audit | Unknown | Owner must evaluate current terms, payout eligibility, tax handling, and whether a real maintenance offering exists. Do not create fictional tiers or claims. |

## Remaining manual actions

1. Review and merge follow-up PR #2 after its green hosted checks; do not push these changes directly to `main`.
2. Require pull requests and successful core/Action CI before merging. Protect `main` and configure a release/tag policy through GitHub repository settings or organization rulesets.
3. Add a minimal organization profile README in `ReleaseCheck/.github` and set a factual organization description/website if desired.
4. Decide whether to enable Discussions/Wiki and issue templates based on actual maintainer capacity; do not create superficial issues.
5. Install `zip` or run the release rehearsal on Ubuntu/GitHub Actions, then validate a non-public/test release path. Do not publish v0.1.0 until Phase 16 and Phase 17 gates pass.
6. Decide, with verified account ownership and eligibility, whether to configure GitHub Sponsors, Open Collective, or Drips. Add no funding identifier until the account is real.
7. Clarify the exact GrantFox program before discussing eligibility; current evidence does not establish that ReleaseCheck qualifies.

## Release decision

`v0.1.0` is **BLOCKED**. The core implementation is substantial and testable, and the original audit remediation is merged. The checksum working-directory correction and final publish-job validation remain in follow-up PR #2. There is no public release, no protected main branch, no Action end-to-end evidence, and no actual tag-triggered release rehearsal. The product must not be described as fully release-ready or funded.

## Prioritized next steps

1. Review and merge follow-up PR #2 after its hosted checks pass.
2. Configure branch protection/rulesets and release governance.
3. Run the hosted release-validation job without publishing, then inspect artifacts and checksums.
4. Re-run an independent review of the merged changes and update this report/recommendation.
5. Only then prepare the actual v0.1.0 release and a real Action end-to-end test.
6. Pursue Drips or other funding only with real scoped issues, verified account eligibility, and no fabricated traction claims.

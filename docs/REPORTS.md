# Reports and Exit Status

ReleaseCheck has one evidence model and three renderings: human-readable terminal output, JSON, and SARIF. Renderers do not make independent trust decisions.

## JSON contract

The JSON schema version is `1.0`. The top-level shape is:

```json
{
  "schema_version": "1.0",
  "identity": {"registry": "npm", "name": "example", "version": "1.0.0"},
  "artifact": {"identity": {}, "filename": "example-1.0.0.tgz", "url": "https://example.invalid/example.tgz"},
  "source": {"url": "https://github.com/example/project"},
  "git": {"ref": "v1.0.0", "commit": "...", "immutable": true},
  "comparisons": [],
  "security_observations": [],
  "evidence": [],
  "provenance": [],
  "limitations": [],
  "verdict": "REVIEW"
}
```

Important fields are:

- `identity`: registry, package name, and selected version.
- `artifact`: selected filename, URL, expected registry digest, observed digest, and size.
- `source`: claimed repository identity. This is a claim, not proof.
- `git`: claimed or resolved reference and whether the reference is immutable.
- `comparisons`: path-level identical, source-only, artifact-only, modified, or unverifiable results.
- `security_observations`: deterministic structural and metadata signals.
- `evidence`: normalized registry facts and limitations.
- `provenance`: parsed provenance state and supported fields.
- `limitations`: reasons a stronger conclusion was unavailable.
- `verdict`: policy-neutral summary derived from the evidence.

The default JSON output omits `generated_at` so equivalent evidence produces stable output. The report schema is not a promise that every future internal package remains compatible; schema changes require an explicit versioning decision.

## Verdicts

| Verdict | Meaning |
| --- | --- |
| `MATCH` | Complete available comparison evidence is identical and no warning or limitation prevents that result. |
| `REVIEW` | Differences, warnings, invalid evidence, or insufficient provenance require human review. |
| `INCOMPLETE` | Important source, git, comparison, or other evidence is unavailable. |
| `ERROR` | Reserved for operational or rendering failure rather than a report verdict. |

No verdict means that the package is harmless. A difference is not proof of maliciousness.

## Exit codes

| Code | Meaning |
| ---: | --- |
| `0` | `MATCH` |
| `1` | `REVIEW` or `INCOMPLETE` |
| `2` | Invalid command-line usage |
| `3` | Operational or rendering error |

## SARIF

SARIF output is version `2.1.0`. It maps comparison differences, security warnings, invalid metadata, and non-present provenance states into stable results. SARIF is intended for CI and code-scanning integrations; JSON remains the complete ReleaseCheck evidence contract.

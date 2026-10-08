# Usage Guide

ReleaseCheck performs network retrieval and deterministic inspection. It does not install or execute the package being inspected.

## Build

From the repository root, with Go 1.25 or newer:

```text
go build -trimpath -o releasecheck ./cmd/releasecheck
```

Check the binary and available commands:

```text
releasecheck --version
releasecheck --help
```

## Verify npm

```text
releasecheck verify npm PACKAGE
releasecheck verify npm PACKAGE VERSION
releasecheck verify npm --json --output npm-report.json PACKAGE VERSION
```

The npm path resolves registry metadata, selects the requested version, retrieves the published artifact, uses available repository and git metadata, and compares against a supported source snapshot where possible.

## Verify PyPI

```text
releasecheck verify pypi PACKAGE
releasecheck verify pypi PACKAGE VERSION
releasecheck verify pypi --artifact PACKAGE-VERSION-py3-none-any.whl PACKAGE VERSION
```

PyPI releases can contain multiple files. Use `--artifact` to select a specific wheel or source distribution when the default selection is not the file you intend to inspect. Wheels are built distributions and should not be expected to match source byte-for-byte.

## Output selection

Human-readable output is the default. Select one machine-readable format:

```text
releasecheck verify npm --json PACKAGE
releasecheck verify npm --sarif --output report.sarif PACKAGE
```

`--json` and `--sarif` cannot be used together. `--output` writes the selected report to a local file with restrictive permissions. Reports may contain package names, URLs, paths, hashes, and repository metadata; treat them as potentially sensitive operational data.

## Timeouts and failures

Use `--timeout 60s` to bound registry and source operations. A usage error returns `2`; an operational or rendering error returns `3`. A valid report returns `0` for `MATCH` and `1` for `REVIEW` or `INCOMPLETE`.

Do not interpret `MATCH` as proof of safety. Do not interpret `REVIEW` as proof of maliciousness. Read the evidence, limitations, and comparison categories in the report.

## Safe operation

Do not run package managers, install commands, build commands, or arbitrary scripts around a ReleaseCheck invocation merely to help it inspect a package. ReleaseCheck is designed to inspect downloaded metadata, archives, source snapshots, and deterministic evidence without executing package content.

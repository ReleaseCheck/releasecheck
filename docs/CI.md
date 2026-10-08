# CI Integration

ReleaseCheck can be used as a review gate while preserving the full report as an artifact.

## Generic shell example

```text
releasecheck verify npm --json --output releasecheck.json PACKAGE VERSION
status=$?
cat releasecheck.json
exit "$status"
```

The command returns `0` for a complete `MATCH`, `1` for `REVIEW` or `INCOMPLETE`, `2` for usage errors, and `3` for operational or rendering errors. Teams may choose to allow code `1` while collecting the report for review, or fail the job on any nonzero code.

## GitHub Actions example

```yaml
- name: Verify published release
  run: |
    releasecheck verify npm --json --output releasecheck.json \
      example-package "${{ github.ref_name }}"

- name: Upload ReleaseCheck report
  if: always()
  uses: actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02 # v4.6.2
  with:
    name: releasecheck-report
    path: releasecheck.json
```

Pin third-party actions to reviewed commit SHAs in production workflows. The repository's own workflows use that policy. The command above is an integration shape, not a claim that every package name or tag maps directly to a registry version.

## Security guidance

Treat the report as untrusted data. Do not interpolate package metadata, URLs, filenames, or report fields into shell commands. Keep the package manager and build system outside the verification job unless a separate, explicitly reviewed job needs them.

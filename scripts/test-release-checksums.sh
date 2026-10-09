#!/usr/bin/env bash
set -euo pipefail

# Regression test for the release workflow's checksum working directory.
test_root="$(mktemp -d)"
trap 'rm -rf "$test_root"' EXIT
mkdir -p "$test_root/dist"
printf 'release artifact\n' > "$test_root/dist/releasecheck-v0.1.0-linux-amd64.tar.gz"
printf 'windows artifact\n' > "$test_root/dist/releasecheck-v0.1.0-windows-amd64.zip"
(cd "$test_root/dist" && sha256sum ./*.tar.gz ./*.zip > SHA256SUMS)

# The manifest names are relative to dist, so the old repository-root command
# must fail. This keeps a future workflow edit from silently reintroducing it.
if (cd "$test_root" && sha256sum --check dist/SHA256SUMS >/dev/null 2>&1); then
  echo 'repository-root checksum validation unexpectedly passed' >&2
  exit 1
fi

(cd "$test_root/dist" && sha256sum --check SHA256SUMS)
echo "release-checksum-validation=PASS"

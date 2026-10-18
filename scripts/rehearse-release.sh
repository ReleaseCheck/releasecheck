#!/usr/bin/env bash
set -euo pipefail

version="${1:?usage: rehearse-release.sh vX.Y.Z}"
if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "release version must match vMAJOR.MINOR.PATCH: $version" >&2
  exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
expected_version="${version#v}-dev"
[[ "$(tr -d '\r\n' < "$root/VERSION")" == "$expected_version" ]] || {
  echo "VERSION must be $expected_version for a rehearsal" >&2
  exit 1
}

bash "$root/scripts/build-release.sh" "$version"
dist="$root/dist"
mapfile -t archives < <(find "$dist" -maxdepth 1 -type f \( -name '*.tar.gz' -o -name '*.zip' \) -printf '%f\n' | sort)
[[ "${#archives[@]}" -eq 5 ]] || { echo "expected five release archives" >&2; exit 1; }
(cd "$dist" && sha256sum --check SHA256SUMS)

linux_archive="$dist/releasecheck-${version}-linux-amd64.tar.gz"
binary="$dist/releasecheck-${version}-linux-amd64/releasecheck"
mkdir -p "${binary%/*}"
tar -xzf "$linux_archive" -C "$dist"
[[ -x "$binary" ]] || { echo "Linux release binary is missing or not executable" >&2; exit 1; }
version_output="$("$binary" --version)"
[[ "$version_output" == "$version (commit "* ]] || { echo "unexpected binary version: $version_output" >&2; exit 1; }

for archive in "${archives[@]}"; do
  [[ -s "$dist/$archive" ]] || { echo "empty release archive: $archive" >&2; exit 1; }
done

cat > "$dist/RELEASE_REHEARSAL.md" <<EOF
# Release rehearsal: $version

- Five archives were built by scripts/build-release.sh.
- SHA256SUMS was verified from inside dist/.
- The Linux amd64 binary reported $version and was executable.
- No Git tag, GitHub Release, package publication, or mutable branch dependency was created.

This file is rehearsal evidence, not release notes and not a public release artifact.
EOF
echo "release rehearsal passed for $version"

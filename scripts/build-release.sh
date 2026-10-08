#!/usr/bin/env bash
set -euo pipefail

version="${1:?usage: build-release.sh vX.Y.Z}"
if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "release version must match vMAJOR.MINOR.PATCH: $version" >&2
  exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist="$root/dist"
rm -rf "$dist"
mkdir -p "$dist"

commit="$(git -C "$root" rev-parse HEAD)"
epoch="$(git -C "$root" show -s --format=%ct HEAD)"
ldflags="-s -w -X github.com/releasecheck/releasecheck/internal/cli.version=${version} -X github.com/releasecheck/releasecheck/internal/cli.commit=${commit}"

targets=("linux amd64" "linux arm64" "darwin amd64" "darwin arm64" "windows amd64")
for target in "${targets[@]}"; do
  read -r goos goarch <<< "$target"
  suffix=""
  if [[ "$goos" == "windows" ]]; then suffix=".exe"; fi
  name="releasecheck-${version}-${goos}-${goarch}"
  binary="$dist/$name/releasecheck${suffix}"
  mkdir -p "$(dirname "$binary")"
  echo "building $goos/$goarch"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build \
    -trimpath -buildvcs=false -ldflags "$ldflags" \
    -o "$binary" ./cmd/releasecheck
  touch -d "@$epoch" "$binary"
  if [[ "$goos" == "windows" ]]; then
    (cd "$dist/$name" && TZ=UTC zip -X -q "$dist/${name}.zip" "releasecheck.exe")
  else
    (cd "$dist" && tar --sort=name --mtime="@$epoch" --owner=0 --group=0 --numeric-owner -cf - "$name" | gzip -n > "${name}.tar.gz")
  fi
  rm -rf "$dist/$name"
done

(cd "$dist" && sha256sum ./*.tar.gz ./*.zip > SHA256SUMS)
echo "release artifacts written to $dist"

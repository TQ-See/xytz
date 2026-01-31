#!/bin/bash

set -e

VERSION=${1:-"dev"}
VERSION=${VERSION#v}
DIST_DIR="dist"
TEMP_DIR=$(mktemp -d)

echo "Building xytz version: $VERSION"

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

platforms=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

for platform in "${platforms[@]}"; do
  IFS='/' read -r os arch <<<"$platform"
  archive_name="xytz-v${VERSION}-${os}-${arch}.tar.gz"

  echo "Building for $os/$arch..."
  GOOS=$os GOARCH=$arch go build \
    -ldflags "-s -w -X github.com/xdagiz/xytz/internal/version.Version=${VERSION}" \
    -o "${TEMP_DIR}/xytz${os:+.$os}${arch:+.$arch}" \
    .

  if [ "$os" = "windows" ]; then
    mv "${TEMP_DIR}/xytz${os:+.$os}${arch:+.$arch}" "${TEMP_DIR}/xytz.exe"
  else
    mv "${TEMP_DIR}/xytz${os:+.$os}${arch:+.$arch}" "${TEMP_DIR}/xytz"
  fi

  tar -czf "${DIST_DIR}/${archive_name}" -C "$TEMP_DIR" .
done

rm -rf "$TEMP_DIR"

echo "Generating checksums..."
cd "$DIST_DIR"
sha256sum ./* >checksums.txt
cd ..

echo "Build complete"
ls -la "$DIST_DIR/"

#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
VERSION=${VERSION:-0.0.0-pre-alpha.1}
DIST="$ROOT/dist/$VERSION"
COMMIT=${COMMIT:-$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || printf 'unknown')}
BUILD_DATE=${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}

rm -rf "$DIST"
mkdir -p "$DIST"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
  -ldflags "-s -w -X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildDate=$BUILD_DATE" \
  -o "$DIST/praetorium-linux-amd64" "$ROOT/cmd/praetorium"
cp "$ROOT/scripts/install.sh" "$DIST/install.sh"
cp "$ROOT/README.md" "$ROOT/LICENSE" "$ROOT/VERSION.md" "$DIST/"
(cd "$DIST" && shasum -a 256 praetorium-linux-amd64 install.sh > SHA256SUMS)
printf '%s\n' "Release artifacts written to $DIST"

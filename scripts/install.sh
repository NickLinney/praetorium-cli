#!/bin/sh
set -eu

VERSION=0.0.0-pre-alpha.1
INSTALL_DIR=/usr/local/bin
BASE_URL=${PRAETORIUM_RELEASE_BASE_URL:-}

usage() {
  printf '%s\n' 'Usage: install.sh [--version VERSION] [--base-url URL] [--install-dir PATH]'
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version) VERSION=$2; shift 2 ;;
    --base-url) BASE_URL=$2; shift 2 ;;
    --install-dir) INSTALL_DIR=$2; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) printf 'install.sh: unknown option %s\n' "$1" >&2; exit 2 ;;
  esac
done

case "$(uname -s):$(uname -m)" in
  Linux:x86_64|Linux:amd64) ;;
  *) printf '%s\n' 'install.sh: certified target is Debian GNU/Linux amd64' >&2; exit 3 ;;
esac

if [ -z "$BASE_URL" ]; then
  printf '%s\n' '--base-url or PRAETORIUM_RELEASE_BASE_URL is required' >&2
  exit 2
fi

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT HUP INT TERM
binary_url=${BASE_URL%/}/praetorium-linux-amd64
checksums_url=${BASE_URL%/}/SHA256SUMS
curl --fail --location --silent --show-error "$binary_url" -o "$tmpdir/praetorium-linux-amd64"
curl --fail --location --silent --show-error "$checksums_url" -o "$tmpdir/SHA256SUMS"
(cd "$tmpdir" && grep '  praetorium-linux-amd64$' SHA256SUMS | sha256sum -c -)
mkdir -p "$INSTALL_DIR"
if [ -w "$INSTALL_DIR" ]; then
  install -m 0755 "$tmpdir/praetorium-linux-amd64" "$INSTALL_DIR/praetorium"
else
  sudo install -m 0755 "$tmpdir/praetorium-linux-amd64" "$INSTALL_DIR/praetorium"
fi
printf 'Installed praetorium %s to %s/praetorium\n' "$VERSION" "$INSTALL_DIR"

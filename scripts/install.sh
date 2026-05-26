#!/usr/bin/env bash
#
# hibot CLI installer.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/volcengine/hiagent-go-sdk/main/scripts/install.sh | sh
#
# Environment overrides:
#   HIBOT_VERSION   Version to install: 1.0.0, v1.0.0, or cmd/hibot/v1.0.0
#   HIBOT_PREFIX    Install prefix (default: /usr/local)
#   HIBOT_BIN_DIR   Binary destination directory (default: $HIBOT_PREFIX/bin)
#   HIBOT_REPO      GitHub repo (default: volcengine/hiagent-go-sdk)
#
set -euo pipefail

REPO="${HIBOT_REPO:-volcengine/hiagent-go-sdk}"
PREFIX="${HIBOT_PREFIX:-/usr/local}"
BIN_DIR="${HIBOT_BIN_DIR:-$PREFIX/bin}"
VERSION="${HIBOT_VERSION:-}"

err() {
  echo "[hibot-install] error: $*" >&2
  exit 1
}

info() {
  echo "[hibot-install] $*"
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || err "required command '$1' not found in PATH"
}

urlencode_tag() {
  printf '%s' "$1" | sed 's#/#%2F#g'
}

need_cmd curl
need_cmd grep
need_cmd head
need_cmd install
need_cmd mkdir
need_cmd sed
need_cmd tar
need_cmd uname

OS_RAW="$(uname -s)"
ARCH_RAW="$(uname -m)"

case "$OS_RAW" in
  Linux) GOOS="linux" ;;
  Darwin) GOOS="darwin" ;;
  MINGW*|MSYS*|CYGWIN*) err "Windows shell detected; install on Windows by downloading the .zip release manually." ;;
  *) err "unsupported OS: $OS_RAW" ;;
esac

case "$ARCH_RAW" in
  x86_64|amd64) GOARCH="amd64" ;;
  arm64|aarch64) GOARCH="arm64" ;;
  *) err "unsupported architecture: $ARCH_RAW" ;;
esac

if [ -z "$VERSION" ]; then
  info "resolving latest cmd/hibot release tag from github.com/$REPO ..."
  VERSION="$(
    curl -fsSL "https://api.github.com/repos/$REPO/releases?per_page=100" \
      | grep -E '"tag_name":[[:space:]]*"cmd/hibot/v[^"]+"' \
      | head -n 1 \
      | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/' \
      || true
  )"
  [ -n "$VERSION" ] || err "could not determine latest cmd/hibot release"
fi

case "$VERSION" in
  cmd/hibot/v*)
    TAG="$VERSION"
    BARE="${VERSION#cmd/hibot/v}"
    ;;
  cmd/hibot/*)
    TAG="$VERSION"
    BARE="${VERSION#cmd/hibot/}"
    BARE="${BARE#v}"
    ;;
  v*)
    TAG="cmd/hibot/$VERSION"
    BARE="${VERSION#v}"
    ;;
  *)
    TAG="cmd/hibot/v$VERSION"
    BARE="$VERSION"
    ;;
esac

TAG_PATH="$(urlencode_tag "$TAG")"
ARCHIVE="hibot_${BARE}_${GOOS}_${GOARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$TAG_PATH/$ARCHIVE"
SUMS_URL="https://github.com/$REPO/releases/download/$TAG_PATH/checksums.txt"

TMP="$(mktemp -d -t hibot-install.XXXXXX)"
trap 'rm -rf "$TMP"' EXIT

info "downloading $URL"
curl -fSL --retry 3 -o "$TMP/$ARCHIVE" "$URL" || err "download failed"

if curl -fsSL -o "$TMP/checksums.txt" "$SUMS_URL"; then
  info "verifying SHA-256 checksum"
  if command -v sha256sum >/dev/null 2>&1; then
    (cd "$TMP" && grep " $ARCHIVE\$" checksums.txt | sha256sum -c -) \
      || err "checksum mismatch"
  elif command -v shasum >/dev/null 2>&1; then
    (cd "$TMP" && grep " $ARCHIVE\$" checksums.txt | shasum -a 256 -c -) \
      || err "checksum mismatch"
  else
    info "warning: no sha256sum/shasum found, skipping checksum verification"
  fi
else
  info "warning: checksums.txt not available, skipping verification"
fi

tar -xzf "$TMP/$ARCHIVE" -C "$TMP"

mkdir -p "$BIN_DIR" 2>/dev/null || {
  info "cannot create $BIN_DIR without sudo; retrying with sudo"
  sudo mkdir -p "$BIN_DIR"
}

if [ -w "$BIN_DIR" ]; then
  install -m 0755 "$TMP/hibot" "$BIN_DIR/hibot"
else
  info "$BIN_DIR is not writable; using sudo"
  sudo install -m 0755 "$TMP/hibot" "$BIN_DIR/hibot"
fi

info "installed: $BIN_DIR/hibot"
"$BIN_DIR/hibot" version || true

case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *) info "warning: $BIN_DIR is not in your PATH; add it to your shell profile." ;;
esac

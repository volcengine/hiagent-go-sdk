#!/usr/bin/env bash
#
# Prepare the independent cmd/hibot Go module for GoReleaser.
#
# Local snapshots keep the ../../hibot replace so they can build from the
# working tree. Real cmd/hibot/v* releases remove that replace and resolve the
# SDK submodule from the matching hibot/v* tag.
set -euo pipefail

GO_MOD="cmd/hibot/go.mod"
SDK_MODULE="github.com/volcengine/hiagent-go-sdk/hibot"
SDK_REPLACE="../../hibot"
TAG="${GORELEASER_CURRENT_TAG:-}"

if [ ! -f "$GO_MOD" ]; then
  echo "error: $GO_MOD not found (run from repo root)" >&2
  exit 1
fi

if [ -z "$TAG" ] || [ "${TAG#cmd/hibot/v}" = "$TAG" ]; then
  echo "[prepare-cli-module] local snapshot; keeping $SDK_MODULE => $SDK_REPLACE"
  (cd cmd/hibot && go mod download)
  exit 0
fi

VERSION="${TAG#cmd/hibot/v}"
SDK_TAG="hibot/v$VERSION"

python3 - "$GO_MOD" "$SDK_MODULE" "$SDK_REPLACE" <<'PY'
import io
import re
import sys

path, module, replacement = sys.argv[1:]
with io.open(path, encoding="utf-8") as f:
    text = f.read()

pattern = re.compile(
    r"\n(?:// [^\n]*\n)*replace\s+"
    + re.escape(module)
    + r"\s*=>\s*"
    + re.escape(replacement)
    + r"\s*\n",
    re.MULTILINE,
)
new_text, _ = pattern.subn("\n", text)
new_text = re.sub(r"\n{3,}\Z", "\n\n", new_text)

with io.open(path, "w", encoding="utf-8") as f:
    f.write(new_text)
PY

echo "[prepare-cli-module] resolving $SDK_MODULE@$SDK_TAG"
(cd cmd/hibot && go get "$SDK_MODULE@$SDK_TAG" && go mod tidy && go mod download)

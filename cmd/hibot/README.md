# Hibot CLI

Command-line interface for the Hibot Managed Agent platform. `hibot` is a thin
wrapper around the official [Go SDK](../../hibot) and exposes the same TOP API
surface — Agents, Sessions, Skills, MCPs, Resources, Prompts, Environments,
Models and Uploads — plus a streaming `chat` command.

## Install

Pick the method that fits your environment.

### 1. One-line installer (Linux & macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/volcengine/hibot-sdk/main/scripts/install.sh | bash
```

Pin a specific version with `HIBOT_VERSION=cmd/hibot/v0.1.0`. Override the install
prefix with `HIBOT_PREFIX=$HOME/.local` (no `sudo` required) or
`HIBOT_BIN_DIR=/path/to/bin`.

### 2. Homebrew (macOS / Linux)

```bash
brew install volcengine/tap/hibot
brew upgrade hibot
```

### 3. Pre-built binaries

Download the matching archive from the
[GitHub Releases page](https://github.com/volcengine/hiagent-go-sdk/releases),
extract it, and place the `hibot` binary somewhere on your `PATH`. Each
release ships with `checksums.txt` (SHA-256) for verification.

### 4. `go install` (Go ≥ 1.22)

```bash
go install github.com/volcengine/hiagent-go-sdk/cmd/hibot@latest
# or pin a tag:
go install github.com/volcengine/hiagent-go-sdk/cmd/hibot@cmd/hibot/v0.1.0
```

### Verify

```bash
hibot version
# hibot 0.1.0 darwin/arm64 (commit ..., built ...)
```

## Build from source (development only)

The repository uses a `go.work` file so the CLI can be developed against the
local Go SDK source. From the repo root:

```bash
git clone https://github.com/volcengine/hiagent-go-sdk.git
cd hiagent-go-sdk/cmd/hibot
go build -o bin/hibot ./
./bin/hibot --help
```

> The `replace github.com/volcengine/hiagent-go-sdk/hibot => ../../hibot` directive in
> `cmd/hibot/go.mod` is for local development only. Remove it before publishing
> a tagged CLI module.

## Configure

Configuration precedence (highest wins):

1. CLI flags (`--endpoint`, `--ak`, `--sk`, `--workspace-id`, ...)
2. Environment variables (`HIBOT_ENDPOINT`, `HIBOT_AK`, `HIBOT_SK`,
   `HIBOT_WORKSPACE_ID`, `HIBOT_REGION`, `HIBOT_SERVER_SERVICE`,
   `HIBOT_GATEWAY_SERVICE`, `HIBOT_MODEL_SERVICE`, `HIBOT_UP_SERVICE`)
3. Config file at `$HOME/.hibot/config.yaml` (override path with
   `--config-file`).

Bootstrap a config file from current flags:

```
hibot --endpoint=https://open.volcengineapi.com \
      --ak=AK --sk=SK --workspace-id=ws-xxx --region=cn-beijing \
      config init
```

Other helpers:

```
hibot config view              # show resolved values (AK/SK masked)
hibot config set workspace_id ws-xxx
```

## Output

`-o/--output` selects the output format: `table` (default), `json`, or `yaml`.

```
hibot agents list
hibot agents list -o json
hibot agents list -o yaml
```

`-v/--verbose` enables extra event logs in the streaming `chat` command.

## Commands

```
hibot version
hibot config init|view|set
hibot agents create|list|get|update|delete
hibot sessions create|list|get|delete|archive
hibot sessions messages list|get|inject
hibot chat <session-id> [--input ... | stdin] [--stream]
hibot models list|get|create|delete
hibot models providers list|list-models
hibot skills list|get|delete|upload|versions
hibot mcps list|get|create|delete|test
hibot resources list|get-by-name|create|delete
hibot resources directories list|create|delete
hibot prompts list|create|update|delete
hibot environments list|get|create|delete
hibot uploads blob
```

Many flags accept `@/path/to/file` to read content from a file (e.g.
`--system @prompt.txt`, `--content @body.md`, `--env-vars @env.json`).

## Streaming chat

```
echo "Tell me a joke" | hibot chat sess-123 --stream
hibot chat sess-123 --stream --input "@prompts/run.md"
```

Streaming behaviour:

- `delta` events are written to stdout verbatim (no newlines added).
- `completed` prints `\n[completed message_id=...]` and exits 0.
- `failed` returns a non-zero exit code with the error message.
- Other events (tool start/complete, approvals, run cancel, ...) are silenced
  unless `-v` is passed.

## Exit codes

- `0` — success
- `1` — API error (`hibot.APIError`) or unexpected runtime error
- `2` — user error (missing flag, invalid argument, etc.)

`hibot.APIError` is rendered with `status=<HTTP> request_id=<id> code=<code>
message=<msg>` so you can hand the request_id to support.

## Tests

```
cd cmd/hibot
go vet ./...
go test ./...
```

Tests under `internal/cmd` mock the TOP HTTP layer with `httptest` so they run
fully offline.

## Examples

See [`examples/hibot/README.md`](../../examples/hibot/README.md) for end-to-end recipes.

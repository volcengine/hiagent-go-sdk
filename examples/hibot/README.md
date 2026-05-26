# Hibot Go SDK Examples

These examples are split by customer integration scenario. Each directory is a
standalone `package main` and can be copied into a customer's project.

## Common Environment

All examples read these variables:

```bash
export HIBOT_ENDPOINT="https://<top-host>"
export HIBOT_AK="<access-key>"
export HIBOT_SK="<secret-key>"
export HIBOT_WORKSPACE_ID="<workspace-id>"
export HIBOT_MODEL_ID="doubao-seed-2.0-pro-260215" # optional
```

## Scenarios

| Directory | Scenario | Extra Environment |
| --- | --- | --- |
| `basic_agent` | Create a model-backed Agent, create a Session, send one non-streaming chat request. | none |
| `streaming_chat` | Create an Agent and consume `ChatStreaming` SSE events. | none |
| `peer_session` | Bind a Session to an IM channel (Feishu / WeCom) by passing explicit `Channel` + `PeerKind` + `PeerID`. | `HIBOT_PEER_CHANNEL`, `HIBOT_PEER_KIND`, `HIBOT_PEER_ID` optional |
| `skill_upload` | Upload a local Skill zip, create a Skill version, bind it to an Agent. | `HIBOT_SKILL_FILE` |
| `mcp_agent` | Register a streamable HTTP MCP server and bind it to an Agent. | `HIBOT_GITHUB_MCP_ENDPOINT`, `HIBOT_GITHUB_CREDENTIAL_NAME` optional |
| `resource_agent` | Upload a local resource file, create a Resource, bind it to an Agent. | `HIBOT_RESOURCE_FILE` |
| `comprehensive_managed_agent` | Full private-deployment flow: Prompt, Skill, Resource, MCP, Agent, streaming chat (webchat default). | `HIBOT_SKILL_FILE`, `HIBOT_RESOURCE_FILE`, `HIBOT_GITHUB_MCP_ENDPOINT` |

## Run

```bash
cd hiagent-go-sdk/examples/hibot
go run ./basic_agent
go run ./streaming_chat
go run ./comprehensive_managed_agent
```

## E2E Closed-Loop Fixtures

`../../testdata/hibot/` 下提供了 `examples/e2e` 测试使用的最小 fixture：

| Fixture | 用途 |
| --- | --- |
| `../../testdata/hibot/runbook.md` | 上传为 Resource，文档中嵌入了 `HIBOT-E2E-RESOURCE-TOKEN-2026-05-23`，测试会询问 Agent 复述该 token 来验证 Resource 闭环。 |
| `../../testdata/hibot/skill/SKILL.md` | 由测试动态打包成 zip 上传为 Skill；front matter 的 `name`/`description` 用于通过 server 端 `parseSkillManifest`，触发后 Agent 应返回 `PULSE_OK_E2E`。 |

执行真实环境闭环测试：

```bash
export HIBOT_E2E_TOP_HOST="https://<top-host>"
export HIBOT_E2E_AK="<access-key>"
export HIBOT_E2E_SK="<secret-key>"
export HIBOT_E2E_WORKSPACE="<workspace-id>"
# 可选：HIBOT_E2E_KEEP_AGENT=1 保留 Agent/Resource/Skill，便于事后排障
go test -run TestRealEnvResourceSkillLoop -count=1 -v ./e2e/...
```

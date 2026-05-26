# Hibot Go SDK

Hibot 平台的官方 Go SDK，封装了 Hibot 私有化部署下的 Agent / Session / Skill / MCP / Resource 等核心资源，使开发者能够通过类型安全的 Go API 一站式创建托管 Agent 并发起会话。

> **本 SDK 的一致性原则**：所有可复用资源（Model、Skill、MCP、Resource、Prompt）均由 `client.V1.*` 资源 client 显式创建并持久化；Session 仅引用资源 ID。Agent 一次创建，按 ID 长期复用；不要在请求路径里反复 `Agents.New`。

## 核心概念

| 概念 | 角色 | 对应 Claude Managed Agents |
|---|---|---|
| **Agent** | Model + System Prompt + Skills + MCPs + Resources 的可复用定义 | Agent |
| **Environment** | 运行容器模板（镜像、资源、环境变量） | Environment |
| **Session** | 一次会话上下文，绑定 Agent 与 Peer 身份 | Session |
| **Skill** | 上传到平台的可执行能力包（zip artifact + 版本） | Tools (file-resource skill) |
| **MCP** | 外部 Streamable HTTP / Stdio MCP Server | MCP Servers |
| **Resource** | 知识 / 文档 / 数据集合，挂载到 Agent | File / Resource |
| **Prompt** | 可复用 System Prompt 模板 | System prompt template |

与 Claude SDK 的关键差异：

- 路由分层：CRUD（Agent / Session / Skill / MCP / Resource）走 `hibot-server`；流式 Chat 走 `hibot-gateway`。SDK 内部根据 Action 自动路由，调用方无感知。
- 资源版本化命名：使用 `client.V1.*`（稳定接口），不使用 `Beta` 前缀。
- 鉴权：使用 TOP AK/SK + WorkspaceID（多租户隔离），`TenantID` 由服务端从 AK/SK 解析。
- 流事件：SDK 将 Gateway 多种事件名（`message_delta` / `message.chunk` / `run_completed` 等）归一化为统一三态：`delta` / `completed` / `failed`。

## 安装

```bash
go get github.com/volcengine/hiagent-go-sdk/hibot@latest
```

`go.mod` 模块路径与导入路径均为 `github.com/volcengine/hiagent-go-sdk/hibot`。

## Client 初始化

```go
import (
    "context"

    "github.com/volcengine/hiagent-go-sdk/hibot"
)

ctx := context.Background()

client, err := hibot.NewClient(hibot.Config{
    Endpoint:    "https://<top-host>",
    AccessKey:   "<access-key>",
    SecretKey:   "<secret-key>",
    WorkspaceID: "<workspace-id>",
})
if err != nil {
    panic(err)
}
```

`Config` 字段：

| 字段 | 必填 | 说明 |
|---|---|---|
| `Endpoint` | 是 | TOP 网关地址 |
| `AccessKey` / `SecretKey` | 是 | TOP AK/SK；服务端凭此解析 TenantID |
| `WorkspaceID` | 是 | 工作空间 ID；所有资源在此空间隔离 |
| `Region` | 否 | 默认 `cn-north-1` |
| `HTTPClient` | 否 | 自定义 `*http.Client`；默认 30s 超时 |
| `ServerService` / `GatewayService` / `ModelService` / `UpService` | 否 | 私有化部署下覆盖 TOP service 名；默认值分别为 `hibot-server` / `hibot-gateway` / `aigw-server` / `up`，模型相关 Action（`GetModel` / `ListModel` / `ListProvider` / `ListModelProvider` / `GetProvider` / `GetModelProviderCredentialSchema`，API Version `2023-08-01`）通过 `aigw-server` 路由 |

资源 client 全部挂在 `client.V1` 下：`Uploads` / `Environments` / `Models` / `Prompts` / `Resources` / `MCPs` / `Skills` / `Agents` / `Sessions`。

## 选择一个 Model

```go
model, err := client.V1.Models.Get(ctx, hibot.V1ModelGetParams{
    ID: hibot.V1ManagedAgentModelDoubaoSeedPro, // 或填入私有化部署中已注册的 ModelID
})
if err != nil {
    panic(err)
}
fmt.Println(model.ID)
```

## 创建 / 复用 Environment

`Agents.New` 在不传 `EnvID` 时会自动选择默认 Environment。需要自定义运行容器时再显式创建：

```go
env, err := client.V1.Environments.New(ctx, hibot.V1EnvironmentNewParams{
    Name:        "my-dev-env",
    ImageType:   "default",
    CPULimit:    "2",
    MemoryLimit: "4Gi",
})
```

## 创建 Agent（先建一次，重复使用）

⚠️ **没有 inline agent 配置**：`Model` / `System` / `Tools` / `Resources` 都挂在 Agent 对象上，Session 只引用 `AgentID`。

最小示例：

```go
agent, err := client.V1.Agents.New(ctx, hibot.V1AgentNewParams{
    Name:   "Coding Assistant",
    Model:  hibot.V1ManagedAgentModelConfigParams{ID: model.ID},
    System: hibot.String("你是一个简洁的 Hibot SDK 示例助手。"),
})
```

带 Skill / MCP / Resource：

```go
agent, err := client.V1.Agents.New(ctx, hibot.V1AgentNewParams{
    Name:   "Comprehensive Agent",
    Model:  hibot.V1ManagedAgentModelConfigParams{ID: model.ID},
    System: hibot.String(prompt.Content),
    Tools: []hibot.V1AgentNewParamsToolUnion{
        {OfSkill: &hibot.V1ManagedAgentSkillToolParams{
            Type:           hibot.V1ManagedAgentSkillToolParamsTypeSkill,
            SkillVersionID: skill.ID,
        }},
        {OfMCP: &hibot.V1ManagedAgentMCPToolParams{
            Type: hibot.V1ManagedAgentMCPToolParamsTypeMCP,
            ID:   mcp.ID,
        }},
    },
    Resources: []hibot.V1ManagedAgentResourceRefParams{{ID: resource.ID}},
})
```

更新 / 删除 Agent：`client.V1.Agents.Update` / `Delete`；查询：`Get` / `BatchGet` / `List`。

## 创建 Session

主流程（webchat 单体集成）不需要传 `Peer`，SDK 会自动注入
`Channel=webchat / PeerKind=system / PeerID=AgentID` 兜底，保证服务端
SessionKey 唯一确定：

```go
session, err := client.V1.Sessions.New(ctx, hibot.V1SessionNewParams{
    AgentID: agent.ID,
})
```

需要把会话挂到飞书 / 企微 等 IM 渠道时，再显式传 `Peer`：

```go
session, err := client.V1.Sessions.New(ctx, hibot.V1SessionNewParams{
    AgentID: agent.ID,
    Peer: &hibot.V1SessionPeerParams{
        Channel:  "feishu",   // 留空时维持 webchat
        PeerKind: "user",     // user / group / bot / system
        PeerID:   "ou_xxx",
    },
})
```

会话维度操作：`List` / `Get` / `GetByKey` / `Archive` / `Delete` / `ListMessages` / `GetMessage` / `InjectMessage`。

## 发送一条聊天（非流式）

```go
message, err := client.V1.Sessions.Chat(ctx, session.ID, hibot.V1SessionChatParams{
    Input: "请用一句话介绍 Hibot Agent。",
})
if err != nil {
    panic(err)
}
fmt.Printf("message_id=%s content=%s\n", message.ID, message.Content)
```

`Chat` 内部消费 SSE 流到 `completed` 后返回最终 `V1Message`。

## 流式聊天 (SSE)

```go
stream := client.V1.Sessions.ChatStreaming(ctx, session.ID, hibot.V1SessionChatParams{
    Input: "请流式输出一个三步排障计划。",
})
defer stream.Close()

for stream.Next() {
    event := stream.Current()
    switch event.Type {
    case hibot.V1SessionChatEventDelta:
        fmt.Print(event.Delta.Text)
    case hibot.V1SessionChatEventCompleted:
        fmt.Println("\n[completed]")
    case hibot.V1SessionChatEventFailed:
        log.Fatalf("chat failed: %s", event.Error.Message)
    }
}
if err := stream.Err(); err != nil {
    panic(err)
}
final, _ := stream.FinalMessage()
fmt.Println(final.ID)
```

事件类型常量（`hibot.V1SessionChatEvent*`）：

- `Delta`：增量文本（`event.Delta.Text`）
- `Completed`：流结束，含完整 `V1Message`
- `Failed`：流失败，`event.Error` 为详情
- `RunCancelling` / `RunCancelled`：运行被取消
- `ApprovalRequest` / `ApprovalResponded`：HITL 审批节点
- `ToolStart` / `ToolComplete`：工具调用观测事件

如需把整段流累加成最终消息，可直接调用 `stream.Accumulate()`。

## 上传 Skill 并绑定到 Agent

```go
f, _ := os.Open("./my-skill.zip")
defer f.Close()

blob, err := client.V1.Uploads.UploadBlob(ctx, hibot.V1UploadBlobParams{
    Filename:    "my-skill.zip",
    ContentType: "application/zip",
}, f)

enabled := true
skill, err := client.V1.Skills.New(ctx, hibot.V1SkillNewParams{
    Name:        "my-skill",
    Description: "Local skill uploaded by SDK example.",
    BlobID:      blob.BlobID,
    Enabled:     &enabled,
    Version:     "1.0.0",
})
// 然后把 skill.ID 作为 SkillVersionID 加到 V1AgentNewParams.Tools
```

> 私有化部署下当前 `UploadBlob` 仍依赖公有云 artifact 入口，待服务端补齐私有化 artifact 通道后即可端到端落地。

## 注册 MCP Server

```go
mcp, err := client.V1.MCPs.New(ctx, hibot.V1MCPNewParams{
    Name:      "github-mcp",
    Transport: hibot.V1MCPTransportStreamableHTTP,
    Endpoint:  "https://api.githubcopilot.com/mcp/",
    Credential: &hibot.V1CredentialRefParams{
        Name: "github-token", // 凭据注册在凭据中心
    },
})
```

## 上传知识 Resource

```go
f, _ := os.Open("./handbook.pdf")
defer f.Close()

blob, _ := client.V1.Uploads.UploadBlob(ctx, hibot.V1UploadBlobParams{
    Filename:    "handbook.pdf",
    ContentType: "application/pdf",
}, f)

resource, err := client.V1.Resources.New(ctx, hibot.V1ResourceNewParams{
    Name:   "engineering-handbook",
    Type:   hibot.V1ResourceTypeDocumentCollection,
    BlobID: blob.BlobID,
})
```

`V1ManagedAgentResourceRefParams` 同时支持单文件 (`ID`) 与目录 (`DirectoryID`) 维度绑定。

## 端到端综合示例（Prompt + Skill + Resource + MCP + Agent + Session + 流式 Chat）

完整可运行代码见 [examples/hibot/comprehensive_managed_agent/main.go](../examples/hibot/comprehensive_managed_agent/main.go)。

## Examples

`examples/` 下每个目录都是独立 `package main`，直接体现一个集成场景：

| 目录 | 场景 |
|---|---|
| [basic_agent](../examples/hibot/basic_agent/main.go) | 创建 Agent + Session，单次非流式 Chat |
| [streaming_chat](../examples/hibot/streaming_chat/main.go) | 消费 SSE 流式输出 |
| [peer_session](../examples/hibot/peer_session/main.go) | 把会话挂到飞书 / 企微 IM 渠道（显式 Channel/PeerKind/PeerID） |
| [skill_upload](../examples/hibot/skill_upload/main.go) | 上传本地 Skill zip 并绑定 Agent |
| [mcp_agent](../examples/hibot/mcp_agent/main.go) | 注册 Streamable HTTP MCP 并绑定 Agent |
| [resource_agent](../examples/hibot/resource_agent/main.go) | 上传 Resource 并绑定 Agent |
| [comprehensive_managed_agent](../examples/hibot/comprehensive_managed_agent/main.go) | 私有化全要素：Prompt + Skill + Resource + MCP + Agent + 流式 Chat |
| [e2e](../examples/hibot/e2e/e2e_test.go) | 真实环境与 Mock 服务端 E2E 测试 |

通用环境变量：

```bash
export HIBOT_ENDPOINT="https://<top-host>"
export HIBOT_AK="<access-key>"
export HIBOT_SK="<secret-key>"
export HIBOT_WORKSPACE_ID="<workspace-id>"
export HIBOT_MODEL_ID="doubao-seed-2.0-pro-260215"   # 可选
```

运行：

```bash
cd hiagent-go-sdk/examples/hibot
go run ./basic_agent
go run ./streaming_chat
go run ./comprehensive_managed_agent
```

## 错误处理

服务端非 2xx 响应会被包装为 `hibot.APIError`：

```go
if err != nil {
    var apiErr *hibot.APIError
    if errors.As(err, &apiErr) {
        log.Printf("hibot API error: status=%d message=%s", apiErr.StatusCode, apiErr.Message)
    }
}
```

## 设计原则

- **无状态客户端**：SDK 不在内部维护任何 session→agent 的隐式 map，线程安全。
- **Fail-fast**：资源不存在 / 必填缺失立即返回错误，绝不返回伪造空对象。
- **IDL 对齐**：手写结构体严格对齐服务端 Thrift IDL（如 `AgentSkillInput.SkillVersionID` 落到 `ID`、`MCP.Endpoint` 落到 `URL`）；调用方无需关心字段映射。
- **WorkspaceID 仅在 ActionRequest 顶层透传**，不会被隐式注入到 Payload 内部。
- **私有化兼容**：SSE 事件名、Service 名、UP / Server / Gateway / Model service 全部可通过 `Config` 覆盖。

## 相关文档

- Examples 入口：[examples/hibot/README.md](../examples/hibot/README.md)

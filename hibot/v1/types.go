package v1

const (
	V1ManagedAgentModelDoubaoSeedPro = "doubao-seed-2.0-pro-260215"

	V1ResourceTypeDocumentCollection = "document_collection"

	V1MCPTransportStreamableHTTP = "streamable_http"

	V1ManagedAgentSkillToolParamsTypeSkill = "skill"
	V1ManagedAgentMCPToolParamsTypeMCP     = "mcp"

	// SSE 事件名严格对齐 hibot-gateway (internal/components/gateway/ssehub.go)。
	V1SessionChatEventDelta             = "delta"
	V1SessionChatEventCompleted         = "completed"
	V1SessionChatEventFailed            = "failed"
	V1SessionChatEventRunCancelling     = "run_cancelling"
	V1SessionChatEventRunCancelled      = "run_cancelled"
	V1SessionChatEventApprovalRequest   = "approval_request"
	V1SessionChatEventApprovalResponded = "approval_responded"
	V1SessionChatEventToolStart         = "tool_start"
	V1SessionChatEventToolComplete      = "tool_complete"
)

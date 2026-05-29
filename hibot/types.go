package hibot

import (
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/response"
	v1 "github.com/volcengine/hiagent-go-sdk/hibot/v1"
)

const (
	V1ManagedAgentModelDoubaoSeedPro = v1.V1ManagedAgentModelDoubaoSeedPro

	V1ResourceTypeDocumentCollection = v1.V1ResourceTypeDocumentCollection

	V1MCPTransportStreamableHTTP = v1.V1MCPTransportStreamableHTTP

	V1ManagedAgentSkillToolParamsTypeSkill = v1.V1ManagedAgentSkillToolParamsTypeSkill
	V1ManagedAgentMCPToolParamsTypeMCP     = v1.V1ManagedAgentMCPToolParamsTypeMCP

	V1SessionChatEventDelta             = v1.V1SessionChatEventDelta
	V1SessionChatEventCompleted         = v1.V1SessionChatEventCompleted
	V1SessionChatEventFailed            = v1.V1SessionChatEventFailed
	V1SessionChatEventRunCancelling     = v1.V1SessionChatEventRunCancelling
	V1SessionChatEventRunCancelled      = v1.V1SessionChatEventRunCancelled
	V1SessionChatEventApprovalRequest   = v1.V1SessionChatEventApprovalRequest
	V1SessionChatEventApprovalResponded = v1.V1SessionChatEventApprovalResponded
	V1SessionChatEventToolStart         = v1.V1SessionChatEventToolStart
	V1SessionChatEventToolComplete      = v1.V1SessionChatEventToolComplete
)

type APIError = response.APIError

type V1 = v1.Client

type UploadsService = v1.UploadsService
type EnvironmentsService = v1.EnvironmentsService
type ModelsService = v1.ModelsService
type PromptsService = v1.PromptsService
type ResourcesService = v1.ResourcesService
type MCPsService = v1.MCPsService
type SkillsService = v1.SkillsService
type AgentsService = v1.AgentsService
type SessionsService = v1.SessionsService

type V1Model = v1.V1Model
type V1UploadBlob = v1.V1UploadBlob
type V1Prompt = v1.V1Prompt
type V1Resource = v1.V1Resource
type V1MCP = v1.V1MCP
type V1SkillVersion = v1.V1SkillVersion
type V1Agent = v1.V1Agent
type V1Environment = v1.V1Environment
type V1Session = v1.V1Session
type V1Message = v1.V1Message
type V1MessageFile = v1.V1MessageFile

type V1ModelGetParams = v1.V1ModelGetParams
type V1ModelListParams = v1.V1ModelListParams
type V1ModelList = v1.V1ModelList
type V1UploadBlobParams = v1.V1UploadBlobParams
type V1PromptNewParams = v1.V1PromptNewParams
type V1ResourceNewParams = v1.V1ResourceNewParams
type V1MCPNewParams = v1.V1MCPNewParams
type V1MCPResolveParams = v1.V1MCPResolveParams
type V1MCPUpdateParams = v1.V1MCPUpdateParams
type V1MCPCredentialInputParams = v1.V1MCPCredentialInputParams
type V1SkillCredentialInputParams = v1.V1SkillCredentialInputParams
type V1CredentialSecretInputParams = v1.V1CredentialSecretInputParams
type V1SkillResolveVersionParams = v1.V1SkillResolveVersionParams
type V1SkillNewParams = v1.V1SkillNewParams
type V1AgentNewParams = v1.V1AgentNewParams
type V1AgentListParams = v1.V1AgentListParams
type V1EnvironmentNewParams = v1.V1EnvironmentNewParams
type V1EnvironmentListParams = v1.V1EnvironmentListParams
type V1ManagedAgentModelConfigParams = v1.V1ManagedAgentModelConfigParams
type V1AgentNewParamsToolUnion = v1.V1AgentNewParamsToolUnion
type V1ManagedAgentSkillToolParams = v1.V1ManagedAgentSkillToolParams
type V1ManagedAgentMCPToolParams = v1.V1ManagedAgentMCPToolParams
type V1ManagedAgentResourceRefParams = v1.V1ManagedAgentResourceRefParams
type V1SessionNewParams = v1.V1SessionNewParams
type V1SessionPeerParams = v1.V1SessionPeerParams
type V1SessionChatParams = v1.V1SessionChatParams
type V1SessionChatEvent = v1.V1SessionChatEvent
type V1SessionTextDelta = v1.V1SessionTextDelta
type V1SessionChatError = v1.V1SessionChatError
type V1SessionChatStream = v1.V1SessionChatStream

type V1ResourceDeleteParams = v1.V1ResourceDeleteParams
type V1SkillDeleteParams = v1.V1SkillDeleteParams
type V1AgentDeleteParams = v1.V1AgentDeleteParams

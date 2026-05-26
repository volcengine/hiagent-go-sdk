package v1

type V1Agent struct {
	ID           string                `json:"ID"`
	WorkspaceID  string                `json:"WorkspaceID,omitempty"`
	Name         string                `json:"Name,omitempty"`
	Description  string                `json:"Description,omitempty"`
	ModelID      string                `json:"ModelID,omitempty"`
	EnvID        string                `json:"EnvID,omitempty"`
	SystemPrompt string                `json:"SystemPrompt,omitempty"`
	Skills       []V1AgentSkillBinding `json:"Skills,omitempty"`
	MCPs         []V1AgentMCPBinding   `json:"MCPs,omitempty"`
	ResourceIDs  []string              `json:"ResourceIDs,omitempty"`
	CreatedAt    string                `json:"CreatedAt,omitempty"`
	UpdatedAt    string                `json:"UpdatedAt,omitempty"`
	CreatedBy    string                `json:"CreatedBy,omitempty"`
	UpdatedBy    string                `json:"UpdatedBy,omitempty"`
}

type V1AgentSkillBinding struct {
	ID      string `json:"ID"`
	Enabled *bool  `json:"Enabled,omitempty"`
}

type V1AgentMCPBinding struct {
	ID            string   `json:"ID"`
	Enabled       bool     `json:"Enabled"`
	ToolAllowlist []string `json:"ToolAllowlist,omitempty"`
	ToolDenylist  []string `json:"ToolDenylist,omitempty"`
}

type V1AgentNewParams struct {
	WorkspaceID string                            `json:"WorkspaceID,omitempty"`
	Name        string                            `json:"Name,omitempty"`
	EnvID       string                            `json:"EnvID,omitempty"`
	Model       V1ManagedAgentModelConfigParams   `json:"-"`
	System      *string                           `json:"SystemPrompt,omitempty"`
	Tools       []V1AgentNewParamsToolUnion       `json:"-"`
	Resources   []V1ManagedAgentResourceRefParams `json:"-"`
}

type V1AgentNewParamsToolUnion struct {
	OfSkill *V1ManagedAgentSkillToolParams `json:"-"`
	OfMCP   *V1ManagedAgentMCPToolParams   `json:"-"`
}

type V1ManagedAgentSkillToolParams struct {
	Type           string `json:"Type,omitempty"`
	SkillVersionID string `json:"SkillVersionID,omitempty"`
}

type V1ManagedAgentMCPToolParams struct {
	Type string `json:"Type,omitempty"`
	ID   string `json:"ID,omitempty"`
}

type V1AgentListParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	Keyword     string `json:"Keyword,omitempty"`
}

type V1AgentGetParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	AgentID     string `json:"AgentID,omitempty"`
}

type V1AgentBatchGetParams struct {
	WorkspaceID string   `json:"WorkspaceID,omitempty"`
	AgentIDs    []string `json:"AgentIDs"`
}

type V1AgentUpdateParams struct {
	WorkspaceID    string                            `json:"WorkspaceID,omitempty"`
	AgentID        string                            `json:"AgentID,omitempty"`
	Description    *string                           `json:"-"`
	ModelID        *string                           `json:"-"`
	EnvID          *string                           `json:"-"`
	System         *string                           `json:"-"`
	Skills         []V1ManagedAgentSkillToolParams   `json:"-"`
	MCPs           []V1ManagedAgentMCPToolParams     `json:"-"`
	Resources      []V1ManagedAgentResourceRefParams `json:"-"`
	ResetResources bool                              `json:"-"`
}

type V1AgentDeleteParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	AgentID     string `json:"AgentID,omitempty"`
}

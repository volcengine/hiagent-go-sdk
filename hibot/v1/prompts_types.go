package v1

type V1Prompt struct {
	ID        string `json:"ID"`
	Name      string `json:"Name"`
	Content   string `json:"SystemPrompt"`
	CreatedAt string `json:"CreatedAt,omitempty"`
	UpdatedAt string `json:"UpdatedAt,omitempty"`
	CreatedBy string `json:"CreatedBy,omitempty"`
	UpdatedBy string `json:"UpdatedBy,omitempty"`
}

type V1PromptNewParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	Name        string `json:"Name,omitempty"`
	Content     string `json:"SystemPrompt,omitempty"`
}

type V1PromptListParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
}

type V1PromptUpdateParams struct {
	WorkspaceID string  `json:"WorkspaceID,omitempty"`
	ID          string  `json:"ID,omitempty"`
	Name        *string `json:"-"`
	Content     *string `json:"-"`
}

type V1PromptDeleteParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	ID          string `json:"ID,omitempty"`
}

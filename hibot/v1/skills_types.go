package v1

type V1Skill struct {
	ID                   string `json:"ID"`
	SkillID              string `json:"SkillID,omitempty"`
	Name                 string `json:"Name,omitempty"`
	Description          string `json:"Description,omitempty"`
	Source               string `json:"Source,omitempty"`
	Version              string `json:"Version,omitempty"`
	ArtifactID           string `json:"ArtifactID,omitempty"`
	Enabled              bool   `json:"Enabled,omitempty"`
	CredentialProviderID string `json:"CredentialProviderID,omitempty"`
	SlugID               string `json:"SlugID,omitempty"`
	CreatedAt            string `json:"CreatedAt,omitempty"`
	UpdatedAt            string `json:"UpdatedAt,omitempty"`
}

type V1SkillVersion struct {
	ID                   string `json:"ID"`
	SkillID              string `json:"SkillID,omitempty"`
	Name                 string `json:"Name,omitempty"`
	Version              string `json:"Version,omitempty"`
	Description          string `json:"Description,omitempty"`
	Source               string `json:"Source,omitempty"`
	ArtifactID           string `json:"ArtifactID,omitempty"`
	Enabled              bool   `json:"Enabled,omitempty"`
	CredentialProviderID string `json:"CredentialProviderID,omitempty"`
	SlugID               string `json:"SlugID,omitempty"`
	CreatedAt            string `json:"CreatedAt,omitempty"`
	Constraint           string `json:"-"`
}

type V1SkillResolveVersionParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	ID          string `json:"ID,omitempty"`
	Name        string `json:"Name,omitempty"`
	Constraint  string `json:"Constraint,omitempty"`
}

type V1SkillNewParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	SkillID     string `json:"SkillID,omitempty"`
	Name        string `json:"Name,omitempty"`
	Description string `json:"Description,omitempty"`
	Source      string `json:"Source,omitempty"`
	BlobID      string `json:"BlobID,omitempty"`
	Enabled     *bool  `json:"Enabled,omitempty"`
	Version     string `json:"Version,omitempty"`
	SlugID      string `json:"SlugID,omitempty"`
}

type V1SkillListParams struct {
	WorkspaceID string       `json:"WorkspaceID,omitempty"`
	Keyword     string       `json:"Keyword,omitempty"`
	Source      string       `json:"Source,omitempty"`
	Name        string       `json:"Name,omitempty"`
	SlugID      string       `json:"SlugID,omitempty"`
	Page        *V1PageInput `json:"Page,omitempty"`
}

type V1SkillGetParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	ID          string `json:"ID,omitempty"`
	SkillID     string `json:"SkillID,omitempty"`
	Version     string `json:"Version,omitempty"`
}

type V1SkillUpdateParams struct {
	WorkspaceID string  `json:"WorkspaceID,omitempty"`
	ID          string  `json:"ID,omitempty"`
	SkillID     string  `json:"SkillID,omitempty"`
	Version     string  `json:"Version,omitempty"`
	Description *string `json:"-"`
	Source      *string `json:"-"`
	ArtifactID  *string `json:"-"`
	Enabled     *bool   `json:"-"`
	NewVersion  *string `json:"-"`
	SlugID      *string `json:"-"`
}

type V1SkillDeleteParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	ID          string `json:"ID,omitempty"`
	SkillID     string `json:"SkillID,omitempty"`
	Version     string `json:"Version,omitempty"`
}

type V1SkillVersionListParams struct {
	WorkspaceID string       `json:"WorkspaceID,omitempty"`
	SkillID     string       `json:"SkillID,omitempty"`
	SortBy      string       `json:"SortBy,omitempty"`
	SortOrder   string       `json:"SortOrder,omitempty"`
	Page        *V1PageInput `json:"Page,omitempty"`
}

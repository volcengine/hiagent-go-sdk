package v1

type V1Resource struct {
	ID          string `json:"ID"`
	Name        string `json:"Name,omitempty"`
	Type        string `json:"Type,omitempty"`
	ArtifactID  string `json:"ArtifactID,omitempty"`
	Size        int64  `json:"Size,omitempty"`
	Extension   string `json:"Extension,omitempty"`
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	DirectoryID string `json:"DirectoryID,omitempty"`
	CreatedAt   string `json:"CreatedAt,omitempty"`
	UpdatedAt   string `json:"UpdatedAt,omitempty"`
	CreatedBy   string `json:"CreatedBy,omitempty"`
	UpdatedBy   string `json:"UpdatedBy,omitempty"`
}

type V1ResourceNewParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	Name        string `json:"Name,omitempty"`
	Type        string `json:"Type,omitempty"`
	// BlobID 来自先前 Uploads.NewBlob 上传产物的 ArtifactID/BlobID。
	BlobID string `json:"BlobID,omitempty"`
	// DirectoryID 可选；不传则放在工作空间根目录。
	DirectoryID string `json:"DirectoryID,omitempty"`
}

type V1ResourceListParams struct {
	WorkspaceID string       `json:"WorkspaceID,omitempty"`
	DirectoryID string       `json:"DirectoryID,omitempty"`
	Name        string       `json:"Name,omitempty"`
	Page        *V1PageInput `json:"Page,omitempty"`
}

type V1ResourceUpdateParams struct {
	WorkspaceID string  `json:"WorkspaceID,omitempty"`
	ResourceID  string  `json:"ResourceID,omitempty"`
	Name        string  `json:"Name,omitempty"`
	DirectoryID *string `json:"-"`
}

type V1ResourceDeleteParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	ResourceID  string `json:"ResourceID,omitempty"`
	DirectoryID string `json:"DirectoryID,omitempty"`
}

type V1ResourceGetByNameParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	Name        string `json:"Name,omitempty"`
	DirectoryID string `json:"DirectoryID,omitempty"`
}

type V1ResourceBatchGetParams struct {
	WorkspaceID string   `json:"WorkspaceID,omitempty"`
	IDs         []string `json:"IDs,omitempty"`
}

type V1ResourceList struct {
	Items []V1Resource `json:"Items"`
	Page  *V1Page      `json:"Page,omitempty"`
}

type V1ManagedAgentResourceRefParams struct {
	// ID 对应单个 Resource ID。
	ID string `json:"ID,omitempty"`
	// DirectoryID 对应目录 ID（绑定整个目录下资源）。
	DirectoryID string `json:"DirectoryID,omitempty"`
}

type V1Directory struct {
	ID            string `json:"ID"`
	Name          string `json:"Name,omitempty"`
	WorkspaceID   string `json:"WorkspaceID,omitempty"`
	CreatedAt     string `json:"CreatedAt,omitempty"`
	UpdatedAt     string `json:"UpdatedAt,omitempty"`
	CreatedBy     string `json:"CreatedBy,omitempty"`
	UpdatedBy     string `json:"UpdatedBy,omitempty"`
	ResourceCount int64  `json:"ResourceCount,omitempty"`
}

type V1DirectoryNewParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	Name        string `json:"Name,omitempty"`
}

type V1DirectoryListParams struct {
	WorkspaceID string       `json:"WorkspaceID,omitempty"`
	Name        string       `json:"Name,omitempty"`
	Page        *V1PageInput `json:"Page,omitempty"`
}

type V1DirectoryUpdateParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	DirectoryID string `json:"DirectoryID,omitempty"`
	Name        string `json:"Name,omitempty"`
}

type V1DirectoryDeleteParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	DirectoryID string `json:"DirectoryID,omitempty"`
}

type V1DirectoryGetByNameParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	Name        string `json:"Name,omitempty"`
}

type V1DirectoryList struct {
	Items []V1Directory `json:"Items"`
	Page  *V1Page       `json:"Page,omitempty"`
}

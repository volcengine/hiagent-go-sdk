package v1

type V1Session struct {
	ID                 string `json:"ID"`
	AgentID            string `json:"AgentID,omitempty"`
	SessionKey         string `json:"SessionKey,omitempty"`
	Status             string `json:"Status,omitempty"`
	Channel            string `json:"Channel,omitempty"`
	PeerKind           string `json:"PeerKind,omitempty"`
	PeerID             string `json:"PeerID,omitempty"`
	RiskLevel          string `json:"RiskLevel,omitempty"`
	MessageCount       int32  `json:"MessageCount,omitempty"`
	LastMessageAt      string `json:"LastMessageAt,omitempty"`
	LastMessageContent string `json:"LastMessageContent,omitempty"`
	Summary            string `json:"Summary,omitempty"`
	CreatedAt          string `json:"CreatedAt,omitempty"`
	UpdatedAt          string `json:"UpdatedAt,omitempty"`
	ArchivedAt         string `json:"ArchivedAt,omitempty"`
}

type V1Message struct {
	ID         string          `json:"ID"`
	SessionID  string          `json:"SessionID,omitempty"`
	RunID      string          `json:"RunID,omitempty"`
	Role       string          `json:"Role,omitempty"`
	Content    string          `json:"Content,omitempty"`
	Visibility string          `json:"Visibility,omitempty"`
	CreatedAt  string          `json:"CreatedAt,omitempty"`
	Files      []V1MessageFile `json:"Files,omitempty"`
}

type V1MessageFile struct {
	FileID      string `json:"FileID,omitempty"`
	Name        string `json:"Name,omitempty"`
	ContentType string `json:"ContentType,omitempty"`
	URL         string `json:"URL,omitempty"`
	URI         string `json:"URI,omitempty"`
	StoragePath string `json:"StoragePath,omitempty"`
	SizeBytes   int64  `json:"SizeBytes,omitempty"`
}

type V1Page struct {
	PageNum  int32 `json:"PageNum,omitempty"`
	PageSize int32 `json:"PageSize,omitempty"`
	Total    int64 `json:"Total,omitempty"`
}

type V1PageInput struct {
	PageNum  int32 `json:"PageNum,omitempty"`
	PageSize int32 `json:"PageSize,omitempty"`
}

type V1SessionNewParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	AgentID     string `json:"AgentID,omitempty"`
	// Peer 仅在显式指定 IM 渠道（飞书 / 企微 等）时才需要传入。
	// 主流程（webchat 单体集成）不要填，SDK 会注入 webchat / system / AgentID
	// 默认值，让服务端 SessionKey 唯一确定。
	Peer *V1SessionPeerParams `json:"-"`
}

// V1SessionPeerParams 用于在 IM/SaaS 多租户嵌入场景下指定渠道与对端身份。
// 主流程（webchat）创建 Session 时无须填写本结构。
type V1SessionPeerParams struct {
	// Channel 取值如 "feishu" / "wecom" / "slack" 等；留空时由 SDK 维持 "webchat"。
	Channel  string `json:"Channel,omitempty"`
	PeerKind string `json:"PeerKind,omitempty"`
	PeerID   string `json:"PeerID,omitempty"`
}

type V1SessionChatParams struct {
	WorkspaceID     string `json:"WorkspaceID,omitempty"`
	AgentID         string `json:"AgentID,omitempty"`
	Input           string `json:"Content,omitempty"`
	ClientMessageID string `json:"ClientMessageID,omitempty"`
}

type V1SessionChatEvent struct {
	Type      string             `json:"type,omitempty"`
	RequestID string             `json:"request_id,omitempty"`
	Delta     V1SessionTextDelta `json:"delta,omitempty"`
	Error     V1SessionChatError `json:"error,omitempty"`
	Message   *V1Message         `json:"message,omitempty"`
	rawData   string
}

type V1SessionTextDelta struct {
	Text string `json:"text,omitempty"`
}

type V1SessionChatError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type V1SessionListParams struct {
	WorkspaceID string       `json:"WorkspaceID,omitempty"`
	AgentID     string       `json:"AgentID,omitempty"`
	Status      string       `json:"Status,omitempty"`
	Channel     string       `json:"Channel,omitempty"`
	Page        *V1PageInput `json:"Page,omitempty"`
}

type V1SessionGetParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	SessionID   string `json:"SessionID,omitempty"`
}

type V1SessionGetByKeyParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	AgentID     string `json:"AgentID,omitempty"`
	SessionKey  string `json:"SessionKey,omitempty"`
}

type V1SessionArchiveParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	SessionID   string `json:"SessionID,omitempty"`
	Summary     string `json:"-"`
	Consolidate *bool  `json:"-"`
}

type V1SessionDeleteParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	SessionID   string `json:"SessionID,omitempty"`
}

type V1MessageListParams struct {
	WorkspaceID string       `json:"WorkspaceID,omitempty"`
	SessionID   string       `json:"SessionID,omitempty"`
	Visibility  string       `json:"Visibility,omitempty"`
	Page        *V1PageInput `json:"Page,omitempty"`
}

type V1MessageGetParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	SessionID   string `json:"SessionID,omitempty"`
	MessageID   string `json:"MessageID,omitempty"`
}

type V1MessageInjectParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	SessionID   string `json:"SessionID,omitempty"`
	Role        string `json:"-"`
	Content     string `json:"-"`
}

type V1MessageList struct {
	Items []V1Message `json:"Items"`
	Page  *V1Page     `json:"Page,omitempty"`
}

type V1SessionList struct {
	Items []V1Session `json:"Items"`
	Page  *V1Page     `json:"Page,omitempty"`
}

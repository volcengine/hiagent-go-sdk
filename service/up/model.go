package up

// UploadRawRequest ...
type UploadRawRequest struct {
	ID          string  `json:"Id"`
	Sha256      string  `json:"Sha256"`
	Expire      string  `json:"Expire"`
	ContentType string  `json:"ContentType"`
	Filename    *string `json:"Filename"`
}

// UploadRawResponse ...
type UploadRawResponse struct {
	Path   string `json:"Path"`
	Size   int64  `json:"Size"`
	Sha256 string `json:"Sha256"`
}

// DeleteRequest ...
type DeleteRequest struct {
	ID     string `json:"Id"`
	Sha256 string `json:"Sha256"`
}

// DownloadKeyResponse ...
type DownloadKeyResponse struct {
	Key  string `json:"Key"`
	Size int64  `json:"Size"`
}

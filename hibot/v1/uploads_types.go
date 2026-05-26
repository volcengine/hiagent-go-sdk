package v1

type V1UploadBlob struct {
	BlobID string `json:"BlobID"`
}

type V1UploadBlobParams struct {
	Filename    string `json:"-"`
	ContentType string `json:"-"`
}

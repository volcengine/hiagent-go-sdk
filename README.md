# Overview

English | [中文README](README.zh_CN.md)

HiAgent-SDK is the SDK of the HiAgent product from Volcano Engine. Developers can use this SDK to quickly develop
functions and improve development efficiency. HiAgent-SDK provides a complete AI native application development suite,
including a rich set of development components and application example code.

## Architecture

![img.png](img.png)

## Quick Start

```go
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/volcengine/hiagent-go-sdk/service/up"
)

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func main() {
	ctx := context.Background()

	// Get credentials from environment
	ak := os.Getenv("VOLC_ACCESSKEY")
	sk := os.Getenv("VOLC_SECRETKEY")
	uploadEndpoint := os.Getenv("HIAGENT_UP_UPLOAD_ENDPOINT")

	// Create upload client
	client := up.New(uploadEndpoint, ak, sk)

	// 1. Open file to upload
	testFile, err := os.Open("example.txt")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer testFile.Close()

	// Calculate file hash
	fileHash, err := calculateSHA256("example.txt")
	if err != nil {
		log.Fatalf("Failed to calculate SHA256: %v", err)
	}

	// 2. Upload file
	uploadReq := up.UploadRawRequest{
		ID:          strings.ReplaceAll(uuid.New().String(), "-", ""),
		ContentType: "text/plain",
		Expire:      "15h",
		Sha256:      fileHash,
	}

	uploadResp, err := client.UploadRaw(ctx, uploadReq, testFile)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	fmt.Printf("Upload successful. Path: %s, Size: %d\n", uploadResp.Path, uploadResp.Size)

	// 3. Get download key
	downloadKeyResp, err := client.DownloadKey(ctx, uploadResp.Path)
	if err != nil {
		log.Fatalf("Get download key failed: %v", err)
	}
	fmt.Printf("Download key obtained: %s\n", downloadKeyResp.Key)

	// 4. Download file
	downloadBody, err := client.Download(ctx, uploadResp.Path, downloadKeyResp.Key)
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	// Save downloaded file
	saveFile, err := os.Create("downloaded.txt")
	if err != nil {
		log.Fatalf("Failed to create save file: %v", err)
	}
	defer saveFile.Close()

	_, err = io.Copy(saveFile, downloadBody)
	if err != nil {
		log.Fatalf("Failed to save downloaded file: %v", err)
	}
	fmt.Println("File downloaded successfully to downloaded.txt")

	// 5. Delete file
	deleteReq := up.DeleteRequest{
		ID:     uploadReq.ID,
		Sha256: fileHash,
	}

	err = client.Delete(ctx, deleteReq)
	if err != nil {
		log.Fatalf("Delete failed: %v", err)
	}
	fmt.Println("File deleted successfully")
}
```

## Security and privacy

This project takes security seriously.
For vulnerability reporting and supported versions, see [SECURITY.md](SECURITY.md)

## License

This project is licensed under the [Apache-2.0 License](LICENSE).

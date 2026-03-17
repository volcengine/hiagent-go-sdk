package up

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestFull(t *testing.T) {
	ak := os.Getenv("VOLC_ACCESSKEY")
	sk := os.Getenv("VOLC_SECRETKEY")
	uploadEndpoint := os.Getenv("HIAGENT_UP_UPLOAD_ENDPOINT")

	if os.Getenv("OBSERVE_INTEGRATION_TESTS") == "" {
		t.Skip("skipping integration test; set OBSERVE_INTEGRATION_TESTS=1 to enable")
	}

	// Create client
	client := New(uploadEndpoint, ak, sk)
	ctx := context.Background()

	// Create test data directory
	testDataDir := "test_data"
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	// 1. Create test file
	testFilePath := "test_data/test.txt"
	testContent := "This is a test file for upload functionality.\nIt contains some sample text for testing.\n"
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer func() {
		if err := os.Remove(testFilePath); err != nil {
			t.Logf("Warning: failed to remove test file: %v", err)
		}
	}()

	testFile, err := os.Open(testFilePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		if closeErr := testFile.Close(); closeErr != nil {
			t.Logf("Warning: failed to close test file: %v", closeErr)
		}
	}()

	testFileHash, err := calculateSHA256(testFilePath)
	if err != nil {
		t.Fatalf("Failed to calculate SHA256: %v", err)
	}

	uploadReq := UploadRawRequest{
		ID:          strings.ReplaceAll(uuid.New().String(), "-", ""),
		ContentType: "plain/text",
		Expire:      "15h",
		Sha256:      testFileHash,
	}

	t.Logf("Upload request - ID: %s", uploadReq.ID)
	t.Logf("Upload request - Expire: %s", uploadReq.Expire)
	t.Logf("Upload request - SHA256: %s", uploadReq.Sha256)
	t.Logf("Upload request - ContentType: %s", uploadReq.ContentType)

	// Reopen file for upload
	if _, err := testFile.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek to beginning of file: %v", err)
	}
	uploadResp, err := client.UploadRaw(ctx, uploadReq, testFile)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if uploadResp.Sha256 != testFileHash {
		t.Fatalf("SHA256 mismatch: expected %s, got %s", testFileHash, uploadResp.Sha256)
	}
	t.Logf("Upload successful. Path: %s", uploadResp.Path)

	// 2. Get download key
	downloadKeyResp, err := client.DownloadKey(ctx, uploadResp.Path)
	if err != nil {
		t.Fatalf("Get download key failed: %v", err)
	}

	if len(downloadKeyResp.Key) <= 10 {
		t.Fatalf("Key length should be greater than 10, got %d", len(downloadKeyResp.Key))
	}
	t.Logf("Download key obtained: %s", downloadKeyResp.Key)

	// 3. Download file
	saveTo := "test_data/download.txt"
	downloadBody, err := client.Download(ctx, uploadResp.Path, downloadKeyResp.Key)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Save downloaded file
	saveFile, err := os.Create(saveTo)
	if err != nil {
		t.Fatalf("Failed to create save file: %v", err)
	}
	defer func() {
		if closeErr := saveFile.Close(); closeErr != nil {
			t.Logf("Warning: failed to close save file: %v", closeErr)
		}
		// Clean up downloaded file
		if err := os.Remove(saveTo); err != nil {
			t.Logf("Warning: failed to remove downloaded file: %v", err)
		}
	}()

	_, err = io.Copy(saveFile, downloadBody)
	if err != nil {
		t.Fatalf("Failed to save downloaded file: %v", err)
	}

	downloadedFileHash, err := calculateSHA256(saveTo)
	if err != nil {
		t.Fatalf("Failed to calculate SHA256 of downloaded file: %v", err)
	}

	// Verify downloaded file matches original
	if downloadedFileHash != testFileHash {
		t.Fatalf("SHA256 mismatch for downloaded file: expected %s, got %s", testFileHash, downloadedFileHash)
	}
	t.Logf("Download successful. Saved to: %s", saveTo)

	// 4. Delete file
	deleteReq := DeleteRequest{
		Sha256: testFileHash,
		ID:     uploadReq.ID,
	}

	err = client.Delete(ctx, deleteReq)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	t.Logf("Delete successful.")
}

// calculateSHA256 calculates the SHA256 hash of a file
func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Fatalf("close file failed: %v", closeErr)
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

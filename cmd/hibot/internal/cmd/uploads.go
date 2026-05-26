package cmd

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/volcengine/hiagent-go-sdk/hibot"
)

func newUploadsCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uploads",
		Short: "Upload artifacts to Hibot",
	}
	cmd.AddCommand(newUploadBlobCmd(v))
	return cmd
}

func newUploadBlobCmd(v *viper.Viper) *cobra.Command {
	var (
		filePath    string
		contentType string
	)
	cmd := &cobra.Command{
		Use:   "blob",
		Short: "Upload a file as a blob and print its BlobID",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return newUserError("--file is required")
			}
			f, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("open %s: %w", filePath, err)
			}
			defer f.Close()
			if contentType == "" {
				contentType = mime.TypeByExtension(filepath.Ext(filePath))
				if contentType == "" {
					contentType = "application/octet-stream"
				}
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			blob, err := client.V1.Uploads.UploadBlob(context.Background(), hibot.V1UploadBlobParams{
				Filename:    filepath.Base(filePath),
				ContentType: contentType,
			}, f)
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(blob, []string{"BLOB_ID"}, [][]string{{blob.BlobID}})
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "Local file path (required)")
	cmd.Flags().StringVar(&contentType, "content-type", "", "MIME type (auto-detected from extension if empty)")
	return cmd
}

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
	hibotv1 "github.com/volcengine/hiagent-go-sdk/hibot/v1"
)

func newResourcesCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resources",
		Aliases: []string{"resource"},
		Short:   "Manage Resources",
	}
	cmd.AddCommand(newResourcesListCmd(v))
	cmd.AddCommand(newResourcesGetByNameCmd(v))
	cmd.AddCommand(newResourcesCreateCmd(v))
	cmd.AddCommand(newResourcesDeleteCmd(v))
	cmd.AddCommand(newDirectoriesCmd(v))
	return cmd
}

func newResourcesListCmd(v *viper.Viper) *cobra.Command {
	var directoryID string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			result, err := client.V1.Resources.List(context.Background(), hibotv1.V1ResourceListParams{DirectoryID: directoryID})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(result.Items))
			for _, r := range result.Items {
				rows = append(rows, []string{r.ID, r.Name, r.Type, r.DirectoryID})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(result, []string{"ID", "NAME", "TYPE", "DIRECTORY_ID"}, rows)
		},
	}
	cmd.Flags().StringVar(&directoryID, "directory-id", "", "Filter by directory ID")
	return cmd
}

func newResourcesGetByNameCmd(v *viper.Viper) *cobra.Command {
	var directoryID string
	cmd := &cobra.Command{
		Use:   "get-by-name <name>",
		Short: "Get a resource by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			r, err := client.V1.Resources.GetByName(context.Background(), hibotv1.V1ResourceGetByNameParams{
				Name: args[0], DirectoryID: directoryID,
			})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(r,
				[]string{"ID", "NAME", "TYPE", "DIRECTORY_ID"},
				[][]string{{r.ID, r.Name, r.Type, r.DirectoryID}})
		},
	}
	cmd.Flags().StringVar(&directoryID, "directory-id", "", "Restrict lookup to a directory")
	return cmd
}

func newResourcesCreateCmd(v *viper.Viper) *cobra.Command {
	var (
		name        string
		resType     string
		directoryID string
		filePath    string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource by uploading a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return newUserError("--name is required")
			}
			if filePath == "" {
				return newUserError("--file is required")
			}
			f, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("open %s: %w", filePath, err)
			}
			defer f.Close()
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			ctx := context.Background()
			contentType := mime.TypeByExtension(filepath.Ext(filePath))
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			blob, err := client.V1.Uploads.UploadBlob(ctx, hibot.V1UploadBlobParams{
				Filename:    filepath.Base(filePath),
				ContentType: contentType,
			}, f)
			if err != nil {
				return err
			}
			r, err := client.V1.Resources.New(ctx, hibot.V1ResourceNewParams{
				Name:        name,
				Type:        resType,
				BlobID:      blob.BlobID,
				DirectoryID: directoryID,
			})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(r,
				[]string{"ID", "NAME", "TYPE", "DIRECTORY_ID"},
				[][]string{{r.ID, r.Name, r.Type, r.DirectoryID}})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Resource name (required)")
	cmd.Flags().StringVar(&resType, "type", "", "Resource type (e.g. document_collection)")
	cmd.Flags().StringVar(&directoryID, "directory-id", "", "Directory to place the resource in")
	cmd.Flags().StringVar(&filePath, "file", "", "Local file to upload (required)")
	return cmd
}

func newResourcesDeleteCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <resource-id>",
		Short: "Delete a resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			if err := client.V1.Resources.Delete(context.Background(), hibot.V1ResourceDeleteParams{ResourceID: args[0]}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "resource %s deleted\n", args[0])
			return nil
		},
	}
}

func newDirectoriesCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "directories",
		Short: "Manage resource directories",
	}
	cmd.AddCommand(newDirectoriesListCmd(v))
	cmd.AddCommand(newDirectoriesCreateCmd(v))
	cmd.AddCommand(newDirectoriesDeleteCmd(v))
	return cmd
}

func newDirectoriesListCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			result, err := client.V1.Resources.Directories.List(context.Background(), hibotv1.V1DirectoryListParams{})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(result.Items))
			for _, d := range result.Items {
				rows = append(rows, []string{d.ID, d.Name, fmt.Sprintf("%d", d.ResourceCount)})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(result, []string{"ID", "NAME", "RESOURCE_COUNT"}, rows)
		},
	}
}

func newDirectoriesCreateCmd(v *viper.Viper) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return newUserError("--name is required")
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			d, err := client.V1.Resources.Directories.New(context.Background(), hibotv1.V1DirectoryNewParams{Name: name})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(d,
				[]string{"ID", "NAME"},
				[][]string{{d.ID, d.Name}})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Directory name (required)")
	return cmd
}

func newDirectoriesDeleteCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <directory-id>",
		Short: "Delete a resource directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			if err := client.V1.Resources.Directories.Delete(context.Background(), hibotv1.V1DirectoryDeleteParams{DirectoryID: args[0]}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "directory %s deleted\n", args[0])
			return nil
		},
	}
}

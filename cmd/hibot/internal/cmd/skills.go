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

func newSkillsCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skills",
		Aliases: []string{"skill"},
		Short:   "Manage Skills",
	}
	cmd.AddCommand(newSkillsListCmd(v))
	cmd.AddCommand(newSkillsGetCmd(v))
	cmd.AddCommand(newSkillsDeleteCmd(v))
	cmd.AddCommand(newSkillsUploadCmd(v))
	cmd.AddCommand(newSkillsVersionsCmd(v))
	return cmd
}

func newSkillsListCmd(v *viper.Viper) *cobra.Command {
	var keyword string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			items, err := client.V1.Skills.List(context.Background(), hibotv1.V1SkillListParams{Keyword: keyword})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(items))
			for _, s := range items {
				rows = append(rows, []string{s.ID, s.SkillID, s.Name, s.Version, s.Source})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(items, []string{"ID", "SKILL_ID", "NAME", "VERSION", "SOURCE"}, rows)
		},
	}
	cmd.Flags().StringVar(&keyword, "keyword", "", "Filter by name keyword")
	return cmd
}

func newSkillsGetCmd(v *viper.Viper) *cobra.Command {
	var (
		id      string
		skillID string
		version string
	)
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a skill (by --id or --skill-id [--version])",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" && skillID == "" {
				return newUserError("--id or --skill-id is required")
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			skill, err := client.V1.Skills.Get(context.Background(), hibotv1.V1SkillGetParams{
				ID: id, SkillID: skillID, Version: version,
			})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(skill,
				[]string{"ID", "SKILL_ID", "NAME", "VERSION", "SOURCE"},
				[][]string{{skill.ID, skill.SkillID, skill.Name, skill.Version, skill.Source}})
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Skill version ID")
	cmd.Flags().StringVar(&skillID, "skill-id", "", "Skill ID")
	cmd.Flags().StringVar(&version, "version", "", "Skill version (with --skill-id)")
	return cmd
}

func newSkillsDeleteCmd(v *viper.Viper) *cobra.Command {
	var (
		id      string
		skillID string
		version string
	)
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a skill (by --id or --skill-id [--version])",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" && skillID == "" {
				return newUserError("--id or --skill-id is required")
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			if err := client.V1.Skills.Delete(context.Background(), hibotv1.V1SkillDeleteParams{
				ID: id, SkillID: skillID, Version: version,
			}); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "skill deleted")
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Skill version ID")
	cmd.Flags().StringVar(&skillID, "skill-id", "", "Skill ID")
	cmd.Flags().StringVar(&version, "version", "", "Skill version (with --skill-id)")
	return cmd
}

func newSkillsUploadCmd(v *viper.Viper) *cobra.Command {
	var (
		name        string
		description string
		version     string
		filePath    string
		source      string
	)
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a skill bundle and register it as a skill",
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
			skill, err := client.V1.Skills.New(ctx, hibot.V1SkillNewParams{
				Name:        name,
				Description: description,
				Version:     version,
				BlobID:      blob.BlobID,
				Source:      source,
			})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(skill,
				[]string{"ID", "SKILL_ID", "NAME", "VERSION"},
				[][]string{{skill.ID, skill.SkillID, skill.Name, skill.Version}})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Skill name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Skill description")
	cmd.Flags().StringVar(&version, "version", "", "Skill version (e.g. 1.0.0)")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to the skill bundle (required)")
	cmd.Flags().StringVar(&source, "source", "", "Source label (default \"manual\")")
	return cmd
}

func newSkillsVersionsCmd(v *viper.Viper) *cobra.Command {
	var skillID string
	cmd := &cobra.Command{
		Use:   "versions",
		Short: "List versions of a skill",
		RunE: func(cmd *cobra.Command, args []string) error {
			if skillID == "" {
				return newUserError("--skill-id is required")
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			items, err := client.V1.Skills.ListVersions(context.Background(), hibotv1.V1SkillVersionListParams{SkillID: skillID})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(items))
			for _, sv := range items {
				rows = append(rows, []string{sv.ID, sv.SkillID, sv.Version, sv.CreatedAt})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(items, []string{"ID", "SKILL_ID", "VERSION", "CREATED_AT"}, rows)
		},
	}
	cmd.Flags().StringVar(&skillID, "skill-id", "", "Skill ID (required)")
	return cmd
}

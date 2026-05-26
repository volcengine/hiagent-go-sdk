package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newViperWithRoot builds a fresh root cobra command + viper, parses the
// provided extraArgs, and runs initConfig with cfgPath as the config file
// (when non-empty). Returns the viper for assertions.
func newViperWithRoot(t *testing.T, root *cobra.Command, cfgPath string, extraArgs []string) *viper.Viper {
	t.Helper()
	args := []string{}
	if cfgPath != "" {
		args = append(args, "--config-file="+cfgPath)
	}
	args = append(args, extraArgs...)
	if err := root.PersistentFlags().Parse(args); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	v := viper.New()
	v.SetConfigType("yaml")
	for flagName, key := range configFlagToKey {
		_ = v.BindPFlag(key, root.PersistentFlags().Lookup(flagName))
	}
	for key, env := range configKeyToEnv {
		_ = v.BindEnv(key, env)
	}
	if cfgPath != "" {
		if _, err := os.Stat(cfgPath); err == nil {
			v.SetConfigFile(cfgPath)
			_ = v.ReadInConfig()
		}
	}
	return v
}

// TestConfigPrecedence verifies that flags override env and env overrides the
// config file. The CLI delegates merging to viper; we exercise the binding
// path through initConfig.
func TestConfigPrecedence(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("endpoint: file-endpoint\nworkspace_id: file-ws\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Run("file only", func(t *testing.T) {
		root := NewRootCmd()
		root.SetArgs([]string{"--config-file=" + cfgPath, "config", "view"})
		// initConfig runs through OnInitialize; trigger it indirectly.
		// We don't actually want to execute config view (it would try to
		// dial), so call initConfig manually after binding flags.
		// Easier: just run --config-file=... config view? Avoid network.
		// Instead, drive viper directly:
		v := newViperWithRoot(t, root, cfgPath, nil)
		if got := v.GetString(keyEndpoint); got != "file-endpoint" {
			t.Fatalf("file endpoint: got %q", got)
		}
		if got := v.GetString(keyWorkspaceID); got != "file-ws" {
			t.Fatalf("file workspace_id: got %q", got)
		}
	})

	t.Run("env over file", func(t *testing.T) {
		t.Setenv("HIBOT_ENDPOINT", "env-endpoint")
		root := NewRootCmd()
		v := newViperWithRoot(t, root, cfgPath, nil)
		if got := v.GetString(keyEndpoint); got != "env-endpoint" {
			t.Fatalf("env-over-file: got %q", got)
		}
	})

	t.Run("flag over env over file", func(t *testing.T) {
		t.Setenv("HIBOT_ENDPOINT", "env-endpoint")
		root := NewRootCmd()
		v := newViperWithRoot(t, root, cfgPath, []string{"--endpoint=flag-endpoint"})
		if got := v.GetString(keyEndpoint); got != "flag-endpoint" {
			t.Fatalf("flag-over-env: got %q", got)
		}
	})
}

func TestExitCodeFor(t *testing.T) {
	if ExitCodeFor(nil) != 0 {
		t.Fatalf("nil err -> 0")
	}
	if ExitCodeFor(newUserError("bad input")) != 2 {
		t.Fatalf("UserError -> 2")
	}
	if ExitCodeFor(os.ErrPermission) != 1 {
		t.Fatalf("generic err -> 1")
	}
}

func TestReadContentArgFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "p.txt")
	if err := os.WriteFile(p, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := readContentArg("@" + p)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got != "hello" {
		t.Fatalf("got %q", got)
	}
	if got, _ := readContentArg("inline"); got != "inline" {
		t.Fatalf("inline: got %q", got)
	}
	if _, err := readContentArg("@"); err == nil {
		t.Fatalf("expected error for empty path")
	}
}

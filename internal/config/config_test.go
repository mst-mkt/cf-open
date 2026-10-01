package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := Load(Options{Path: "/nonexistent/path/wrangler.json"})
	if err == nil {
		t.Error("Load() expected error for nonexistent file, got nil")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	t.Parallel()

	_, err := Load(Options{})
	if err == nil {
		t.Error("Load() expected error for empty path with no wrangler config, got nil")
	}
}

func TestLoad_WranglerTypeScriptConfig(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"wrangler.config.ts", "Wrangler.Config.TS"} {
		configPath := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(configPath, []byte("export default {}\n"), 0o644); err != nil {
			t.Fatalf("テスト設定ファイルの書き込みに失敗: %v", err)
		}

		_, err := Load(Options{Path: configPath})

		if err == nil || !strings.Contains(err.Error(), "pass cloudflare.config.ts instead") {
			t.Errorf("Load(%q) error = %v, want a wrangler.config.ts error", name, err)
		}
	}
}

func TestFindConfig_Priority(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"cloudflare.config.ts", "wrangler.jsonc"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o644); err != nil {
			t.Fatalf("テスト設定ファイルの書き込みに失敗: %v", err)
		}
	}
	t.Chdir(dir)

	got := findConfig()

	if got != "cloudflare.config.ts" {
		t.Errorf("findConfig() = %q, want %q", got, "cloudflare.config.ts")
	}
}

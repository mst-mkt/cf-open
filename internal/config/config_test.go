package config

import (
	"os"
	"path/filepath"
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

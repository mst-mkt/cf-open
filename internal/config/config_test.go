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
	t.Chdir(t.TempDir())

	_, err := Load(Options{})
	if err == nil {
		t.Error("Load() expected error for empty path with no wrangler config, got nil")
	}
}

func TestLoad_DiscoveredPath(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"wrangler.json":    `{"name": "parent-worker"}`,
		"src/nested/.keep": "",
	})
	t.Chdir(filepath.Join(root, "src", "nested"))

	cfg, err := Load(Options{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Name != "parent-worker" {
		t.Errorf("Name = %q, want %q", cfg.Name, "parent-worker")
	}
	if filepath.Base(cfg.Path) != "wrangler.json" || !filepath.IsAbs(cfg.Path) {
		t.Errorf("Path = %q, want an absolute path to wrangler.json", cfg.Path)
	}
}

func TestLoad_GivenPath(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "wrangler.json")
	writeFiles(t, filepath.Dir(configPath), map[string]string{"wrangler.json": `{"name": "given-worker"}`})

	cfg, err := Load(Options{Path: configPath})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Path != configPath {
		t.Errorf("Path = %q, want %q", cfg.Path, configPath)
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

func TestFindConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		files []string
		start string
		want  string
	}{
		{
			name:  "同じディレクトリでの優先順",
			files: []string{"cloudflare.config.ts", "wrangler.jsonc", "wrangler.toml"},
			start: ".",
			want:  "cloudflare.config.ts",
		},
		{
			name:  "親ディレクトリの設定ファイル",
			files: []string{"wrangler.toml", "packages/app/src/.keep"},
			start: "packages/app/src",
			want:  "wrangler.toml",
		},
		{
			name:  "最も近いディレクトリの設定ファイル",
			files: []string{"cloudflare.config.ts", "packages/app/wrangler.jsonc", "packages/app/src/.keep"},
			start: "packages/app/src",
			want:  "packages/app/wrangler.jsonc",
		},
		{
			name:  "設定ファイルと同名のディレクトリ",
			files: []string{"wrangler.toml", "app/wrangler.jsonc/.keep"},
			start: "app",
			want:  "wrangler.toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			files := make(map[string]string, len(tt.files))
			for _, name := range tt.files {
				files[name] = ""
			}
			writeFiles(t, root, files)

			got := findConfig(filepath.Join(root, filepath.FromSlash(tt.start)))

			if want := filepath.Join(root, filepath.FromSlash(tt.want)); got != want {
				t.Errorf("findConfig() = %q, want %q", got, want)
			}
		})
	}
}

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()

	for name, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("テスト用ファイルの書き込みに失敗: %v", err)
		}
	}
}

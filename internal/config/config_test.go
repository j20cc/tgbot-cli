package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTokenPriority(t *testing.T) {
	t.Setenv("TG_BOT_TOKEN", "env-token")

	tok, err := ResolveToken(TokenOptions{TokenFlag: "flag-token"})
	if err != nil {
		t.Fatalf("ResolveToken returned error: %v", err)
	}
	if tok != "flag-token" {
		t.Fatalf("expected flag-token, got %q", tok)
	}
}

func TestResolveTokenFromEnv(t *testing.T) {
	t.Setenv("TG_BOT_TOKEN", "env-token")
	tok, err := ResolveToken(TokenOptions{})
	if err != nil {
		t.Fatalf("ResolveToken returned error: %v", err)
	}
	if tok != "env-token" {
		t.Fatalf("expected env-token, got %q", tok)
	}
}

func TestResolveTokenFromRawFile(t *testing.T) {
	t.Setenv("TG_BOT_TOKEN", "")
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfg, []byte("raw-file-token\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	tok, err := ResolveToken(TokenOptions{ConfigPath: cfg})
	if err != nil {
		t.Fatalf("ResolveToken returned error: %v", err)
	}
	if tok != "raw-file-token" {
		t.Fatalf("expected raw-file-token, got %q", tok)
	}
}

func TestResolveTokenFromJSONProfile(t *testing.T) {
	t.Setenv("TG_BOT_TOKEN", "")
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	payload := `{"active_profile":"dev","profiles":{"dev":{"token":"json-token"}}}`
	if err := os.WriteFile(cfg, []byte(payload), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	tok, err := ResolveToken(TokenOptions{ConfigPath: cfg})
	if err != nil {
		t.Fatalf("ResolveToken returned error: %v", err)
	}
	if tok != "json-token" {
		t.Fatalf("expected json-token, got %q", tok)
	}
}

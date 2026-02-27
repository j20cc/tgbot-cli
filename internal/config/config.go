package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type TokenOptions struct {
	TokenFlag  string
	ConfigPath string
	Profile    string
}

type fileConfig struct {
	ActiveProfile string                   `json:"active_profile"`
	Profiles      map[string]profileConfig `json:"profiles"`
}

type profileConfig struct {
	Token string `json:"token"`
}

func ResolveToken(opts TokenOptions) (string, error) {
	if opts.TokenFlag != "" {
		return opts.TokenFlag, nil
	}

	if envToken := os.Getenv("TG_BOT_TOKEN"); envToken != "" {
		return envToken, nil
	}

	cfgPath, err := resolveConfigPath(opts.ConfigPath)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("token not found: use --token / TG_BOT_TOKEN / %s", cfgPath)
		}
		return "", fmt.Errorf("read config: %w", err)
	}

	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return "", errors.New("config is empty")
	}

	// If file content is not JSON object, treat it as raw token text for convenience.
	if !strings.HasPrefix(trimmed, "{") {
		return trimmed, nil
	}

	var cfg fileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse config json: %w", err)
	}

	profile := opts.Profile
	if profile == "" {
		profile = cfg.ActiveProfile
	}
	if profile == "" {
		return "", errors.New("no profile specified and active_profile is empty")
	}

	p, ok := cfg.Profiles[profile]
	if !ok {
		return "", fmt.Errorf("profile %q not found", profile)
	}
	if p.Token == "" {
		return "", fmt.Errorf("profile %q has empty token", profile)
	}

	return p.Token, nil
}

func resolveConfigPath(pathFlag string) (string, error) {
	if pathFlag != "" {
		return pathFlag, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".tgbot-cli", "config.json"), nil
}

// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package core defines the on-disk config structures and helpers for
// resolving the active profile.
package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EnvAPIKey is the env var that, when set, overrides the configured key.
const EnvAPIKey = "JODOO_API_KEY"

// EnvBaseURL overrides the default base URL.
const EnvBaseURL = "JODOO_BASE_URL"

// EnvProfile overrides the default profile.
const EnvProfile = "JODOO_PROFILE"

// EnvHome overrides the config dir (~/.jodoo-cli).
const EnvHome = "JODOO_CLI_HOME"

// DefaultBaseURL is the public Jodoo API endpoint.
const DefaultBaseURL = "https://api.jodoo.com/api"

// DefaultProfile is the profile name when none is selected.
const DefaultProfile = "default"

// CliConfig is the in-memory representation of one profile.
type CliConfig struct {
	Profile string `json:"profile"`
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url,omitempty"`
	// Notes is a free-form label so users can recognize a profile.
	Notes string `json:"notes,omitempty"`
}

// ConfigFile is the on-disk JSON layout: a map of profiles + the default.
type ConfigFile struct {
	Default  string                `json:"default"`
	Profiles map[string]*CliConfig `json:"profiles"`
}

// ConfigError is returned when the config cannot be loaded or is missing.
type ConfigError struct {
	Code    int
	Type    string
	Message string
	Hint    string
}

func (e *ConfigError) Error() string { return e.Message }

// HomeDir returns the jodoo-cli config directory (~/.jodoo-cli by default).
func HomeDir() (string, error) {
	if v := os.Getenv(EnvHome); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".jodoo-cli"), nil
}

// ConfigPath returns the path to config.json.
func ConfigPath() (string, error) {
	dir, err := HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// EnsureHomeDir creates the config dir with mode 0700 if missing.
func EnsureHomeDir() (string, error) {
	dir, err := HomeDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(dir, 0o700)
	}
	return dir, nil
}

// LoadFile reads and parses config.json. Returns an empty ConfigFile if
// the file does not exist (caller can treat this as "no profiles yet").
func LoadFile() (*ConfigFile, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ConfigFile{Profiles: map[string]*CliConfig{}}, nil
		}
		return nil, err
	}
	cf := &ConfigFile{}
	if err := json.Unmarshal(data, cf); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if cf.Profiles == nil {
		cf.Profiles = map[string]*CliConfig{}
	}
	return cf, nil
}

// SaveFile writes config.json atomically with mode 0600.
func SaveFile(cf *ConfigFile) error {
	if _, err := EnsureHomeDir(); err != nil {
		return err
	}
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ResolveProfileName returns the profile to use, considering:
// 1. explicit override (from --profile flag)
// 2. JODOO_PROFILE env
// 3. ConfigFile.Default
// 4. "default"
func ResolveProfileName(override string, cf *ConfigFile) string {
	if override != "" {
		return override
	}
	if v := os.Getenv(EnvProfile); v != "" {
		return v
	}
	if cf != nil && cf.Default != "" {
		return cf.Default
	}
	return DefaultProfile
}

// LoadResolved loads the config file and returns the active profile,
// merged with env overrides (JODOO_API_KEY, JODOO_BASE_URL).
//
// If the resolved profile has no API key (and JODOO_API_KEY is not set),
// returns a ConfigError with hint to run `jodoo-cli config init`.
func LoadResolved(profileOverride string) (*CliConfig, error) {
	cf, err := LoadFile()
	if err != nil {
		return nil, err
	}
	name := ResolveProfileName(profileOverride, cf)

	cfg := &CliConfig{Profile: name}
	if p, ok := cf.Profiles[name]; ok && p != nil {
		cfg.APIKey = p.APIKey
		cfg.BaseURL = p.BaseURL
		cfg.Notes = p.Notes
	}
	if v := os.Getenv(EnvAPIKey); v != "" {
		cfg.APIKey = strings.TrimSpace(v)
	}
	if v := os.Getenv(EnvBaseURL); v != "" {
		cfg.BaseURL = strings.TrimSpace(v)
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.APIKey == "" {
		return cfg, &ConfigError{
			Code:    2,
			Type:    "config_missing",
			Message: fmt.Sprintf("no API key set for profile %q", name),
			Hint:    "run `jodoo-cli config init` or export JODOO_API_KEY=...",
		}
	}
	return cfg, nil
}

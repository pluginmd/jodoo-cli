// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package credential handles API key persistence.
//
// Jodoo auth is a simple bearer token — no OAuth flows, no refresh, no
// scopes. The key is stored in ~/.jodoo-cli/config.json by default. For
// extra safety the user can choose to keep it in the OS keychain instead
// (macOS Keychain / Linux Secret Service / Windows Credential Manager)
// by using the --use-keychain flag during `jodoo-cli config init`.
package credential

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"

	"jodoo-cli/internal/core"
)

// KeychainService is the service name used under the OS secret store.
const KeychainService = "jodoo-cli"

// KeychainAccount builds a stable keychain account key for a profile.
func KeychainAccount(profile string) string {
	if profile == "" {
		profile = core.DefaultProfile
	}
	return "apikey:" + profile
}

// SetInKeychain stores the API key in the OS keychain under the profile.
func SetInKeychain(profile, apiKey string) error {
	return keyring.Set(KeychainService, KeychainAccount(profile), apiKey)
}

// GetFromKeychain returns the API key from the OS keychain (empty if not set).
func GetFromKeychain(profile string) (string, error) {
	v, err := keyring.Get(KeychainService, KeychainAccount(profile))
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(v), nil
}

// DeleteFromKeychain removes the profile's API key from the keychain.
func DeleteFromKeychain(profile string) error {
	err := keyring.Delete(KeychainService, KeychainAccount(profile))
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return err
	}
	return nil
}

// Resolve returns the API key for a profile, consulting (in order):
//  1. JODOO_API_KEY env var
//  2. keychain (if available)
//  3. config.json (via core.CliConfig.APIKey)
//
// cfg.APIKey is assumed to already reflect env + file merge from core.LoadResolved.
// This helper only escalates to the keychain when the config file did not
// provide a value.
func Resolve(profile string, cfg *core.CliConfig) (string, error) {
	if cfg != nil && cfg.APIKey != "" {
		return cfg.APIKey, nil
	}
	v, err := GetFromKeychain(profile)
	if err != nil {
		return "", fmt.Errorf("read keychain: %w", err)
	}
	return v, nil
}

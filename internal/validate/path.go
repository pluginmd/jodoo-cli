// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package validate contains small input-safety helpers used across
// commands and shortcuts.
package validate

import (
	"errors"
	"path/filepath"
	"strings"
)

// SafeInputPath rejects obviously unsafe paths (empty, traversal-only).
// It does NOT touch the filesystem — call os.Stat after if you need to
// confirm the file exists.
func SafeInputPath(p string) (string, error) {
	if strings.TrimSpace(p) == "" {
		return "", errors.New("empty path")
	}
	clean := filepath.Clean(p)
	if clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "..\\") {
		return "", errors.New("path traversal not allowed")
	}
	return clean, nil
}

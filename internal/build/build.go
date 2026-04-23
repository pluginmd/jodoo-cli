// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package build holds version metadata injected at build time via -ldflags.
package build

// Version is set via -ldflags "-X jodoo-cli/internal/build.Version=...".
var Version = "dev"

// Date is set via -ldflags "-X jodoo-cli/internal/build.Date=...".
var Date = "unknown"

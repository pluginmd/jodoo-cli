// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package cmd

import "github.com/spf13/pflag"

// GlobalOptions holds top-level flags persisted on the root command.
type GlobalOptions struct {
	Profile string
}

// RegisterGlobalFlags wires --profile onto the root command's persistent
// flag set. Other "global"-feeling flags (--format, --jq, --dry-run) are
// registered per-shortcut by the shortcuts framework so subcommands that
// don't use them stay clean.
func RegisterGlobalFlags(fs *pflag.FlagSet, g *GlobalOptions) {
	fs.StringVar(&g.Profile, "profile", "", "config profile to use (overrides JODOO_PROFILE / config default)")
}

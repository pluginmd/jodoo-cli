// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package config implements `jodoo-cli config <subcmd>`.
package config

import (
	"github.com/spf13/cobra"

	"jodoo-cli/internal/cmdutil"
)

// NewCmdConfig builds the parent `config` command.
func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage jodoo-cli configuration (API key, base URL, profiles)",
		Long: `Manage jodoo-cli configuration.

Subcommands:
    init     interactively configure an API key (and optional base URL)
    show     print the resolved profile (API key is masked)
    remove   remove a profile from config.json (and the keychain)
    set-default  pick the default profile

Storage:
    By default the API key is written to ~/.jodoo-cli/config.json (mode 0600).
    Pass --use-keychain to keep it in the OS keychain instead. Either way
    the env var JODOO_API_KEY (when set) overrides whatever is on disk.`,
	}
	cmd.AddCommand(newInitCmd(f))
	cmd.AddCommand(newShowCmd(f))
	cmd.AddCommand(newRemoveCmd(f))
	cmd.AddCommand(newSetDefaultCmd(f))
	return cmd
}

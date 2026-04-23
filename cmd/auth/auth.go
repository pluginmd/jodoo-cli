// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package auth implements `jodoo-cli auth <subcmd>`.
//
// Jodoo auth is a single Bearer token (the "API key" you generate at
// Open Platform → API Key). There is no OAuth flow, no token refresh,
// no scope grant. The subcommands here just wrap key set/show/clear.
package auth

import (
	"github.com/spf13/cobra"

	"jodoo-cli/internal/cmdutil"
)

// NewCmdAuth builds the parent `auth` command.
func NewCmdAuth(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage the Jodoo API key (Bearer token)",
		Long: `Manage the Jodoo API key.

Jodoo authentication is a simple Bearer token. Generate it at:
  Open Platform → API Key → Create API Key

Subcommands:
    set       set or replace the API key for a profile
    status    print which profile is active and whether a key is set
    clear     remove the API key from the active profile`,
	}
	cmd.AddCommand(newSetCmd(f))
	cmd.AddCommand(newStatusCmd(f))
	cmd.AddCommand(newClearCmd(f))
	return cmd
}

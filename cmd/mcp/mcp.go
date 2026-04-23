// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package mcp implements a Model Context Protocol server on top of the
// existing shortcut registry. It lets MCP-aware clients (Claude Desktop,
// Claude Code, Cursor, Zed, …) drive Jodoo with the same curated commands
// a human uses on the terminal.
//
// See docs/MCP/ for architecture, setup, and security notes.
package mcp

import (
	"github.com/spf13/cobra"

	"jodoo-cli/internal/cmdutil"
)

// ServeOptions are the knobs the `mcp serve` subcommand exposes.
type ServeOptions struct {
	HTTPAddr string   // empty = stdio; ":8765" or "127.0.0.1:8765" = HTTP
	Token    string   // HTTP Bearer auth (falls back to MCP_TOKEN env)
	LogFile  string   // divert server logs; stderr otherwise
	Allow    []string // glob allow-list of tool names
	Deny     []string // glob deny-list of tool names (applied after allow)
	ReadOnly bool     // hide non-read shortcuts
}

// NewCmdMcp returns the `mcp` command group.
func NewCmdMcp(f *cmdutil.Factory) *cobra.Command {
	c := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol layer (Claude Desktop, etc.)",
		Long: `Expose jodoo-cli shortcuts to MCP-aware clients.

  jodoo-cli mcp serve                 # speak MCP on stdio (default)
  jodoo-cli mcp serve --read-only     # hide every non-read shortcut
  jodoo-cli mcp serve --log-file /tmp/jodoo-mcp.log

Every curated +shortcut becomes a tool named {service}.{command-without-plus}
(e.g. +data-list → jodoo.data-list). Risk gates carry over — high-risk
writes require an explicit confirm:true in the tool arguments.

See docs/MCP/ for the full guide.`,
	}
	c.AddCommand(newCmdServe(f))
	return c
}

func newCmdServe(f *cmdutil.Factory) *cobra.Command {
	opts := &ServeOptions{}
	c := &cobra.Command{
		Use:   "serve",
		Short: "Run an MCP server over stdio",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return Serve(cmd.Context(), f, opts)
		},
	}
	c.Flags().StringVar(&opts.HTTPAddr, "http", "", "bind address for HTTP transport (e.g. 127.0.0.1:8765; empty = stdio)")
	c.Flags().StringVar(&opts.Token, "token", "", "Bearer token for HTTP auth (or set MCP_TOKEN env)")
	c.Flags().StringVar(&opts.LogFile, "log-file", "", "write diagnostic logs here (stderr otherwise)")
	c.Flags().StringSliceVar(&opts.Allow, "allow", nil, "allow-list of tool names (glob, repeatable)")
	c.Flags().StringSliceVar(&opts.Deny, "deny", nil, "deny-list of tool names (glob, repeatable)")
	c.Flags().BoolVar(&opts.ReadOnly, "read-only", false, "only expose shortcuts with Risk=read")
	return c
}

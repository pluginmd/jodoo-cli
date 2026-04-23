// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package common is the declarative shortcut framework: each Jodoo
// shortcut declares its flags + hooks (Validate / DryRun / Execute) and
// the framework wires it to a cobra subcommand.
package common

import "context"

// Flag.Input source constants.
const (
	File  = "file"  // support @path to read value from a file
	Stdin = "stdin" // support - to read value from stdin
)

// Flag describes a CLI flag for a shortcut.
type Flag struct {
	Name     string   // e.g. "app-id"
	Type     string   // "string" (default) | "bool" | "int" | "string_array"
	Default  string   // default value as string
	Desc     string   // help text
	Hidden   bool     // hidden from --help, still readable at runtime
	Required bool
	Enum     []string // allowed values; empty = unconstrained
	Input    []string // extra input sources: File (@path), Stdin (-); empty = flag value only
}

// Shortcut represents a high-level CLI command bound to a Jodoo endpoint.
type Shortcut struct {
	Service     string   // bucket name shown under root, e.g. "jodoo"
	Command     string   // subcommand verb, e.g. "+data-list" (with leading "+")
	Description string   // short help text
	Risk        string   // "read" | "write" | "high-risk-write" (empty = "read")
	Flags       []Flag   // user-facing flags
	HasFormat   bool     // auto-inject --format flag
	Tips        []string // optional tips appended to --help

	// Hooks.
	DryRun   func(ctx context.Context, runtime *RuntimeContext) *DryRunAPI
	Validate func(ctx context.Context, runtime *RuntimeContext) error
	Execute  func(ctx context.Context, runtime *RuntimeContext) error
}

// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package shortcuts wires every Shortcut declared under shortcuts/jodoo
// (and any future package) onto the cobra root command.
package shortcuts

import (
	"sort"

	"github.com/spf13/cobra"

	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/shortcuts/common"
	"jodoo-cli/shortcuts/jodoo"
)

// Provider is implemented by any shortcut bundle.
type Provider interface {
	Service() string
	Description() string
	All() []common.Shortcut
}

// providers returns the registered bundles, in display order.
func providers() []Provider {
	return []Provider{jodooProvider{}}
}

// Providers exposes the registered bundles to callers outside this package
// (the MCP layer iterates them to build its tool registry).
func Providers() []Provider {
	return providers()
}

type jodooProvider struct{}

func (jodooProvider) Service() string { return "jodoo" }
func (jodooProvider) Description() string {
	return "Jodoo shortcuts (apps, forms, records, files, workflow, contacts)"
}
func (jodooProvider) All() []common.Shortcut { return jodoo.Shortcuts() }

// RegisterShortcuts mounts each bundle as a subcommand of root and each
// shortcut as a verb under that subcommand:
//
//	jodoo-cli jodoo +app-list
//	jodoo-cli jodoo +data-list
//	...
func RegisterShortcuts(root *cobra.Command, f *cmdutil.Factory) {
	for _, p := range providers() {
		svc := p.Service()
		bucket := &cobra.Command{
			Use:   svc,
			Short: p.Description(),
		}
		shortcuts := p.All()
		// stable, alphabetical order so --help is predictable
		sort.SliceStable(shortcuts, func(i, j int) bool {
			return shortcuts[i].Command < shortcuts[j].Command
		})
		for _, sc := range shortcuts {
			sc.Mount(bucket, f)
		}
		root.AddCommand(bucket)
	}
}

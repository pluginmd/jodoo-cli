// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/internal/core"
)

func newSetDefaultCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-default <profile>",
		Short: "Set the default profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cf, err := core.LoadFile()
			if err != nil {
				return err
			}
			if _, ok := cf.Profiles[name]; !ok {
				return fmt.Errorf("profile %q not found (run `jodoo-cli config init --profile %s` first)", name, name)
			}
			cf.Default = name
			if err := core.SaveFile(cf); err != nil {
				return err
			}
			f.ResetConfig()
			fmt.Fprintf(f.IOStreams.Out, "✓ default profile is now %q\n", name)
			return nil
		},
	}
	return cmd
}

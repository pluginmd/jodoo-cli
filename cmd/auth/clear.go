// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/internal/core"
	"jodoo-cli/internal/credential"
)

func newClearCmd(f *cmdutil.Factory) *cobra.Command {
	var profile string
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Remove the API key from a profile (config.json + keychain)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if profile == "" {
				cf, _ := core.LoadFile()
				profile = core.ResolveProfileName(f.ProfileOverride, cf)
			}
			cf, err := core.LoadFile()
			if err != nil {
				return err
			}
			if p, ok := cf.Profiles[profile]; ok && p != nil {
				p.APIKey = ""
				if err := core.SaveFile(cf); err != nil {
					return err
				}
			}
			if err := credential.DeleteFromKeychain(profile); err != nil {
				fmt.Fprintf(f.IOStreams.ErrOut, "warning: keychain delete: %v\n", err)
			}
			f.ResetConfig()
			fmt.Fprintf(f.IOStreams.Out, "✓ cleared API key for profile %q\n", profile)
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "profile to clear (default: active profile)")
	return cmd
}

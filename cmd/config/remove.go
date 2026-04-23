// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/internal/core"
	"jodoo-cli/internal/credential"
)

func newRemoveCmd(f *cmdutil.Factory) *cobra.Command {
	var profile string
	var keepKeychain bool
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Delete a profile from config.json (and the keychain)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if profile == "" {
				return fmt.Errorf("--profile is required")
			}
			cf, err := core.LoadFile()
			if err != nil {
				return err
			}
			if _, ok := cf.Profiles[profile]; !ok {
				return fmt.Errorf("profile %q not found", profile)
			}
			delete(cf.Profiles, profile)
			if cf.Default == profile {
				cf.Default = ""
				for k := range cf.Profiles {
					cf.Default = k
					break
				}
			}
			if err := core.SaveFile(cf); err != nil {
				return err
			}
			f.ResetConfig()
			if !keepKeychain {
				if err := credential.DeleteFromKeychain(profile); err != nil {
					fmt.Fprintf(f.IOStreams.ErrOut, "warning: keychain delete: %v\n", err)
				}
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ removed profile %q (default=%q)\n", profile, cf.Default)
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "profile to remove")
	cmd.Flags().BoolVar(&keepKeychain, "keep-keychain", false, "do not delete the matching keychain entry")
	return cmd
}

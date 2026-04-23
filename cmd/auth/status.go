// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/internal/core"
	"jodoo-cli/internal/credential"
	"jodoo-cli/internal/output"
)

func newStatusCmd(f *cmdutil.Factory) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show whether an API key is configured for the active profile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cf, err := core.LoadFile()
			if err != nil {
				return err
			}
			profile := core.ResolveProfileName(f.ProfileOverride, cf)

			fileKey := ""
			if p, ok := cf.Profiles[profile]; ok && p != nil {
				fileKey = p.APIKey
			}
			kcKey, _ := credential.GetFromKeychain(profile)
			envKey := os.Getenv(core.EnvAPIKey)

			source := "none"
			switch {
			case envKey != "":
				source = "env (JODOO_API_KEY)"
			case fileKey != "":
				source = "config.json"
			case kcKey != "":
				source = "keychain"
			}

			snapshot := map[string]interface{}{
				"profile":     profile,
				"has_key":     envKey != "" || fileKey != "" || kcKey != "",
				"source":      source,
				"in_config":   fileKey != "",
				"in_keychain": kcKey != "",
				"in_env":      envKey != "",
			}
			if jsonOut {
				output.PrintJson(f.IOStreams.Out, snapshot)
				return nil
			}
			out := f.IOStreams.Out
			fmt.Fprintf(out, "profile      : %s\n", profile)
			fmt.Fprintf(out, "key configured: %v\n", snapshot["has_key"])
			fmt.Fprintf(out, "source       : %s\n", source)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "machine-readable JSON output")
	return cmd
}

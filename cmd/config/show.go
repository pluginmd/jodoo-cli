// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/internal/core"
	"jodoo-cli/internal/credential"
	"jodoo-cli/internal/output"
)

func newShowCmd(f *cmdutil.Factory) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Print resolved profile (API key masked)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cf, err := core.LoadFile()
			if err != nil {
				return err
			}
			f.ResetConfig()
			cfg, _ := core.LoadResolved(f.ProfileOverride)
			if cfg == nil {
				cfg = &core.CliConfig{Profile: core.ResolveProfileName(f.ProfileOverride, cf)}
			}
			if cfg.APIKey == "" {
				if v, _ := credential.GetFromKeychain(cfg.Profile); v != "" {
					cfg.APIKey = v
				}
			}
			snapshot := map[string]interface{}{
				"profile":  cfg.Profile,
				"base_url": cfg.BaseURL,
				"notes":    cfg.Notes,
				"api_key":  maskKey(cfg.APIKey),
				"default":  cf.Default,
				"profiles": profileNames(cf),
			}
			if jsonOut {
				output.PrintJson(f.IOStreams.Out, snapshot)
				return nil
			}
			out := f.IOStreams.Out
			fmt.Fprintf(out, "active profile : %s\n", cfg.Profile)
			fmt.Fprintf(out, "base url       : %s\n", cfg.BaseURL)
			fmt.Fprintf(out, "api key        : %s\n", snapshot["api_key"])
			if cfg.Notes != "" {
				fmt.Fprintf(out, "notes          : %s\n", cfg.Notes)
			}
			fmt.Fprintf(out, "default profile: %s\n", cf.Default)
			fmt.Fprintf(out, "profiles       : %v\n", snapshot["profiles"])
			path, _ := core.ConfigPath()
			fmt.Fprintf(out, "config file    : %s\n", path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "machine-readable JSON output")
	return cmd
}

func profileNames(cf *core.ConfigFile) []string {
	out := make([]string, 0, len(cf.Profiles))
	for k := range cf.Profiles {
		out = append(out, k)
	}
	return out
}

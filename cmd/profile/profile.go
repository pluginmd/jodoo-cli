// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package profile implements `jodoo-cli profile <subcmd>`.
package profile

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/internal/core"
	"jodoo-cli/internal/credential"
	"jodoo-cli/internal/output"
)

// NewCmdProfile builds the parent `profile` command.
func NewCmdProfile(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "List / switch / rename / remove configured profiles",
	}
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newUseCmd(f))
	cmd.AddCommand(newRenameCmd(f))
	cmd.AddCommand(newRemoveCmd(f))
	return cmd
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured profiles",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cf, err := core.LoadFile()
			if err != nil {
				return err
			}
			names := make([]string, 0, len(cf.Profiles))
			for n := range cf.Profiles {
				names = append(names, n)
			}
			sort.Strings(names)
			if jsonOut {
				output.PrintJson(f.IOStreams.Out, map[string]interface{}{
					"default":  cf.Default,
					"profiles": names,
				})
				return nil
			}
			out := f.IOStreams.Out
			fmt.Fprintf(out, "default: %s\n", cf.Default)
			fmt.Fprintln(out, "profiles:")
			for _, n := range names {
				marker := "  "
				if n == cf.Default {
					marker = "* "
				}
				p := cf.Profiles[n]
				baseURL := ""
				notes := ""
				if p != nil {
					baseURL = p.BaseURL
					notes = p.Notes
				}
				fmt.Fprintf(out, "%s%-15s  %s  %s\n", marker, n, baseURL, notes)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "machine-readable JSON output")
	return cmd
}

func newUseCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <profile>",
		Short: "Set the default profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cf, err := core.LoadFile()
			if err != nil {
				return err
			}
			if _, ok := cf.Profiles[args[0]]; !ok {
				return fmt.Errorf("profile %q not found", args[0])
			}
			cf.Default = args[0]
			if err := core.SaveFile(cf); err != nil {
				return err
			}
			f.ResetConfig()
			fmt.Fprintf(f.IOStreams.Out, "✓ default profile is now %q\n", args[0])
			return nil
		},
	}
	return cmd
}

func newRenameCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <old> <new>",
		Short: "Rename a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			old, neu := args[0], args[1]
			cf, err := core.LoadFile()
			if err != nil {
				return err
			}
			p, ok := cf.Profiles[old]
			if !ok {
				return fmt.Errorf("profile %q not found", old)
			}
			if _, exists := cf.Profiles[neu]; exists {
				return fmt.Errorf("profile %q already exists", neu)
			}
			p.Profile = neu
			cf.Profiles[neu] = p
			delete(cf.Profiles, old)
			if cf.Default == old {
				cf.Default = neu
			}
			if err := core.SaveFile(cf); err != nil {
				return err
			}
			// Migrate keychain entry, if any
			if v, _ := credential.GetFromKeychain(old); v != "" {
				_ = credential.SetInKeychain(neu, v)
				_ = credential.DeleteFromKeychain(old)
			}
			f.ResetConfig()
			fmt.Fprintf(f.IOStreams.Out, "✓ renamed %q → %q\n", old, neu)
			return nil
		},
	}
	return cmd
}

func newRemoveCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <profile>",
		Short: "Remove a profile (and its keychain entry, if any)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cf, err := core.LoadFile()
			if err != nil {
				return err
			}
			if _, ok := cf.Profiles[name]; !ok {
				return fmt.Errorf("profile %q not found", name)
			}
			delete(cf.Profiles, name)
			if cf.Default == name {
				cf.Default = ""
				for k := range cf.Profiles {
					cf.Default = k
					break
				}
			}
			if err := core.SaveFile(cf); err != nil {
				return err
			}
			_ = credential.DeleteFromKeychain(name)
			f.ResetConfig()
			fmt.Fprintf(f.IOStreams.Out, "✓ removed %q (default=%q)\n", name, cf.Default)
			return nil
		},
	}
	return cmd
}

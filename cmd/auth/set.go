// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/internal/core"
	"jodoo-cli/internal/credential"
)

func newSetCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		profile     string
		keyFlag     string
		useKeychain bool
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set the API key for a profile (prompted if --api-key not given)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			io := f.IOStreams
			if profile == "" {
				profile = core.DefaultProfile
			}
			key := strings.TrimSpace(keyFlag)
			if key == "" {
				v, err := readSecret(io.In, io.Out, "Jodoo API key (input hidden): ")
				if err != nil {
					return err
				}
				key = strings.TrimSpace(v)
			}
			if key == "" {
				return fmt.Errorf("API key cannot be empty")
			}

			cf, err := core.LoadFile()
			if err != nil {
				return err
			}
			entry := cf.Profiles[profile]
			if entry == nil {
				entry = &core.CliConfig{Profile: profile, BaseURL: core.DefaultBaseURL}
			}
			if useKeychain {
				if err := credential.SetInKeychain(profile, key); err != nil {
					return fmt.Errorf("write keychain: %w", err)
				}
				entry.APIKey = "" // make sure config file does not also hold it
				fmt.Fprintf(io.Out, "✓ stored API key in OS keychain for profile %q\n", profile)
			} else {
				entry.APIKey = key
				fmt.Fprintf(io.Out, "✓ stored API key in config.json for profile %q\n", profile)
			}
			cf.Profiles[profile] = entry
			if cf.Default == "" {
				cf.Default = profile
			}
			if err := core.SaveFile(cf); err != nil {
				return err
			}
			f.ResetConfig()
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "profile to update (default: \"default\")")
	cmd.Flags().StringVar(&keyFlag, "api-key", "", "the API key (otherwise prompted)")
	cmd.Flags().BoolVar(&useKeychain, "use-keychain", false, "store in OS keychain instead of config.json")
	return cmd
}

func readSecret(in interface{}, out interface{ Write([]byte) (int, error) }, label string) (string, error) {
	fmt.Fprint(out, label)
	if file, ok := in.(*os.File); ok {
		fd := int(file.Fd())
		if term.IsTerminal(fd) {
			b, err := term.ReadPassword(fd)
			fmt.Fprintln(out)
			if err != nil {
				return "", err
			}
			return string(b), nil
		}
	}
	if rdr, ok := in.(interface{ Read(p []byte) (int, error) }); ok {
		s := bufio.NewScanner(stdin{rdr})
		if s.Scan() {
			return s.Text(), nil
		}
		if err := s.Err(); err != nil {
			return "", err
		}
	}
	return "", nil
}

type stdin struct{ r interface{ Read(p []byte) (int, error) } }

func (s stdin) Read(p []byte) (int, error) { return s.r.Read(p) }

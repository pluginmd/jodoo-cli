// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package config

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

func newInitCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		apiKeyFlag     string
		baseURLFlag    string
		profileName    string
		useKeychain    bool
		nonInteractive bool
		notes          string
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Configure an API key (interactive or via flags)",
		Long: `Configure jodoo-cli for first use.

Examples:
    jodoo-cli config init                    # interactive prompts
    jodoo-cli config init --api-key sk-...   # one-shot, profile=default
    jodoo-cli config init --profile prod --api-key sk-...
    jodoo-cli config init --use-keychain     # store key in OS keychain`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			io := f.IOStreams
			profile := strings.TrimSpace(profileName)
			if profile == "" {
				profile = core.DefaultProfile
			}
			apiKey := strings.TrimSpace(apiKeyFlag)
			baseURL := strings.TrimSpace(baseURLFlag)

			if apiKey == "" && nonInteractive {
				return fmt.Errorf("--api-key is required in non-interactive mode")
			}
			if apiKey == "" {
				k, err := promptSecret(io.In, io.Out, "Jodoo API key (input hidden): ")
				if err != nil {
					return err
				}
				apiKey = strings.TrimSpace(k)
				if apiKey == "" {
					return fmt.Errorf("API key cannot be empty")
				}
			}
			if baseURL == "" && !nonInteractive {
				v, _ := promptLine(io.In, io.Out,
					fmt.Sprintf("Base URL [%s]: ", core.DefaultBaseURL))
				baseURL = strings.TrimSpace(v)
			}
			if baseURL == "" {
				baseURL = core.DefaultBaseURL
			}

			cf, err := core.LoadFile()
			if err != nil {
				return err
			}
			if cf.Profiles == nil {
				cf.Profiles = map[string]*core.CliConfig{}
			}
			entry := &core.CliConfig{Profile: profile, BaseURL: baseURL, Notes: notes}
			if useKeychain {
				if err := credential.SetInKeychain(profile, apiKey); err != nil {
					return fmt.Errorf("write keychain: %w", err)
				}
				fmt.Fprintf(io.Out, "✓ API key stored in OS keychain (service=%s, account=%s)\n",
					credential.KeychainService, credential.KeychainAccount(profile))
			} else {
				entry.APIKey = apiKey
			}
			cf.Profiles[profile] = entry
			if cf.Default == "" {
				cf.Default = profile
			}
			if err := core.SaveFile(cf); err != nil {
				return err
			}
			path, _ := core.ConfigPath()
			fmt.Fprintf(io.Out, "✓ profile %q saved to %s\n", profile, path)
			fmt.Fprintf(io.Out, "  base URL: %s\n", baseURL)
			if useKeychain {
				fmt.Fprintln(io.Out, "  api key:  (in keychain)")
			} else {
				fmt.Fprintf(io.Out, "  api key:  %s\n", maskKey(apiKey))
			}
			fmt.Fprintln(io.Out, "Next: run `jodoo-cli doctor` to verify connectivity.")
			return nil
		},
	}
	cmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "Jodoo API key (otherwise prompted)")
	cmd.Flags().StringVar(&baseURLFlag, "base-url", "", "API base URL (defaults to "+core.DefaultBaseURL+")")
	cmd.Flags().StringVar(&profileName, "profile", "", "profile name (defaults to \"default\")")
	cmd.Flags().StringVar(&notes, "notes", "", "free-form description for this profile")
	cmd.Flags().BoolVar(&useKeychain, "use-keychain", false, "store API key in OS keychain instead of config.json")
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "fail instead of prompting for missing values")
	return cmd
}

// promptSecret reads a hidden line from a TTY; falls back to a plain
// bufio.Reader when stdin is not a terminal (so piped input still works).
func promptSecret(in interface{}, out interface{ Write([]byte) (int, error) }, label string) (string, error) {
	fmt.Fprint(out, label)
	if f, ok := in.(*os.File); ok {
		fd := int(f.Fd())
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
		s := bufio.NewScanner(readerWrap{rdr})
		if s.Scan() {
			return s.Text(), nil
		}
		if err := s.Err(); err != nil {
			return "", err
		}
	}
	return "", nil
}

func promptLine(in interface{}, out interface{ Write([]byte) (int, error) }, label string) (string, error) {
	fmt.Fprint(out, label)
	if rdr, ok := in.(interface{ Read(p []byte) (int, error) }); ok {
		s := bufio.NewScanner(readerWrap{rdr})
		if s.Scan() {
			return s.Text(), nil
		}
		if err := s.Err(); err != nil {
			return "", err
		}
	}
	return "", nil
}

type readerWrap struct {
	r interface{ Read(p []byte) (int, error) }
}

func (w readerWrap) Read(p []byte) (int, error) { return w.r.Read(p) }

func maskKey(k string) string {
	if len(k) <= 6 {
		return "***"
	}
	return k[:3] + "***" + k[len(k)-3:]
}

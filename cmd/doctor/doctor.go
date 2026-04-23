// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package doctor implements `jodoo-cli doctor` — a connectivity check.
package doctor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"jodoo-cli/internal/build"
	"jodoo-cli/internal/client"
	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/internal/core"
	"jodoo-cli/internal/credential"
	"jodoo-cli/internal/output"
)

// NewCmdDoctor builds the `doctor` command.
func NewCmdDoctor(f *cmdutil.Factory) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Verify config, credentials, and Jodoo API connectivity",
		RunE: func(cmd *cobra.Command, _ []string) error {
			results := []check{
				checkVersion(),
				checkConfigFile(),
				checkProfile(f),
				checkAPIKey(f),
				checkConnectivity(cmd.Context(), f),
			}
			if jsonOut {
				output.PrintJson(f.IOStreams.Out, map[string]interface{}{
					"results": resultsAsJSON(results),
					"ok":      allOK(results),
				})
				if !allOK(results) {
					return output.New(output.ExitGeneric, "doctor", "one or more checks failed")
				}
				return nil
			}
			out := f.IOStreams.Out
			for _, r := range results {
				icon := "✓"
				if !r.OK {
					icon = "✗"
				}
				fmt.Fprintf(out, "%s %-22s %s\n", icon, r.Name, r.Message)
				if !r.OK && r.Hint != "" {
					fmt.Fprintf(out, "    hint: %s\n", r.Hint)
				}
			}
			if !allOK(results) {
				return output.New(output.ExitGeneric, "doctor", "one or more checks failed")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "machine-readable JSON output")
	return cmd
}

type check struct {
	Name    string
	OK      bool
	Message string
	Hint    string
}

func resultsAsJSON(rs []check) []map[string]interface{} {
	out := make([]map[string]interface{}, len(rs))
	for i, r := range rs {
		out[i] = map[string]interface{}{
			"name":    r.Name,
			"ok":      r.OK,
			"message": r.Message,
		}
		if r.Hint != "" {
			out[i]["hint"] = r.Hint
		}
	}
	return out
}

func allOK(rs []check) bool {
	for _, r := range rs {
		if !r.OK {
			return false
		}
	}
	return true
}

func checkVersion() check {
	return check{Name: "version", OK: true, Message: fmt.Sprintf("jodoo-cli %s (built %s)", build.Version, build.Date)}
}

func checkConfigFile() check {
	path, _ := core.ConfigPath()
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return check{Name: "config file", OK: false,
				Message: fmt.Sprintf("not found: %s", path),
				Hint:    "run `jodoo-cli config init`"}
		}
		return check{Name: "config file", OK: false, Message: err.Error()}
	}
	return check{Name: "config file", OK: true, Message: path}
}

func checkProfile(f *cmdutil.Factory) check {
	cf, err := core.LoadFile()
	if err != nil {
		return check{Name: "profile", OK: false, Message: err.Error()}
	}
	name := core.ResolveProfileName(f.ProfileOverride, cf)
	if _, ok := cf.Profiles[name]; !ok && name != core.DefaultProfile {
		return check{Name: "profile", OK: false,
			Message: fmt.Sprintf("profile %q not in config", name),
			Hint:    "run `jodoo-cli config init --profile " + name + "`"}
	}
	return check{Name: "profile", OK: true, Message: name}
}

func checkAPIKey(f *cmdutil.Factory) check {
	f.ResetConfig()
	cfg, err := f.Config()
	if err == nil && cfg.APIKey != "" {
		return check{Name: "api key", OK: true, Message: "configured (" + maskKey(cfg.APIKey) + ")"}
	}
	// fall back to keychain probe
	cf, _ := core.LoadFile()
	prof := core.ResolveProfileName(f.ProfileOverride, cf)
	if v, _ := credential.GetFromKeychain(prof); v != "" {
		return check{Name: "api key", OK: true, Message: "configured in keychain (" + maskKey(v) + ")"}
	}
	return check{Name: "api key", OK: false,
		Message: "no API key resolved",
		Hint:    "run `jodoo-cli auth set` or export JODOO_API_KEY"}
}

func checkConnectivity(ctx context.Context, f *cmdutil.Factory) check {
	cfg, err := f.Config()
	if err != nil {
		return check{Name: "connectivity", OK: false, Message: "skipped (no api key)"}
	}
	c := client.New(cfg)
	c.HTTP.Timeout = 10 * time.Second

	cctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	resp, err := c.Do(cctx, client.Request{Path: "/v5/app/list", Body: map[string]interface{}{"limit": 1, "skip": 0}})
	if err != nil {
		var exitErr *output.ExitError
		if errors.As(err, &exitErr) {
			msg := exitErr.Detail.Message
			hint := exitErr.Detail.Hint
			return check{Name: "connectivity", OK: false, Message: msg, Hint: hint}
		}
		return check{Name: "connectivity", OK: false, Message: err.Error()}
	}
	count := 0
	if v, ok := resp.Data["apps"].([]interface{}); ok {
		count = len(v)
	}
	return check{Name: "connectivity", OK: true,
		Message: fmt.Sprintf("ok (HTTP %d, sample apps=%d, base=%s)", resp.Status, count, cfg.BaseURL)}
}

func maskKey(k string) string {
	if len(k) <= 6 {
		return "***"
	}
	return k[:3] + "***" + k[len(k)-3:]
}

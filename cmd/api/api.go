// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package api implements `jodoo-cli api <path>` — the raw escape hatch
// for endpoints not yet wrapped as a shortcut.
//
// Every Jodoo endpoint is HTTP POST with a JSON body, so unlike the
// Lark / basecli `api` command we don't need a method positional. The
// path is the only positional argument.
package api

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"jodoo-cli/internal/client"
	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/internal/output"
	"jodoo-cli/internal/validate"
)

// NewCmdApi builds the `api` command.
func NewCmdApi(f *cmdutil.Factory) *cobra.Command {
	var (
		dataFlag   string
		dryRun     bool
		jqExpr     string
		formatFlag string
		raw        bool
	)
	cmd := &cobra.Command{
		Use:   "api <path>",
		Short: "Send a raw POST to any Jodoo API endpoint",
		Long: `Send a raw POST to any Jodoo endpoint.

Every Jodoo endpoint is HTTP POST with a JSON body. The path is appended
to the configured base URL (default: ` + "https://api.jodoo.com/api" + `).

Examples:
    jodoo-cli api /v5/app/list --data '{"limit":100,"skip":0}'
    jodoo-cli api /v5/app/entry/data/list \
        --data '{"app_id":"...","entry_id":"...","limit":50}'
    jodoo-cli api /v5/corp/member/get --data '{"username":"alice"}' --jq '.user'
    jodoo-cli api /v5/app/list --data @./payload.json
    jodoo-cli api /v5/app/list --data - <<< '{"limit":10}'

Use --dry-run to inspect the request without sending it.
Use --raw to print the full envelope (including code/msg) instead of just data.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "http") {
				path = "/" + path
			}
			body, err := loadBodyFlag(f.IOStreams.In, dataFlag)
			if err != nil {
				return err
			}
			if err := output.ValidateJqFlags(jqExpr, "", formatFlag); err != nil {
				return output.ErrValidation("%v", err)
			}

			cfg, err := f.MustConfig()
			if err != nil {
				return err
			}
			c := client.New(cfg)

			req := client.Request{Path: path, Body: body}
			if dryRun {
				dry, err := c.BuildDryRun(req)
				if err != nil {
					return err
				}
				output.PrintJson(f.IOStreams.Out, dry)
				return nil
			}

			resp, err := c.Do(cmd.Context(), req)
			if err != nil {
				// Mark as Raw so the root error handler doesn't strip the
				// raw envelope detail (the user asked for the raw API).
				if exitErr, ok := err.(*output.ExitError); ok {
					exitErr.Raw = true
				}
				return err
			}

			payload := interface{}(resp.PayloadOnly())
			if raw {
				payload = resp.Data
			}

			if formatFlag != "" && formatFlag != "json" && formatFlag != "pretty" {
				fm, _ := output.ParseFormat(formatFlag)
				output.FormatValue(f.IOStreams.Out, payload, fm)
				return nil
			}
			return output.WriteEnvelope(f.IOStreams.Out, payload, nil, jqExpr)
		},
	}
	cmd.Flags().StringVar(&dataFlag, "data", "", "request body JSON (use @path for file, - for stdin)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print request without executing")
	cmd.Flags().StringVarP(&jqExpr, "jq", "q", "", "jq expression to filter JSON output")
	cmd.Flags().StringVar(&formatFlag, "format", "json", "output format: json | pretty | table | ndjson | csv")
	cmd.Flags().BoolVar(&raw, "raw", false, "include the full envelope (code, msg) in output")
	return cmd
}

// loadBodyFlag resolves the --data flag, supporting @file and - (stdin).
// Returns nil for an empty value (Jodoo accepts an empty `{}` body for
// some endpoints — the client will inject `{}` itself).
func loadBodyFlag(in io.Reader, raw string) (interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var bytes []byte
	switch {
	case raw == "-":
		b, err := io.ReadAll(in)
		if err != nil {
			return nil, output.ErrValidation("read --data from stdin: %v", err)
		}
		bytes = b
	case strings.HasPrefix(raw, "@@"):
		bytes = []byte(raw[1:]) // literal escape
	case strings.HasPrefix(raw, "@"):
		path := strings.TrimSpace(raw[1:])
		safe, err := validate.SafeInputPath(path)
		if err != nil {
			return nil, output.ErrValidation("invalid --data file path: %v", err)
		}
		b, err := os.ReadFile(safe)
		if err != nil {
			return nil, output.ErrValidation("read --data file %s: %v", path, err)
		}
		bytes = b
	default:
		bytes = []byte(raw)
	}
	if len(bytes) == 0 {
		return nil, nil
	}
	var v interface{}
	if err := json.Unmarshal(bytes, &v); err != nil {
		return nil, output.ErrValidation("--data is not valid JSON: %v", err)
	}
	if v == nil {
		return nil, nil
	}
	return v, nil
}

// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"jodoo-cli/internal/client"
	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/internal/core"
	"jodoo-cli/internal/output"
	"jodoo-cli/internal/validate"
)

// RuntimeContext is the per-invocation context handed to a Shortcut's
// Validate / DryRun / Execute hooks.
type RuntimeContext struct {
	ctx       context.Context
	Config    *core.CliConfig
	Cmd       *cobra.Command
	Format    string
	JqExpr    string
	outputErr error

	Factory   *cmdutil.Factory
	apiClient *client.APIClient
}

// Ctx returns the request-scoped context.
func (r *RuntimeContext) Ctx() context.Context { return r.ctx }

// IO returns the IOStreams from the Factory.
func (r *RuntimeContext) IO() *cmdutil.IOStreams { return r.Factory.IOStreams }

// ── Flag accessors ──

func (r *RuntimeContext) Str(name string) string {
	v, _ := r.Cmd.Flags().GetString(name)
	return v
}
func (r *RuntimeContext) Bool(name string) bool {
	v, _ := r.Cmd.Flags().GetBool(name)
	return v
}
func (r *RuntimeContext) Int(name string) int {
	v, _ := r.Cmd.Flags().GetInt(name)
	return v
}
func (r *RuntimeContext) StrArray(name string) []string {
	v, _ := r.Cmd.Flags().GetStringArray(name)
	return v
}

// ── API helpers ──

func (r *RuntimeContext) getAPIClient() (*client.APIClient, error) {
	if r.apiClient != nil {
		return r.apiClient, nil
	}
	c, err := r.Factory.NewAPIClient()
	if err != nil {
		return nil, err
	}
	c.Config = r.Config
	r.apiClient = c
	return c, nil
}

// CallAPI POSTs to a Jodoo endpoint and returns the payload (envelope
// stripped). Errors propagate as *output.ExitError.
func (r *RuntimeContext) CallAPI(path string, body interface{}) (map[string]interface{}, error) {
	c, err := r.getAPIClient()
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(r.ctx, client.Request{Path: path, Body: body})
	if err != nil {
		return nil, err
	}
	return resp.PayloadOnly(), nil
}

// CallAPIRaw is like CallAPI but returns the full envelope (including
// "code"/"msg"). Useful for endpoints whose response is essentially the
// envelope itself (e.g. workflow action APIs that return only "status").
func (r *RuntimeContext) CallAPIRaw(path string, body interface{}) (map[string]interface{}, error) {
	c, err := r.getAPIClient()
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(r.ctx, client.Request{Path: path, Body: body})
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// PaginateAll loops Jodoo cursor pagination (data_id-based) until the
// server returns less than `pageSize` items in `listKey`. The bodyBuilder
// receives the data_id of the last record; on the first call it is "".
//
// Returns the merged list and the total fetched.
func (r *RuntimeContext) PaginateAll(
	path, listKey string,
	pageSize int,
	bodyBuilder func(lastID string) map[string]interface{},
) ([]interface{}, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	out := make([]interface{}, 0, pageSize)
	lastID := ""
	for {
		body := bodyBuilder(lastID)
		page, err := r.CallAPI(path, body)
		if err != nil {
			return out, err
		}
		arr, _ := page[listKey].([]interface{})
		out = append(out, arr...)
		if len(arr) < pageSize {
			return out, nil
		}
		// pick the _id of the last record as next cursor
		last, ok := arr[len(arr)-1].(map[string]interface{})
		if !ok {
			return out, nil
		}
		var id string
		if v, ok := last["_id"].(string); ok {
			id = v
		} else if v, ok := last["data_id"].(string); ok {
			id = v
		}
		if id == "" || id == lastID {
			return out, nil
		}
		lastID = id
	}
}

// ── Output helpers ──

// Out writes a success envelope (or jq-filtered output).
func (r *RuntimeContext) Out(data interface{}, meta *output.Meta) {
	if err := output.WriteEnvelope(r.IO().Out, data, meta, r.JqExpr); err != nil {
		fmt.Fprintf(r.IO().ErrOut, "error: %v\n", err)
		if r.outputErr == nil {
			r.outputErr = err
		}
	}
}

// OutFormat dispatches to a printer based on --format.
//
// "json" / "" → standard envelope.
// "pretty"    → calls prettyFn(out) if provided, else falls back to json.
// table/csv/ndjson → output.FormatValue.
//
// jqExpr always wins (routes through Out so users can chain --jq).
func (r *RuntimeContext) OutFormat(data interface{}, meta *output.Meta, prettyFn func(w io.Writer)) {
	if r.JqExpr != "" {
		r.Out(data, meta)
		return
	}
	switch r.Format {
	case "pretty":
		if prettyFn != nil {
			prettyFn(r.IO().Out)
		} else {
			r.Out(data, meta)
		}
	case "json", "":
		r.Out(data, meta)
	default:
		fm, ok := output.ParseFormat(r.Format)
		if !ok {
			fmt.Fprintf(r.IO().ErrOut, "warning: unknown format %q, falling back to json\n", r.Format)
		}
		output.FormatValue(r.IO().Out, data, fm)
	}
}

// ── Mounting ──

// Mount registers the shortcut as a child of `parent` using the cobra
// framework. Called by the package-level register.
func (s Shortcut) Mount(parent *cobra.Command, f *cmdutil.Factory) {
	if s.Execute == nil {
		return
	}
	shortcut := s
	cmd := &cobra.Command{
		Use:   shortcut.Command,
		Short: shortcut.Description,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runShortcut(cmd, f, &shortcut)
		},
	}
	registerShortcutFlags(cmd, &shortcut)
	cmdutil.SetTips(cmd, shortcut.Tips)
	parent.AddCommand(cmd)
}

func runShortcut(cmd *cobra.Command, f *cmdutil.Factory, s *Shortcut) error {
	cfg, err := f.MustConfig()
	if err != nil {
		return err
	}
	r := &RuntimeContext{
		ctx:     cmd.Context(),
		Config:  cfg,
		Cmd:     cmd,
		Factory: f,
	}
	if s.HasFormat {
		r.Format = r.Str("format")
		if r.Format != "" {
			if _, ok := output.ParseFormat(r.Format); !ok {
				return output.ErrValidation("invalid --format %q (allowed: json, pretty, table, ndjson, csv)", r.Format)
			}
		}
	}
	r.JqExpr, _ = cmd.Flags().GetString("jq")

	if err := validateEnumFlags(r, s.Flags); err != nil {
		return err
	}
	if err := resolveInputFlags(r, s.Flags); err != nil {
		return err
	}
	if err := output.ValidateJqFlags(r.JqExpr, "", r.Format); err != nil {
		return output.ErrValidation("%v", err)
	}
	if s.Validate != nil {
		if err := s.Validate(r.ctx, r); err != nil {
			return err
		}
	}
	if r.Bool("dry-run") {
		return handleDryRun(f, r, s)
	}
	if s.Risk == "high-risk-write" && !r.Bool("yes") {
		return output.ErrValidation("this is a high-risk write — pass --yes to confirm (%s)", s.Description)
	}
	if err := s.Execute(r.ctx, r); err != nil {
		return err
	}
	return r.outputErr
}

func handleDryRun(f *cmdutil.Factory, r *RuntimeContext, s *Shortcut) error {
	if s.DryRun == nil {
		return output.ErrValidation("--dry-run is not supported for %s %s", s.Service, s.Command)
	}
	fmt.Fprintln(f.IOStreams.ErrOut, "=== Dry Run ===")
	dry := s.DryRun(r.ctx, r)
	if r.JqExpr != "" {
		return output.JqFilter(f.IOStreams.Out, dry, r.JqExpr)
	}
	if r.Format == "pretty" {
		fmt.Fprint(f.IOStreams.Out, dry.Format())
		return nil
	}
	output.PrintJson(f.IOStreams.Out, dry)
	return nil
}

// ── Flag wiring ──

func registerShortcutFlags(cmd *cobra.Command, s *Shortcut) {
	for _, fl := range s.Flags {
		desc := fl.Desc
		if len(fl.Enum) > 0 {
			desc += " (" + strings.Join(fl.Enum, "|") + ")"
		}
		if len(fl.Input) > 0 {
			hints := []string{}
			if slices.Contains(fl.Input, File) {
				hints = append(hints, "@file")
			}
			if slices.Contains(fl.Input, Stdin) {
				hints = append(hints, "- for stdin")
			}
			desc += " (supports " + strings.Join(hints, ", ") + ")"
		}
		switch fl.Type {
		case "bool":
			def := fl.Default == "true"
			cmd.Flags().Bool(fl.Name, def, desc)
		case "int":
			var d int
			fmt.Sscanf(fl.Default, "%d", &d)
			cmd.Flags().Int(fl.Name, d, desc)
		case "string_array":
			cmd.Flags().StringArray(fl.Name, nil, desc)
		default:
			cmd.Flags().String(fl.Name, fl.Default, desc)
		}
		if fl.Hidden {
			_ = cmd.Flags().MarkHidden(fl.Name)
		}
		if fl.Required {
			cmd.MarkFlagRequired(fl.Name)
		}
		if len(fl.Enum) > 0 {
			vals := fl.Enum
			_ = cmd.RegisterFlagCompletionFunc(fl.Name, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return vals, cobra.ShellCompDirectiveNoFileComp
			})
		}
	}

	cmd.Flags().Bool("dry-run", false, "print request without executing")
	if s.HasFormat {
		cmd.Flags().String("format", "json", "output format: json | pretty | table | ndjson | csv")
		_ = cmd.RegisterFlagCompletionFunc("format", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"json", "pretty", "table", "ndjson", "csv"}, cobra.ShellCompDirectiveNoFileComp
		})
	}
	if s.Risk == "high-risk-write" {
		cmd.Flags().Bool("yes", false, "confirm a high-risk operation")
	}
	cmd.Flags().StringP("jq", "q", "", "jq expression to filter JSON output")
}

// validateEnumFlags rejects values not in the declared Enum set.
func validateEnumFlags(r *RuntimeContext, flags []Flag) error {
	for _, fl := range flags {
		if len(fl.Enum) == 0 {
			continue
		}
		v := r.Str(fl.Name)
		if v == "" {
			continue
		}
		ok := false
		for _, allowed := range fl.Enum {
			if v == allowed {
				ok = true
				break
			}
		}
		if !ok {
			return output.ErrValidation("invalid value %q for --%s, allowed: %s", v, fl.Name, strings.Join(fl.Enum, ", "))
		}
	}
	return nil
}

// resolveInputFlags rewrites @path and - (stdin) for flags that opted in.
func resolveInputFlags(r *RuntimeContext, flags []Flag) error {
	stdinUsed := false
	for _, fl := range flags {
		if len(fl.Input) == 0 {
			continue
		}
		raw, err := r.Cmd.Flags().GetString(fl.Name)
		if err != nil {
			return output.ErrValidation("--%s: only string flags support @file/-", fl.Name)
		}
		if raw == "" {
			continue
		}
		if raw == "-" {
			if !slices.Contains(fl.Input, Stdin) {
				return output.ErrValidation("--%s does not support stdin (-)", fl.Name)
			}
			if stdinUsed {
				return output.ErrValidation("--%s: stdin (-) can only be used once", fl.Name)
			}
			stdinUsed = true
			b, err := io.ReadAll(r.IO().In)
			if err != nil {
				return output.ErrValidation("--%s: read stdin: %v", fl.Name, err)
			}
			r.Cmd.Flags().Set(fl.Name, string(b))
			continue
		}
		if strings.HasPrefix(raw, "@@") {
			r.Cmd.Flags().Set(fl.Name, raw[1:])
			continue
		}
		if strings.HasPrefix(raw, "@") {
			if !slices.Contains(fl.Input, File) {
				return output.ErrValidation("--%s does not support file input (@path)", fl.Name)
			}
			path := strings.TrimSpace(raw[1:])
			if path == "" {
				return output.ErrValidation("--%s: empty file path after @", fl.Name)
			}
			safe, err := validate.SafeInputPath(path)
			if err != nil {
				return output.ErrValidation("--%s: invalid file path %q: %v", fl.Name, path, err)
			}
			b, err := readFileLimited(safe, maxInputSize)
			if err != nil {
				return output.ErrValidation("--%s: %v", fl.Name, err)
			}
			r.Cmd.Flags().Set(fl.Name, string(b))
		}
	}
	return nil
}

const maxInputSize = 16 << 20 // 16MB safety cap for @file inputs

// ── Helpers ──

// ParseJSONObject decodes a JSON object string. Returns a friendly error
// when the value is not an object.
func ParseJSONObject(raw, flagName string) (map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, output.ErrValidation("--%s is required (JSON object)", flagName)
	}
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, output.ErrValidation("--%s is not valid JSON: %v", flagName, err)
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil, output.ErrValidation("--%s must be a JSON object, got %T", flagName, v)
	}
	return m, nil
}

// ParseJSONArray decodes a JSON array string.
func ParseJSONArray(raw, flagName string) ([]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, output.ErrValidation("--%s is required (JSON array)", flagName)
	}
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, output.ErrValidation("--%s is not valid JSON: %v", flagName, err)
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil, output.ErrValidation("--%s must be a JSON array, got %T", flagName, v)
	}
	return arr, nil
}

// ParseIntBounded reads an int flag and clamps it into [min, max]. Used
// for limit flags (Jodoo accepts 1–100 in most APIs).
func ParseIntBounded(r *RuntimeContext, name string, min, max int) int {
	v := r.Int(name)
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// readFileLimited reads at most `limit` bytes from a file.
func readFileLimited(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	r := io.LimitReader(f, limit+1)
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errors.New("file is too large (>16MB)")
	}
	return b, nil
}

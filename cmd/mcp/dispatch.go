// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/shortcuts/common"
)

// execTool forks the same binary to execute the underlying shortcut.
//
// Fork-self buys us: identical behaviour to `jodoo-cli jodoo +…` on the
// terminal, full isolation from the MCP server (panics/stdout-leaks cannot
// corrupt JSON-RPC framing), and safe concurrency (the shortcut runner uses
// shared cobra flag storage that is not goroutine-safe).
func execTool(ctx context.Context, f *cmdutil.Factory, t *Tool, args map[string]interface{}) (*toolResult, int, error) {
	argv, err := buildArgv(t.sc, args)
	if err != nil {
		return &toolResult{
			IsError: true,
			Content: []toolContent{{Type: "text", Text: err.Error()}},
		}, -1, nil
	}

	self, err := os.Executable()
	if err != nil {
		return nil, -1, fmt.Errorf("resolve self binary: %w", err)
	}

	cmd := exec.CommandContext(ctx, self, argv...)
	cmd.Env = os.Environ() // inherit JODOO_API_KEY / JODOO_PROFILE etc.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	exit := 0
	if cmd.ProcessState != nil {
		exit = cmd.ProcessState.ExitCode()
	}

	text := stdout.String()
	isErr := runErr != nil || exit != 0
	if isErr {
		// Surface both stderr and whatever stdout carried so the model has
		// enough to explain the failure. Stderr usually holds the envelope
		// error; stdout is normally empty on failure.
		parts := []string{}
		if s := strings.TrimSpace(stderr.String()); s != "" {
			parts = append(parts, s)
		}
		if s := strings.TrimSpace(stdout.String()); s != "" {
			parts = append(parts, s)
		}
		if len(parts) == 0 {
			parts = append(parts, fmt.Sprintf("jodoo-cli exited with code %d", exit))
		}
		text = strings.Join(parts, "\n\n")
	}

	_ = f // reserved for future in-process dispatch (see 07-roadmap.md)
	return &toolResult{
		IsError: isErr,
		Content: []toolContent{{Type: "text", Text: text}},
	}, exit, nil
}

// buildArgv converts MCP arguments into the argv the CLI expects.
//
//	jodoo-cli jodoo +<command> --flag value …
//
// Rules:
//   - bool true  → "--flag"  (CLI treats presence as true)
//   - bool false → omit
//   - string_array → repeated "--flag value"
//   - object / array → JSON-encoded, passed as the string value
//   - dry_run true   → "--dry-run"
//   - confirm true (high-risk) → "--yes"
//   - format → "--format <v>" (only when HasFormat)
//   - jq → "--jq <expr>"
func buildArgv(sc common.Shortcut, args map[string]interface{}) ([]string, error) {
	argv := []string{sc.Service, sc.Command}

	// Index flags by snake_case key for fast lookup.
	byKey := map[string]common.Flag{}
	for _, fl := range sc.Flags {
		byKey[flagKey(fl.Name)] = fl
	}

	// Consume known flags first.
	for key, raw := range args {
		switch key {
		case "dry_run", "confirm", "format", "jq":
			continue // handled below
		}
		fl, known := byKey[key]
		if !known {
			return nil, fmt.Errorf("unknown argument %q for tool %s.%s", key, sc.Service, strings.TrimPrefix(sc.Command, "+"))
		}
		values, err := renderFlagValue(fl, raw)
		if err != nil {
			return nil, err
		}
		for _, v := range values {
			if fl.Type == "bool" {
				if v == "true" {
					argv = append(argv, "--"+fl.Name)
				}
				continue
			}
			argv = append(argv, "--"+fl.Name, v)
		}
	}

	// Synthetic / shared flags.
	if v, _ := args["dry_run"].(bool); v {
		argv = append(argv, "--dry-run")
	}
	if sc.Risk == "high-risk-write" {
		if v, _ := args["confirm"].(bool); v {
			argv = append(argv, "--yes")
		}
	}
	if sc.HasFormat {
		if v, ok := args["format"].(string); ok && v != "" {
			argv = append(argv, "--format", v)
		} else {
			argv = append(argv, "--format", "json")
		}
	}
	if v, ok := args["jq"].(string); ok && v != "" {
		argv = append(argv, "--jq", v)
	}

	// Enforce required flags — catch the common model mistake before it
	// reaches the CLI (cleaner error message).
	for _, fl := range sc.Flags {
		if !fl.Required {
			continue
		}
		if _, present := args[flagKey(fl.Name)]; !present {
			return nil, fmt.Errorf("missing required argument %q", flagKey(fl.Name))
		}
	}

	return argv, nil
}

// renderFlagValue normalises a JSON argument into one or more string values
// the CLI expects on the command line.
func renderFlagValue(fl common.Flag, raw interface{}) ([]string, error) {
	switch fl.Type {
	case "bool":
		switch v := raw.(type) {
		case bool:
			if v {
				return []string{"true"}, nil
			}
			return []string{"false"}, nil
		}
		return nil, fmt.Errorf("argument %q must be boolean", flagKey(fl.Name))

	case "int":
		switch v := raw.(type) {
		case float64:
			return []string{strconv.Itoa(int(v))}, nil
		case int:
			return []string{strconv.Itoa(v)}, nil
		case string:
			// Accept strings for robustness — some clients stringify numbers.
			if _, err := strconv.Atoi(v); err != nil {
				return nil, fmt.Errorf("argument %q must be integer", flagKey(fl.Name))
			}
			return []string{v}, nil
		}
		return nil, fmt.Errorf("argument %q must be integer", flagKey(fl.Name))

	case "string_array":
		arr, ok := raw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("argument %q must be array of strings", flagKey(fl.Name))
		}
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("argument %q: array items must be strings", flagKey(fl.Name))
			}
			out = append(out, s)
		}
		return out, nil

	default: // string
		switch v := raw.(type) {
		case string:
			return []string{v}, nil
		case nil:
			return nil, nil
		}
		// Anything else (object/array/number/bool) gets JSON-encoded. This is
		// how the CLI consumes --data / --filter / --data-list.
		b, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("argument %q: %v", flagKey(fl.Name), err)
		}
		return []string{string(b)}, nil
	}
}

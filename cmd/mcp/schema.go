// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package mcp

import (
	"strings"

	"jodoo-cli/shortcuts/common"
)

// flagKey turns a CLI flag name ("app-id") into an MCP argument key
// ("app_id"). MCP-shipped servers and SDKs lean on snake_case and Claude
// emits snake_case JSON keys more reliably than kebab-case.
func flagKey(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

// flagsToSchema builds the JSON Schema for a shortcut's input. Synthetic
// arguments (dry_run, confirm, format, jq) are appended so they show up in
// the tool manifest alongside the user-declared flags.
func flagsToSchema(s common.Shortcut) map[string]interface{} {
	props := map[string]interface{}{}
	required := []string{}

	for _, fl := range s.Flags {
		props[flagKey(fl.Name)] = flagToProp(fl)
		if fl.Required {
			required = append(required, flagKey(fl.Name))
		}
	}

	// Synthetic args — always safe to add, never required.
	props["dry_run"] = map[string]interface{}{
		"type":        "boolean",
		"description": "preview the HTTP request without executing it",
	}
	props["jq"] = map[string]interface{}{
		"type":        "string",
		"description": "jq expression to filter the JSON result (optional)",
	}
	if s.HasFormat {
		props["format"] = map[string]interface{}{
			"type":        "string",
			"enum":        []string{"json", "pretty", "table", "ndjson", "csv"},
			"description": "output format (default: json)",
		}
	}
	if s.Risk == "high-risk-write" {
		props["confirm"] = map[string]interface{}{
			"type":        "boolean",
			"description": "must be true to execute this high-risk write (maps to --yes)",
		}
	}

	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           props,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// flagToProp converts a single common.Flag to a JSON Schema property.
func flagToProp(fl common.Flag) map[string]interface{} {
	prop := map[string]interface{}{}
	switch fl.Type {
	case "bool":
		prop["type"] = "boolean"
	case "int":
		prop["type"] = "integer"
	case "string_array":
		prop["type"] = "array"
		prop["items"] = map[string]interface{}{"type": "string"}
	default:
		prop["type"] = "string"
	}
	desc := fl.Desc
	if len(fl.Enum) > 0 {
		prop["enum"] = append([]string(nil), fl.Enum...)
	}
	if len(fl.Input) > 0 {
		hints := []string{}
		for _, in := range fl.Input {
			switch in {
			case common.File:
				hints = append(hints, "@path to load from a file")
			case common.Stdin:
				hints = append(hints, "- to read from stdin")
			}
		}
		if len(hints) > 0 {
			if desc != "" {
				desc += " "
			}
			desc += "(supports " + strings.Join(hints, ", ") + ")"
		}
	}
	if desc != "" {
		prop["description"] = desc
	}
	if fl.Default != "" && fl.Type != "bool" && fl.Type != "string_array" {
		// We don't surface booleans with a string default (usually "false"),
		// nor arrays (nil is idiomatic). Strings/ints carry defaults as-is.
		prop["default"] = defaultFor(fl)
	}
	return prop
}

func defaultFor(fl common.Flag) interface{} {
	switch fl.Type {
	case "int":
		var n int
		_, _ = fmtSscanf(fl.Default, &n)
		return n
	default:
		return fl.Default
	}
}

// fmtSscanf is a thin wrapper avoiding the fmt import here — kept separate
// so tests can inspect parsing without coupling to fmt internals.
func fmtSscanf(s string, out *int) (int, error) {
	var n int
	neg := false
	i := 0
	if len(s) > 0 && s[0] == '-' {
		neg = true
		i = 1
	}
	if i == len(s) {
		return 0, nil
	}
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	*out = n
	return 1, nil
}

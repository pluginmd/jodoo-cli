// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/itchyny/gojq"
)

// JqFilter applies a jq expression to v and writes the result(s) to w.
// Each result is printed on its own line as JSON (compact for primitives,
// indented for objects/arrays).
func JqFilter(w io.Writer, v interface{}, expr string) error {
	q, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("parse jq expression: %w", err)
	}
	// Round-trip through JSON so any custom struct (Envelope) becomes a
	// plain map[string]interface{} (gojq operates on Go-native values).
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal for jq: %w", err)
	}
	var native interface{}
	if err := json.Unmarshal(b, &native); err != nil {
		return fmt.Errorf("unmarshal for jq: %w", err)
	}
	iter := q.Run(native)
	for {
		val, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := val.(error); ok {
			return fmt.Errorf("jq runtime: %w", err)
		}
		switch tv := val.(type) {
		case nil:
			fmt.Fprintln(w, "null")
		case string:
			out, _ := json.Marshal(tv)
			fmt.Fprintln(w, string(out))
		case bool, float64, int, int64:
			fmt.Fprintln(w, jsonOneLine(tv))
		default:
			fmt.Fprintln(w, jsonIndent(tv))
		}
	}
	return nil
}

func jsonOneLine(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func jsonIndent(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// ValidateJqFlags checks that --jq is not combined with non-JSON formats.
func ValidateJqFlags(jqExpr, _ string, format string) error {
	if jqExpr == "" {
		return nil
	}
	switch format {
	case "", "json", "pretty":
		return nil
	}
	return fmt.Errorf("--jq cannot be combined with --format %s (use json or pretty)", format)
}

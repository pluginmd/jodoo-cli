// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package output renders command results: JSON envelopes, pretty text,
// table / csv / ndjson, plus an `--jq` filter pipe.
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Envelope is the standard success envelope written to stdout.
type Envelope struct {
	OK     bool                   `json:"ok"`
	Data   interface{}            `json:"data,omitempty"`
	Meta   *Meta                  `json:"meta,omitempty"`
	Notice map[string]interface{} `json:"_notice,omitempty"`
}

// Meta holds optional metadata (pagination, counts) attached to a result.
type Meta struct {
	Count    int  `json:"count,omitempty"`
	HasMore  bool `json:"has_more,omitempty"`
	NextSkip int  `json:"next_skip,omitempty"`
}

// Format is one of the supported output formats.
type Format string

const (
	FormatJSON   Format = "json"
	FormatPretty Format = "pretty"
	FormatNDJSON Format = "ndjson"
	FormatTable  Format = "table"
	FormatCSV    Format = "csv"
)

// ParseFormat normalizes a user-provided format string.
// Returns the format and true if recognized; falls back to JSON on unknown.
func ParseFormat(s string) (Format, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "json":
		return FormatJSON, true
	case "pretty":
		return FormatPretty, true
	case "ndjson":
		return FormatNDJSON, true
	case "table":
		return FormatTable, true
	case "csv":
		return FormatCSV, true
	}
	return FormatJSON, false
}

// PrintJson writes v as indented JSON.
func PrintJson(w io.Writer, v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(w, "%v\n", v)
		return
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

// WriteEnvelope writes a success envelope (with optional jq filtering).
func WriteEnvelope(w io.Writer, data interface{}, meta *Meta, jqExpr string) error {
	env := Envelope{OK: true, Data: data, Meta: meta}
	if jqExpr != "" {
		return JqFilter(w, env, jqExpr)
	}
	PrintJson(w, env)
	return nil
}

// FormatValue renders raw data in a non-JSON format (table/csv/ndjson).
// It accepts either a slice/[]interface{} or a map containing one of the
// known wrapping keys (apps, forms, widgets, data, list, users, departments,
// roles, tasks, members, logs, cc_list, sysWidgets).
func FormatValue(w io.Writer, v interface{}, f Format) {
	rows := extractRows(v)
	switch f {
	case FormatNDJSON:
		writeNDJSON(w, rows)
	case FormatTable:
		writeTable(w, rows)
	case FormatCSV:
		writeCSV(w, rows)
	default:
		PrintJson(w, v)
	}
}

// extractRows tries to find an array of objects to render as rows.
func extractRows(v interface{}) []map[string]interface{} {
	switch t := v.(type) {
	case []interface{}:
		return toMaps(t)
	case []map[string]interface{}:
		return t
	case map[string]interface{}:
		// pick first array-of-object value with a known key
		preferred := []string{
			"apps", "forms", "widgets", "data", "list", "items",
			"users", "departments", "roles", "tasks", "members",
			"logs", "cc_list", "approveCommentList", "data_list",
			"success_ids", "results",
		}
		for _, k := range preferred {
			if arr, ok := t[k].([]interface{}); ok {
				return toMaps(arr)
			}
		}
		// fallback: search any array-of-object field
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if arr, ok := t[k].([]interface{}); ok && len(arr) > 0 {
				if _, ok := arr[0].(map[string]interface{}); ok {
					return toMaps(arr)
				}
			}
		}
		// single object: treat as one-row table
		return []map[string]interface{}{t}
	}
	return nil
}

func toMaps(in []interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(in))
	for _, x := range in {
		if m, ok := x.(map[string]interface{}); ok {
			out = append(out, m)
			continue
		}
		// scalar row → wrap as {"value": x}
		out = append(out, map[string]interface{}{"value": x})
	}
	return out
}

func writeNDJSON(w io.Writer, rows []map[string]interface{}) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, r := range rows {
		_ = enc.Encode(r)
	}
}

func writeTable(w io.Writer, rows []map[string]interface{}) {
	if len(rows) == 0 {
		fmt.Fprintln(w, "(no rows)")
		return
	}
	cols := collectColumns(rows)
	widths := make(map[string]int, len(cols))
	for _, c := range cols {
		widths[c] = len(c)
	}
	cells := make([][]string, len(rows))
	for i, r := range rows {
		cells[i] = make([]string, len(cols))
		for j, c := range cols {
			cells[i][j] = formatCell(r[c])
			if l := len(cells[i][j]); l > widths[c] {
				widths[c] = l
			}
		}
	}
	// header
	for i, c := range cols {
		if i > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprintf(w, "%-*s", widths[c], c)
	}
	fmt.Fprintln(w)
	for i, c := range cols {
		if i > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprint(w, strings.Repeat("-", widths[c]))
	}
	fmt.Fprintln(w)
	for _, row := range cells {
		for i, c := range cols {
			if i > 0 {
				fmt.Fprint(w, "  ")
			}
			fmt.Fprintf(w, "%-*s", widths[c], row[i])
		}
		fmt.Fprintln(w)
	}
}

func writeCSV(w io.Writer, rows []map[string]interface{}) {
	cols := collectColumns(rows)
	cw := csv.NewWriter(w)
	defer cw.Flush()
	_ = cw.Write(cols)
	for _, r := range rows {
		rec := make([]string, len(cols))
		for i, c := range cols {
			rec[i] = formatCell(r[c])
		}
		_ = cw.Write(rec)
	}
}

// collectColumns returns the union of keys across rows, sorted with a
// preference: id-like keys first, then alphabetical.
func collectColumns(rows []map[string]interface{}) []string {
	seen := map[string]bool{}
	for _, r := range rows {
		for k := range r {
			seen[k] = true
		}
	}
	cols := make([]string, 0, len(seen))
	for k := range seen {
		cols = append(cols, k)
	}
	sort.Slice(cols, func(i, j int) bool {
		pi, pj := colPriority(cols[i]), colPriority(cols[j])
		if pi != pj {
			return pi < pj
		}
		return cols[i] < cols[j]
	})
	return cols
}

func colPriority(s string) int {
	switch s {
	case "_id", "id", "data_id", "app_id", "entry_id", "username", "name",
		"dept_no", "role_no", "instance_id", "task_id":
		return 0
	}
	if strings.HasSuffix(s, "_id") || strings.HasSuffix(s, "_no") {
		return 1
	}
	return 2
}

func formatCell(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		// JSON numbers always come back as float64
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	case []interface{}, map[string]interface{}:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		return string(b)
	default:
		return fmt.Sprintf("%v", t)
	}
}

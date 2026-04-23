// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package mcp

import (
	"reflect"
	"strings"
	"testing"

	"jodoo-cli/shortcuts/common"
)

func baseShortcut() common.Shortcut {
	return common.Shortcut{
		Service:   "jodoo",
		Command:   "+data-list",
		HasFormat: true,
		Flags: []common.Flag{
			{Name: "app-id", Required: true},
			{Name: "entry-id", Required: true},
			{Name: "limit", Type: "int", Default: "100"},
			{Name: "fields", Type: "string_array"},
			{Name: "filter"},    // string (JSON-encoded at call site)
			{Name: "sort-desc", Type: "bool"},
		},
	}
}

func argvWants(argv []string, fragments ...[]string) error {
	// Every fragment ([]string) must appear as consecutive elements in argv.
	// We don't care about the order between fragments, only within.
	for _, frag := range fragments {
		if !containsFragment(argv, frag) {
			return &fragmentMissing{frag: frag, argv: argv}
		}
	}
	return nil
}

type fragmentMissing struct {
	frag []string
	argv []string
}

func (e *fragmentMissing) Error() string {
	return "missing fragment " + strings.Join(e.frag, " ") + " in argv: " + strings.Join(e.argv, " ")
}

func containsFragment(argv, frag []string) bool {
outer:
	for i := 0; i+len(frag) <= len(argv); i++ {
		for j := range frag {
			if argv[i+j] != frag[j] {
				continue outer
			}
		}
		return true
	}
	return false
}

func TestBuildArgv_BasicStringAndInt(t *testing.T) {
	sc := baseShortcut()
	argv, err := buildArgv(sc, map[string]interface{}{
		"app_id":   "A",
		"entry_id": "E",
		"limit":    float64(50), // JSON numbers arrive as float64
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if argv[0] != "jodoo" || argv[1] != "+data-list" {
		t.Fatalf("prefix wrong: %v", argv[:2])
	}
	if err := argvWants(argv,
		[]string{"--app-id", "A"},
		[]string{"--entry-id", "E"},
		[]string{"--limit", "50"},
		[]string{"--format", "json"},
	); err != nil {
		t.Error(err)
	}
}

func TestBuildArgv_BoolTrueEmitsFlagWithoutValue(t *testing.T) {
	sc := baseShortcut()
	argv, err := buildArgv(sc, map[string]interface{}{
		"app_id":    "A",
		"entry_id":  "E",
		"sort_desc": true,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// Expect "--sort-desc" (not followed by "true"/"false").
	found := false
	for i, a := range argv {
		if a == "--sort-desc" {
			found = true
			if i+1 < len(argv) && (argv[i+1] == "true" || argv[i+1] == "false") {
				t.Errorf("--sort-desc should not have a value, got %q", argv[i+1])
			}
		}
	}
	if !found {
		t.Errorf("--sort-desc missing from argv: %v", argv)
	}
}

func TestBuildArgv_BoolFalseOmitted(t *testing.T) {
	sc := baseShortcut()
	argv, err := buildArgv(sc, map[string]interface{}{
		"app_id":    "A",
		"entry_id":  "E",
		"sort_desc": false,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	for _, a := range argv {
		if a == "--sort-desc" {
			t.Errorf("bool=false should omit the flag, got argv: %v", argv)
		}
	}
}

func TestBuildArgv_StringArrayBecomesRepeatedFlags(t *testing.T) {
	sc := baseShortcut()
	argv, err := buildArgv(sc, map[string]interface{}{
		"app_id":   "A",
		"entry_id": "E",
		"fields":   []interface{}{"a", "b", "c"},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	got := []string{}
	for i, a := range argv {
		if a == "--fields" && i+1 < len(argv) {
			got = append(got, argv[i+1])
		}
	}
	if !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("--fields values = %v, want [a b c]", got)
	}
}

func TestBuildArgv_ObjectStringFlagGetsJSONEncoded(t *testing.T) {
	sc := baseShortcut()
	argv, err := buildArgv(sc, map[string]interface{}{
		"app_id":   "A",
		"entry_id": "E",
		"filter":   map[string]interface{}{"rel": "and", "cond": []interface{}{}},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// Find --filter; value should be JSON.
	for i, a := range argv {
		if a == "--filter" && i+1 < len(argv) {
			v := argv[i+1]
			if !strings.Contains(v, `"rel":"and"`) {
				t.Errorf("--filter not JSON-encoded: %q", v)
			}
			return
		}
	}
	t.Errorf("--filter missing from argv: %v", argv)
}

func TestBuildArgv_DryRunAndConfirmMapping(t *testing.T) {
	sc := baseShortcut()
	sc.Risk = "high-risk-write"
	argv, err := buildArgv(sc, map[string]interface{}{
		"app_id":   "A",
		"entry_id": "E",
		"dry_run":  true,
		"confirm":  true,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !containsFragment(argv, []string{"--dry-run"}) {
		t.Errorf("missing --dry-run")
	}
	if !containsFragment(argv, []string{"--yes"}) {
		t.Errorf("missing --yes")
	}
}

func TestBuildArgv_ConfirmIgnoredOnNonHighRisk(t *testing.T) {
	sc := baseShortcut() // Risk = "" (default read)
	argv, err := buildArgv(sc, map[string]interface{}{
		"app_id":   "A",
		"entry_id": "E",
		"confirm":  true,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if containsFragment(argv, []string{"--yes"}) {
		t.Errorf("--yes must not appear on non-high-risk shortcut")
	}
}

func TestBuildArgv_MissingRequiredFails(t *testing.T) {
	sc := baseShortcut()
	if _, err := buildArgv(sc, map[string]interface{}{"app_id": "A"}); err == nil {
		t.Errorf("expected error for missing entry_id")
	}
}

func TestBuildArgv_UnknownArgFails(t *testing.T) {
	sc := baseShortcut()
	_, err := buildArgv(sc, map[string]interface{}{
		"app_id":   "A",
		"entry_id": "E",
		"bogus":    "x",
	})
	if err == nil {
		t.Errorf("expected error for unknown arg")
	}
}

func TestBuildArgv_JqPassesThrough(t *testing.T) {
	sc := baseShortcut()
	argv, err := buildArgv(sc, map[string]interface{}{
		"app_id":   "A",
		"entry_id": "E",
		"jq":       ".data.data | length",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !containsFragment(argv, []string{"--jq", ".data.data | length"}) {
		t.Errorf("--jq not passed through: %v", argv)
	}
}

func TestBuildArgv_IntAcceptsStringNumber(t *testing.T) {
	sc := baseShortcut()
	argv, err := buildArgv(sc, map[string]interface{}{
		"app_id":   "A",
		"entry_id": "E",
		"limit":    "75",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !containsFragment(argv, []string{"--limit", "75"}) {
		t.Errorf("--limit 75 missing: %v", argv)
	}
}

func TestBuildArgv_IntRejectsNonNumericString(t *testing.T) {
	sc := baseShortcut()
	if _, err := buildArgv(sc, map[string]interface{}{
		"app_id":   "A",
		"entry_id": "E",
		"limit":    "not-a-number",
	}); err == nil {
		t.Errorf("expected error for non-numeric limit")
	}
}

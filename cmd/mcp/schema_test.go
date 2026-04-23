// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package mcp

import (
	"reflect"
	"sort"
	"testing"

	"jodoo-cli/shortcuts/common"
)

func TestFlagKey_KebabToSnake(t *testing.T) {
	cases := map[string]string{
		"app-id":         "app_id",
		"transaction-id": "transaction_id",
		"limit":          "limit",
		"already_snake":  "already_snake",
	}
	for in, want := range cases {
		if got := flagKey(in); got != want {
			t.Errorf("flagKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFlagToProp_Types(t *testing.T) {
	cases := []struct {
		name string
		in   common.Flag
		want string // expected "type" JSON Schema value
	}{
		{"string default", common.Flag{Name: "x"}, "string"},
		{"bool", common.Flag{Name: "x", Type: "bool"}, "boolean"},
		{"int", common.Flag{Name: "x", Type: "int"}, "integer"},
		{"string_array", common.Flag{Name: "x", Type: "string_array"}, "array"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := flagToProp(tc.in)
			if p["type"] != tc.want {
				t.Errorf("type = %v, want %v", p["type"], tc.want)
			}
			if tc.in.Type == "string_array" {
				items, ok := p["items"].(map[string]interface{})
				if !ok || items["type"] != "string" {
					t.Errorf("string_array items: %v", p["items"])
				}
			}
		})
	}
}

func TestFlagToProp_EnumAndDescription(t *testing.T) {
	fl := common.Flag{
		Name: "status",
		Desc: "workflow status",
		Enum: []string{"active", "closed"},
	}
	p := flagToProp(fl)
	enum, ok := p["enum"].([]string)
	if !ok || !reflect.DeepEqual(enum, []string{"active", "closed"}) {
		t.Fatalf("enum not preserved: %v", p["enum"])
	}
	if p["description"] != "workflow status" {
		t.Errorf("description not preserved: %v", p["description"])
	}
}

func TestFlagToProp_InputHintsFoldIntoDescription(t *testing.T) {
	fl := common.Flag{
		Name:  "data",
		Desc:  "record JSON",
		Input: []string{common.File, common.Stdin},
	}
	p := flagToProp(fl)
	desc, _ := p["description"].(string)
	if desc == "" || desc == "record JSON" {
		t.Fatalf("expected input hints appended, got %q", desc)
	}
}

func TestFlagsToSchema_SyntheticAndRequired(t *testing.T) {
	sc := common.Shortcut{
		Service:   "jodoo",
		Command:   "+data-list",
		HasFormat: true,
		Risk:      "high-risk-write",
		Flags: []common.Flag{
			{Name: "app-id", Required: true, Desc: "app id"},
			{Name: "entry-id", Required: true, Desc: "entry id"},
			{Name: "limit", Type: "int", Default: "100"},
		},
	}
	s := flagsToSchema(sc)

	props := s["properties"].(map[string]interface{})

	// user flags keyed in snake_case
	for _, key := range []string{"app_id", "entry_id", "limit"} {
		if _, ok := props[key]; !ok {
			t.Errorf("missing property %q", key)
		}
	}
	// synthetic args
	for _, key := range []string{"dry_run", "jq", "format", "confirm"} {
		if _, ok := props[key]; !ok {
			t.Errorf("missing synthetic property %q", key)
		}
	}
	// non-high-risk shortcut must NOT get confirm
	scRead := sc
	scRead.Risk = "read"
	pRead := flagsToSchema(scRead)["properties"].(map[string]interface{})
	if _, has := pRead["confirm"]; has {
		t.Errorf("confirm must not exist for non-high-risk shortcut")
	}
	// required list is exactly the required flags (order-insensitive)
	req, _ := s["required"].([]string)
	sort.Strings(req)
	if !reflect.DeepEqual(req, []string{"app_id", "entry_id"}) {
		t.Errorf("required = %v, want [app_id entry_id]", req)
	}
	if s["additionalProperties"] != false {
		t.Errorf("additionalProperties should be false")
	}
}

func TestFlagsToSchema_NoFormatFlagWhenHasFormatFalse(t *testing.T) {
	sc := common.Shortcut{
		Service:   "jodoo",
		Command:   "+file-upload",
		HasFormat: false,
		Flags:     []common.Flag{{Name: "url", Required: true}},
	}
	props := flagsToSchema(sc)["properties"].(map[string]interface{})
	if _, ok := props["format"]; ok {
		t.Errorf("format must not be injected when HasFormat=false")
	}
}

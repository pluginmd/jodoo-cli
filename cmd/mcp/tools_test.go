// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package mcp

import (
	"strings"
	"testing"
)

func TestGlobMatch(t *testing.T) {
	cases := []struct {
		pat, s string
		want   bool
	}{
		{"*", "anything", true},
		{"jodoo.app-list", "jodoo.app-list", true},
		{"jodoo.app-list", "jodoo.data-list", false},
		{"jodoo.*", "jodoo.app-list", true},
		{"jodoo.*", "other.app-list", false},
		{"jodoo.data-*", "jodoo.data-list", true},
		{"jodoo.data-*", "jodoo.data-batch-delete", true},
		{"jodoo.data-*", "jodoo.app-list", false},
		{"*.data-list", "jodoo.data-list", true},
		{"*.data-*", "jodoo.data-batch-create", true},
		{"a*b*c", "abc", true},
		{"a*b*c", "aXXbYYc", true},
		{"a*b*c", "axxccc", false},
	}
	for _, tc := range cases {
		if got := globMatch(tc.pat, tc.s); got != tc.want {
			t.Errorf("globMatch(%q, %q) = %v, want %v", tc.pat, tc.s, got, tc.want)
		}
	}
}

func TestAllowed_EmptyAllowIsAllowAll(t *testing.T) {
	if !allowed("jodoo.data-list", nil, nil) {
		t.Errorf("empty allow should allow everything")
	}
}

func TestAllowed_DenyAfterAllow(t *testing.T) {
	if allowed("jodoo.data-delete", []string{"jodoo.data-*"}, []string{"jodoo.data-delete"}) {
		t.Errorf("deny must win over allow")
	}
	if !allowed("jodoo.data-list", []string{"jodoo.data-*"}, []string{"jodoo.data-delete"}) {
		t.Errorf("allowed item shouldn't be denied")
	}
}

func TestAllowed_AllowRequiresMatch(t *testing.T) {
	if allowed("jodoo.app-list", []string{"jodoo.data-*"}, nil) {
		t.Errorf("must not allow when no pattern matches")
	}
}

func TestBuildRegistry_ReadOnlyFiltersWrites(t *testing.T) {
	reg, err := buildRegistry(&ServeOptions{ReadOnly: true})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, name := range reg.order {
		tool := reg.tools[name]
		if tool.risk != "" && tool.risk != "read" {
			t.Errorf("read-only leaked non-read tool: %s (risk=%s)", name, tool.risk)
		}
	}
	if len(reg.order) == 0 {
		t.Error("read-only registry should still expose read tools")
	}
}

func TestBuildRegistry_AllowDenyFilters(t *testing.T) {
	reg, err := buildRegistry(&ServeOptions{
		Allow: []string{"jodoo.data-*"},
		Deny:  []string{"jodoo.data-delete", "jodoo.data-batch-delete"},
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, name := range reg.order {
		if !strings.HasPrefix(name, "jodoo.data-") {
			t.Errorf("allow leak: %s", name)
		}
		if name == "jodoo.data-delete" || name == "jodoo.data-batch-delete" {
			t.Errorf("deny leak: %s", name)
		}
	}
}

func TestBuildRegistry_DecoratedDescriptionCarriesRisk(t *testing.T) {
	reg, err := buildRegistry(&ServeOptions{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, name := range reg.order {
		tool := reg.tools[name]
		if !strings.Contains(tool.Description, "[risk=") {
			t.Errorf("description missing risk tag: %s → %q", name, tool.Description)
		}
	}
}

func TestBuildRegistry_SortedOrder(t *testing.T) {
	reg, err := buildRegistry(&ServeOptions{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for i := 1; i < len(reg.order); i++ {
		if reg.order[i-1] > reg.order[i] {
			t.Errorf("registry order not sorted: %q > %q", reg.order[i-1], reg.order[i])
		}
	}
}

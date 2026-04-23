// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/shortcuts"
	"jodoo-cli/shortcuts/common"
)

// Tool is the MCP-facing view of a shortcut. The underlying shortcut is kept
// for dispatch — registry lookups are by Name.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`

	risk string
	sc   common.Shortcut
}

// Registry is the set of tools exposed by one server instance. It is built
// once at startup and not mutated afterwards.
type Registry struct {
	opts  *ServeOptions
	tools map[string]*Tool
	order []string
}

// buildRegistry walks every provider's shortcuts and folds in allow/deny
// and --read-only filters.
func buildRegistry(opts *ServeOptions) (*Registry, error) {
	reg := &Registry{opts: opts, tools: map[string]*Tool{}}
	for _, p := range shortcuts.Providers() {
		svc := p.Service()
		for _, sc := range p.All() {
			// Skip shortcuts without an Execute hook — they're placeholders.
			if sc.Execute == nil {
				continue
			}
			name := svc + "." + strings.TrimPrefix(sc.Command, "+")
			if opts.ReadOnly && sc.Risk != "" && sc.Risk != "read" {
				continue
			}
			if !allowed(name, opts.Allow, opts.Deny) {
				continue
			}
			reg.tools[name] = &Tool{
				Name:        name,
				Description: decorateDescription(sc),
				InputSchema: flagsToSchema(sc),
				risk:        sc.Risk,
				sc:          sc,
			}
			reg.order = append(reg.order, name)
		}
	}
	sort.Strings(reg.order)
	if len(reg.order) == 0 {
		return nil, fmt.Errorf("mcp: no tools exposed (check --allow/--deny/--read-only)")
	}
	return reg, nil
}

// decorateDescription appends the risk tag so the model sees the stakes
// directly in tool discovery. Cheap but meaningfully improves its picks.
func decorateDescription(sc common.Shortcut) string {
	risk := sc.Risk
	if risk == "" {
		risk = "read"
	}
	return fmt.Sprintf("%s [risk=%s]", sc.Description, risk)
}

// allowed applies --allow then --deny. Empty allow = allow all.
func allowed(name string, allow, deny []string) bool {
	if len(allow) > 0 {
		match := false
		for _, pat := range allow {
			if globMatch(pat, name) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	for _, pat := range deny {
		if globMatch(pat, name) {
			return false
		}
	}
	return true
}

// globMatch supports '*' as a wildcard (any number of chars). Good enough
// for tool names; intentionally not a full glob library.
func globMatch(pattern, s string) bool {
	if pattern == "*" {
		return true
	}
	// Split pattern by '*' and match segments in order.
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == s
	}
	if !strings.HasPrefix(s, parts[0]) {
		return false
	}
	s = s[len(parts[0]):]
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(s, parts[i])
		if idx < 0 {
			return false
		}
		s = s[idx+len(parts[i]):]
	}
	return strings.HasSuffix(s, parts[len(parts)-1])
}

// handleToolsList returns the tool manifest in MCP-spec shape.
func handleToolsList(reg *Registry, req *rpcRequest) *rpcResponse {
	list := make([]*Tool, 0, len(reg.order))
	for _, name := range reg.order {
		list = append(list, reg.tools[name])
	}
	return ok(req, map[string]interface{}{"tools": list})
}

// toolCallParams mirrors the MCP tools/call params schema.
type toolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// toolContent is a single text content block inside a tool result.
type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

func handleToolsCall(ctx context.Context, log Logger, f *cmdutil.Factory, reg *Registry, req *rpcRequest) *rpcResponse {
	var p toolCallParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return rpcErr(req, -32602, "invalid params", err.Error())
	}
	t, ok2 := reg.tools[p.Name]
	if !ok2 {
		return rpcErr(req, -32602, "unknown tool", p.Name)
	}

	// High-risk gate enforced before we even fork the child.
	if t.risk == "high-risk-write" {
		if v, _ := p.Arguments["confirm"].(bool); !v {
			logCall(log, t.Name, -1, 0, true)
			return ok(req, &toolResult{
				IsError: true,
				Content: []toolContent{{Type: "text", Text: fmt.Sprintf("%s is a high-risk write — pass \"confirm\": true to execute", t.Name)}},
			})
		}
	}

	start := time.Now()
	res, exit, err := execTool(ctx, f, t, p.Arguments)
	dur := time.Since(start)
	logCall(log, t.Name, exit, dur, err != nil || res.IsError)

	if err != nil {
		return rpcErr(req, -32000, "tool dispatch failed", err.Error())
	}
	return ok(req, res)
}

// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package mcp

import (
	"context"
	"encoding/json"
	"testing"
)

// silentLogger captures nothing — used to keep tests quiet.
type silentLogger struct{}

func (silentLogger) Printf(string, ...interface{}) {}

func mustRawID(t *testing.T, v interface{}) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal id: %v", err)
	}
	return json.RawMessage(b)
}

func TestDispatch_Initialize(t *testing.T) {
	reg, err := buildRegistry(&ServeOptions{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	req := &rpcRequest{JSONRPC: "2.0", ID: mustRawID(t, 1), Method: "initialize"}
	resp := dispatch(context.Background(), silentLogger{}, nil, reg, req)
	if resp == nil || resp.Error != nil {
		t.Fatalf("initialize failed: %+v", resp)
	}
	res, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("result shape: %T", resp.Result)
	}
	if res["protocolVersion"] != mcpProtocolVersion {
		t.Errorf("protocol version mismatch: %v", res["protocolVersion"])
	}
}

func TestDispatch_Notification_NoResponse(t *testing.T) {
	reg, _ := buildRegistry(&ServeOptions{})
	// Notification has no ID.
	req := &rpcRequest{JSONRPC: "2.0", Method: "notifications/initialized"}
	resp := dispatch(context.Background(), silentLogger{}, nil, reg, req)
	if resp != nil {
		t.Errorf("notification must not produce a response, got %+v", resp)
	}
}

func TestDispatch_UnknownMethod_ReturnsMethodNotFound(t *testing.T) {
	reg, _ := buildRegistry(&ServeOptions{})
	req := &rpcRequest{JSONRPC: "2.0", ID: mustRawID(t, 1), Method: "no/such"}
	resp := dispatch(context.Background(), silentLogger{}, nil, reg, req)
	if resp == nil || resp.Error == nil {
		t.Fatalf("expected error response, got %+v", resp)
	}
	if resp.Error.Code != -32601 {
		t.Errorf("want -32601, got %d", resp.Error.Code)
	}
}

func TestDispatch_UnknownNotification_SilentlyIgnored(t *testing.T) {
	reg, _ := buildRegistry(&ServeOptions{})
	req := &rpcRequest{JSONRPC: "2.0", Method: "notifications/whatever"}
	resp := dispatch(context.Background(), silentLogger{}, nil, reg, req)
	if resp != nil {
		t.Errorf("unknown notification should be ignored, got %+v", resp)
	}
}

func TestDispatch_Ping(t *testing.T) {
	reg, _ := buildRegistry(&ServeOptions{})
	req := &rpcRequest{JSONRPC: "2.0", ID: mustRawID(t, 2), Method: "ping"}
	resp := dispatch(context.Background(), silentLogger{}, nil, reg, req)
	if resp == nil || resp.Error != nil {
		t.Fatalf("ping failed: %+v", resp)
	}
}

func TestDispatch_ToolsList_MatchesRegistrySize(t *testing.T) {
	reg, _ := buildRegistry(&ServeOptions{})
	req := &rpcRequest{JSONRPC: "2.0", ID: mustRawID(t, 3), Method: "tools/list"}
	resp := dispatch(context.Background(), silentLogger{}, nil, reg, req)
	if resp == nil || resp.Error != nil {
		t.Fatalf("tools/list failed: %+v", resp)
	}
	res := resp.Result.(map[string]interface{})
	list := res["tools"].([]*Tool)
	if len(list) != len(reg.order) {
		t.Errorf("tools/list returned %d, registry has %d", len(list), len(reg.order))
	}
}

func TestDispatch_ToolsCall_HighRiskWithoutConfirm_IsErrorNoFork(t *testing.T) {
	reg, _ := buildRegistry(&ServeOptions{})
	// Pick a known high-risk tool.
	name := "jodoo.data-delete"
	if _, ok := reg.tools[name]; !ok {
		t.Fatalf("test assumes %s exists in registry", name)
	}
	params, _ := json.Marshal(map[string]interface{}{
		"name":      name,
		"arguments": map[string]interface{}{"app_id": "A", "entry_id": "E", "data_id": "D"},
	})
	req := &rpcRequest{
		JSONRPC: "2.0",
		ID:      mustRawID(t, 4),
		Method:  "tools/call",
		Params:  params,
	}
	resp := dispatch(context.Background(), silentLogger{}, nil, reg, req)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected successful JSON-RPC envelope with tool-level error, got %+v", resp)
	}
	tr, ok := resp.Result.(*toolResult)
	if !ok {
		t.Fatalf("result shape: %T", resp.Result)
	}
	if !tr.IsError {
		t.Errorf("expected isError=true for missing confirm")
	}
	if len(tr.Content) == 0 || tr.Content[0].Text == "" {
		t.Errorf("expected explanatory content, got %+v", tr.Content)
	}
}

func TestDispatch_ToolsCall_UnknownTool(t *testing.T) {
	reg, _ := buildRegistry(&ServeOptions{})
	params, _ := json.Marshal(map[string]interface{}{"name": "jodoo.nope", "arguments": map[string]interface{}{}})
	req := &rpcRequest{JSONRPC: "2.0", ID: mustRawID(t, 5), Method: "tools/call", Params: params}
	resp := dispatch(context.Background(), silentLogger{}, nil, reg, req)
	if resp == nil || resp.Error == nil || resp.Error.Code != -32602 {
		t.Fatalf("expected -32602 invalid params, got %+v", resp)
	}
}

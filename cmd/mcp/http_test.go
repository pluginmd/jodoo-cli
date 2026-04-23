// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package mcp

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer spins up an httptest.Server wired to our MCP handlers.
// The registry is real (walks shortcuts.Providers()) but no API calls are
// made because we never hit tools/call with confirm.
func newTestServer(t *testing.T, token string) (*httptest.Server, *Registry) {
	t.Helper()
	reg, err := buildRegistry(&ServeOptions{})
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/info", handleInfo(reg))
	mux.HandleFunc("/mcp", withAuth(token, handleMCP(nil, silentLogger{}, nil, reg)))
	srv := httptest.NewServer(withAccessLog(silentLogger{}, mux))
	t.Cleanup(srv.Close)
	return srv, reg
}

func TestHTTP_Healthz_OpenNoAuth(t *testing.T) {
	srv, _ := newTestServer(t, "secret")
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("healthz = %d, want 200", resp.StatusCode)
	}
}

func TestHTTP_Info_ReportsToolCount(t *testing.T) {
	srv, reg := newTestServer(t, "")
	resp, err := http.Get(srv.URL + "/info")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var info map[string]interface{}
	if err := json.Unmarshal(b, &info); err != nil {
		t.Fatalf("decode /info: %v (%s)", err, b)
	}
	want := float64(len(reg.order))
	if info["tools"] != want {
		t.Errorf("tools = %v, want %v", info["tools"], want)
	}
}

func TestHTTP_Mcp_RequiresBearer(t *testing.T) {
	srv, _ := newTestServer(t, "secret")
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Errorf("no token = %d, want 401", resp.StatusCode)
	}
	if got := resp.Header.Get("WWW-Authenticate"); !strings.Contains(got, "Bearer") {
		t.Errorf("missing WWW-Authenticate: %q", got)
	}
}

func TestHTTP_Mcp_AcceptsCorrectBearer(t *testing.T) {
	srv, _ := newTestServer(t, "secret")
	req, _ := http.NewRequest("POST", srv.URL+"/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body=%s", resp.StatusCode, b)
	}
	var out rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Error != nil {
		t.Errorf("ping errored: %+v", out.Error)
	}
}

func TestHTTP_Mcp_GetReturnsMethodNotAllowed(t *testing.T) {
	srv, _ := newTestServer(t, "")
	resp, err := http.Get(srv.URL + "/mcp")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET /mcp = %d, want 405", resp.StatusCode)
	}
}

func TestHTTP_Mcp_BadJSONReturnsParseError(t *testing.T) {
	srv, _ := newTestServer(t, "")
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader("{not json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200 (parse error rides on 200)", resp.StatusCode)
	}
	var out rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Error == nil || out.Error.Code != -32700 {
		t.Errorf("expected -32700, got %+v", out.Error)
	}
}

func TestHTTP_Mcp_NotificationReturns204(t *testing.T) {
	srv, _ := newTestServer(t, "")
	// Notification = no id
	resp, err := http.Post(srv.URL+"/mcp", "application/json",
		strings.NewReader(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		t.Errorf("notification = %d, want 204", resp.StatusCode)
	}
}

func TestConstantTimeEqual(t *testing.T) {
	if !constantTimeEqual("abc", "abc") {
		t.Error("equal strings should match")
	}
	if constantTimeEqual("abc", "abd") {
		t.Error("different strings should not match")
	}
	if constantTimeEqual("abc", "abcd") {
		t.Error("different length should not match")
	}
	if !constantTimeEqual("", "") {
		t.Error("empty strings should match")
	}
}

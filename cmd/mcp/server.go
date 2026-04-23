// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"jodoo-cli/internal/build"
	"jodoo-cli/internal/cmdutil"
)

// Protocol version we negotiate with the client. Kept as a constant — when
// MCP publishes a newer revision we bump this and test the capability shape.
const mcpProtocolVersion = "2025-06-18"

// rpcRequest / rpcResponse are the minimal JSON-RPC 2.0 envelopes we need.
// ID is json.RawMessage so we round-trip it exactly (number or string).
type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Serve dispatches to the stdio or HTTP transport loop.
func Serve(ctx context.Context, f *cmdutil.Factory, opts *ServeOptions) error {
	if opts.HTTPAddr != "" {
		return serveHTTP(ctx, f, opts)
	}
	return serveStdio(ctx, f, opts)
}

func serveStdio(ctx context.Context, f *cmdutil.Factory, opts *ServeOptions) error {
	logger, closeLog, err := newLogger(opts.LogFile, f.IOStreams.ErrOut)
	if err != nil {
		return fmt.Errorf("open --log-file: %w", err)
	}
	defer closeLog()

	reg, err := buildRegistry(opts)
	if err != nil {
		return err
	}
	logger.Printf("mcp server starting (tools=%d protocol=%s version=%s)", len(reg.order), mcpProtocolVersion, build.Version)

	in := bufio.NewScanner(f.IOStreams.In)
	in.Buffer(make([]byte, 1<<20), 64<<20) // allow large tool-call payloads

	out := f.IOStreams.Out
	var outMu sync.Mutex
	writeFrame := func(v interface{}) {
		b, err := json.Marshal(v)
		if err != nil {
			logger.Printf("marshal response: %v", err)
			return
		}
		outMu.Lock()
		defer outMu.Unlock()
		_, _ = out.Write(b)
		_, _ = out.Write([]byte("\n"))
	}

	var wg sync.WaitGroup
	for in.Scan() {
		line := bytes.TrimSpace(in.Bytes())
		if len(line) == 0 {
			continue
		}
		frame := make([]byte, len(line))
		copy(frame, line)

		var req rpcRequest
		if err := json.Unmarshal(frame, &req); err != nil {
			writeFrame(&rpcResponse{JSONRPC: "2.0", Error: &rpcError{Code: -32700, Message: "parse error", Data: err.Error()}})
			continue
		}
		if req.JSONRPC != "2.0" && req.JSONRPC != "" {
			writeFrame(&rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32600, Message: "invalid request", Data: "jsonrpc must be \"2.0\""}})
			continue
		}

		wg.Add(1)
		go func(req rpcRequest) {
			defer wg.Done()
			resp := dispatch(ctx, logger, f, reg, &req)
			if req.ID == nil || resp == nil {
				return // notification
			}
			writeFrame(resp)
		}(req)
	}

	// Client closed the connection — wait for in-flight calls to finish.
	wg.Wait()
	if err := in.Err(); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	logger.Printf("mcp server stopped")
	return nil
}

// dispatch routes a parsed request to the right handler. Returns nil when
// the request was a notification (no response expected).
func dispatch(ctx context.Context, log Logger, f *cmdutil.Factory, reg *Registry, req *rpcRequest) *rpcResponse {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in dispatch method=%s: %v", req.Method, r)
		}
	}()

	switch req.Method {
	case "initialize":
		return handleInitialize(req)
	case "notifications/initialized":
		return nil
	case "ping":
		return ok(req, struct{}{})
	case "tools/list":
		return handleToolsList(reg, req)
	case "tools/call":
		return handleToolsCall(ctx, log, f, reg, req)
	case "shutdown":
		return ok(req, struct{}{})
	default:
		if req.ID == nil {
			// Unknown notification — per spec, silently ignore.
			return nil
		}
		return rpcErr(req, -32601, "method not found", req.Method)
	}
}

func handleInitialize(req *rpcRequest) *rpcResponse {
	return ok(req, map[string]interface{}{
		"protocolVersion": mcpProtocolVersion,
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{"listChanged": false},
		},
		"serverInfo": map[string]interface{}{
			"name":    "jodoo-cli",
			"version": build.Version,
		},
	})
}

func ok(req *rpcRequest, result interface{}) *rpcResponse {
	return &rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: result}
}

func rpcErr(req *rpcRequest, code int, msg string, data interface{}) *rpcResponse {
	return &rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: code, Message: msg, Data: data}}
}


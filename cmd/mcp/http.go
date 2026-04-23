// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"jodoo-cli/internal/build"
	"jodoo-cli/internal/cmdutil"
)

// serveHTTP implements the MCP Streamable-HTTP transport, minus SSE
// server-to-client streaming (every tool call is request/response).
//
// Endpoints:
//
//	POST /mcp       JSON-RPC request → JSON-RPC response
//	GET  /mcp       405 (SSE stream not implemented — roadmap)
//	GET  /healthz   liveness probe (no auth)
//	GET  /info      server metadata (no auth; no tool list)
//
// Auth: optional Bearer token via --token or MCP_TOKEN env.
func serveHTTP(ctx context.Context, f *cmdutil.Factory, opts *ServeOptions) error {
	logger, closeLog, err := newLogger(opts.LogFile, f.IOStreams.ErrOut)
	if err != nil {
		return fmt.Errorf("open --log-file: %w", err)
	}
	defer closeLog()

	reg, err := buildRegistry(opts)
	if err != nil {
		return err
	}

	token := resolveToken(opts)
	if token == "" {
		logger.Printf("WARNING: HTTP transport running without auth token — bind to localhost only")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/info", handleInfo(reg))
	mux.HandleFunc("/mcp", withAuth(token, handleMCP(ctx, logger, f, reg)))

	srv := &http.Server{
		Addr:              opts.HTTPAddr,
		Handler:           withAccessLog(logger, mux),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       120 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       5 * time.Minute,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	logger.Printf("mcp HTTP server listening addr=%s tools=%d protocol=%s version=%s auth=%s",
		opts.HTTPAddr, len(reg.order), mcpProtocolVersion, build.Version, ifThen(token != "", "bearer", "none"))

	// Shutdown on context cancel.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Printf("mcp HTTP server stopped")
	return nil
}

// resolveToken returns the auth token from opts or the MCP_TOKEN env var.
// opts wins so a flag can override a leaky env.
func resolveToken(opts *ServeOptions) string {
	if opts.Token != "" {
		return opts.Token
	}
	return os.Getenv("MCP_TOKEN")
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func handleInfo(reg *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"name":            "jodoo-cli",
			"version":         build.Version,
			"protocolVersion": mcpProtocolVersion,
			"tools":           len(reg.order),
		})
	}
}

func handleMCP(ctx context.Context, log Logger, f *cmdutil.Factory, reg *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// SSE stream for server-initiated messages — not implemented.
			http.Error(w, "server-sent events not implemented (see docs/MCP/07-roadmap.md)", http.StatusMethodNotAllowed)
			return
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()

		raw, err := io.ReadAll(io.LimitReader(r.Body, 16<<20)) // 16MB cap
		if err != nil {
			writeHTTPRPCError(w, nil, -32700, "read body", err.Error())
			return
		}
		var req rpcRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			writeHTTPRPCError(w, nil, -32700, "parse error", err.Error())
			return
		}
		if req.JSONRPC != "2.0" && req.JSONRPC != "" {
			writeHTTPRPCError(w, req.ID, -32600, "invalid request", "jsonrpc must be \"2.0\"")
			return
		}

		resp := dispatch(r.Context(), log, f, reg, &req)
		w.Header().Set("Content-Type", "application/json")
		if req.ID == nil || resp == nil {
			// Notification — no body, just 204.
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("encode response: %v", err)
		}
		_ = ctx // kept for future use (cancellation propagation)
	}
}

func writeHTTPRPCError(w http.ResponseWriter, id json.RawMessage, code int, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors ride on 200; HTTP status is transport-level only
	_ = json.NewEncoder(w).Encode(&rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg, Data: data},
	})
}

// withAuth returns a handler that requires a Bearer token match when token
// is non-empty. Empty token = open (expected to bind to localhost).
func withAuth(token string, next http.HandlerFunc) http.HandlerFunc {
	if token == "" {
		return next
	}
	expected := "Bearer " + token
	return func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Authorization")
		if !constantTimeEqual(got, expected) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="jodoo-mcp"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// constantTimeEqual avoids short-circuiting so the auth path isn't timeable.
func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func withAccessLog(log Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sr, r)
		// Never log bodies, query strings, or headers — only the path and
		// status. Matches the stdio logger's privacy stance (see 05-security.md).
		log.Printf("http %s %s -> %d dur=%s", r.Method, sanitisePath(r.URL.Path), sr.status, time.Since(start).Truncate(time.Millisecond))
	})
}

func sanitisePath(p string) string {
	// Strip anything after `?` defensively — Go's URL.Path should already
	// exclude query, but the precaution is free.
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	return p
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func ifThen(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

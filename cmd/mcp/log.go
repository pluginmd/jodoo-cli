// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package mcp

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Logger is the tiny surface the server uses. Backed by the standard library
// log package — wrapped so tests can swap it and the stdio path never
// accidentally writes to os.Stdout (which is reserved for JSON-RPC frames).
type Logger interface {
	Printf(format string, args ...interface{})
}

type stdLogger struct{ l *log.Logger }

func (s stdLogger) Printf(format string, args ...interface{}) {
	s.l.Printf(format, args...)
}

// newLogger returns a logger writing to --log-file (if set) or the provided
// fallback (stderr in practice). The close func is a no-op unless we opened
// a file.
func newLogger(path string, fallback io.Writer) (Logger, func(), error) {
	var w io.Writer = fallback
	closeFn := func() {}
	if path != "" {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return nil, nil, err
		}
		w = f
		closeFn = func() { _ = f.Close() }
	}
	l := log.New(w, "[jodoo-mcp] ", log.LstdFlags|log.Lmicroseconds)
	return stdLogger{l: l}, closeFn, nil
}

// logCall formats a one-line summary of a tool invocation. Arguments and
// responses are intentionally NOT logged (they may contain PII or secrets —
// see docs/MCP/05-security.md).
func logCall(l Logger, tool string, exit int, dur time.Duration, isErr bool) {
	status := "ok"
	if isErr {
		status = "error"
	}
	l.Printf(fmt.Sprintf("tool=%s status=%s exit=%d dur=%s", tool, status, exit, dur.Truncate(time.Millisecond)))
}

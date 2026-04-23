// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Exit codes used across the CLI.
const (
	ExitOK         = 0
	ExitGeneric    = 1
	ExitConfig     = 2
	ExitAuth       = 3
	ExitValidation = 4
	ExitAPI        = 5
	ExitNetwork    = 6
)

// ExitError is the structured error type used by the CLI. The root command
// turns it into a JSON envelope on stderr and uses .Code as exit code.
type ExitError struct {
	Code   int
	Detail *ErrDetail
	Raw    bool // skip enrichment when true (raw `api` command)
}

// Error implements error.
func (e *ExitError) Error() string {
	if e.Detail == nil {
		return fmt.Sprintf("exit %d", e.Code)
	}
	if e.Detail.Hint != "" {
		return fmt.Sprintf("%s [%s] (hint: %s)", e.Detail.Message, e.Detail.Type, e.Detail.Hint)
	}
	return fmt.Sprintf("%s [%s]", e.Detail.Message, e.Detail.Type)
}

// ErrDetail is the body of an error envelope.
type ErrDetail struct {
	Type     string      `json:"type"`
	Code     int         `json:"code,omitempty"`
	Message  string      `json:"message"`
	Hint     string      `json:"hint,omitempty"`
	Endpoint string      `json:"endpoint,omitempty"`
	Detail   interface{} `json:"detail,omitempty"`
}

// New constructs a generic ExitError.
func New(exitCode int, errType, message string) *ExitError {
	return &ExitError{Code: exitCode, Detail: &ErrDetail{Type: errType, Message: message}}
}

// ErrWithHint constructs an ExitError with a hint message.
func ErrWithHint(exitCode int, errType, message, hint string) *ExitError {
	return &ExitError{Code: exitCode, Detail: &ErrDetail{Type: errType, Message: message, Hint: hint}}
}

// ErrAuth constructs an authentication error.
func ErrAuth(format string, args ...interface{}) *ExitError {
	return &ExitError{Code: ExitAuth, Detail: &ErrDetail{Type: "auth", Message: fmt.Sprintf(format, args...)}}
}

// ErrConfig constructs a config error.
func ErrConfig(format string, args ...interface{}) *ExitError {
	return &ExitError{Code: ExitConfig, Detail: &ErrDetail{Type: "config", Message: fmt.Sprintf(format, args...)}}
}

// ErrValidation constructs a validation error (4xx-style on the client side).
func ErrValidation(format string, args ...interface{}) *ExitError {
	return &ExitError{Code: ExitValidation, Detail: &ErrDetail{Type: "validation", Message: fmt.Sprintf(format, args...)}}
}

// ErrAPI constructs an API error from the Jodoo response envelope.
func ErrAPI(code int, message string, raw interface{}) *ExitError {
	hint := jodooHint(code)
	return &ExitError{Code: ExitAPI, Detail: &ErrDetail{
		Type: jodooType(code), Code: code, Message: message, Hint: hint, Detail: raw,
	}}
}

// ErrNetwork constructs a transport-level error.
func ErrNetwork(format string, args ...interface{}) *ExitError {
	return &ExitError{Code: ExitNetwork, Detail: &ErrDetail{Type: "network", Message: fmt.Sprintf(format, args...)}}
}

// WriteErrorEnvelope writes the standard error envelope JSON to w.
func WriteErrorEnvelope(w io.Writer, e *ExitError) {
	env := map[string]interface{}{"ok": false}
	if e.Detail != nil {
		env["error"] = e.Detail
	} else {
		env["error"] = map[string]interface{}{"type": "internal", "message": e.Error()}
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(env); err != nil {
		fmt.Fprintln(w, `{"ok":false,"error":{"type":"internal","message":"failed to marshal error"}}`)
	}
}

// jodooType maps a Jodoo error code to a coarse type label.
func jodooType(code int) string {
	switch code {
	case 8301, 17018:
		return "auth"
	case 8302, 1058, 4009, 4025, 5024:
		return "permission"
	case 8303, 8304:
		return "rate_limit"
	case 17017, 17025, 17026, 17032, 17034, 4815, 1096, 3005:
		return "validation"
	case 7103, 7212, 7216, 7217, 7218, 7219, 17023, 17024:
		return "quota"
	}
	if code >= 1000 && code < 2000 {
		return "member"
	}
	if code >= 2000 && code < 4000 {
		return "form"
	}
	if code >= 4000 && code < 5000 {
		return "data"
	}
	if code >= 5000 && code < 7000 {
		return "workflow"
	}
	if code >= 6000 && code < 7000 {
		return "department"
	}
	return "api"
}

// jodooHint returns a short remediation hint for known codes.
func jodooHint(code int) string {
	switch code {
	case 8301, 17018:
		return "check JODOO_API_KEY or run `jodoo-cli config init`"
	case 8302:
		return "the API key is not authorized for the target app — re-scope on the Open Platform"
	case 8303, 8304:
		return "rate limit exceeded — back off and retry"
	case 17025:
		return "transaction_id must be a UUID"
	case 17026:
		return "use a fresh transaction_id (must be unique)"
	case 17023, 17024:
		return "batch size too large — split into smaller chunks (max 100/req)"
	case 4815:
		return "filter object is malformed — see jodoo +data-list --help for the structure"
	}
	return ""
}

// IsRateLimit reports whether the error is a rate-limit error from Jodoo.
func IsRateLimit(e *ExitError) bool {
	if e == nil || e.Detail == nil {
		return false
	}
	c := e.Detail.Code
	return c == 8303 || c == 8304
}

// PrintHumanError writes a single-line human-readable error to w (used by
// commands that don't write the JSON envelope).
func PrintHumanError(w io.Writer, e *ExitError) {
	if e == nil || e.Detail == nil {
		return
	}
	parts := []string{e.Detail.Message}
	if e.Detail.Hint != "" {
		parts = append(parts, "hint: "+e.Detail.Hint)
	}
	fmt.Fprintln(w, "Error:", strings.Join(parts, " — "))
}

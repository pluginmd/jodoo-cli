// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package common

import (
	"github.com/google/uuid"

	"jodoo-cli/internal/output"
)

// CommonFlags returns flag definitions reused across many shortcuts.
//
// We build them as a function rather than as package-level vars because
// the Required field varies per call site.

// AppIDFlag is the standard --app-id flag.
func AppIDFlag(required bool) Flag {
	return Flag{Name: "app-id", Desc: "Jodoo app ID", Required: required}
}

// EntryIDFlag is the standard --entry-id flag.
func EntryIDFlag(required bool) Flag {
	return Flag{Name: "entry-id", Desc: "Jodoo form (entry) ID", Required: required}
}

// DataIDFlag is the standard --data-id flag.
func DataIDFlag(required bool) Flag {
	return Flag{Name: "data-id", Desc: "Jodoo record (data) ID", Required: required}
}

// LimitFlag returns a --limit flag with sane default.
func LimitFlag(def int) Flag {
	return Flag{Name: "limit", Type: "int", Default: itoa(def),
		Desc: "page size (1-100)"}
}

// SkipFlag returns a --skip flag.
func SkipFlag() Flag {
	return Flag{Name: "skip", Type: "int", Default: "0", Desc: "records to skip"}
}

// UsernameFlag is the --username flag (Jodoo user identifier).
func UsernameFlag(required bool) Flag {
	return Flag{Name: "username", Desc: "Jodoo user (member) ID", Required: required}
}

// DataJSONFlag returns the --data flag accepting a JSON object.
func DataJSONFlag(required bool) Flag {
	return Flag{Name: "data", Desc: "record JSON object",
		Required: required, Input: []string{File, Stdin}}
}

// DataListJSONFlag returns the --data-list flag accepting a JSON array.
func DataListJSONFlag(required bool) Flag {
	return Flag{Name: "data-list", Desc: "array of record JSON objects",
		Required: required, Input: []string{File, Stdin}}
}

// FilterJSONFlag returns the --filter flag.
func FilterJSONFlag() Flag {
	return Flag{Name: "filter", Desc: "filter JSON: {rel,cond:[{field,type,method,value}]}",
		Input: []string{File, Stdin}}
}

// FieldsFlag returns --fields (repeatable).
func FieldsFlag() Flag {
	return Flag{Name: "fields", Type: "string_array", Desc: "fields to return (repeatable)"}
}

// TransactionIDFlag is the standard --transaction-id flag (UUID).
//
// When pass-through, callers may want auto-generation: see EnsureTransactionID.
func TransactionIDFlag(required bool, helpHint string) Flag {
	desc := "transaction_id (UUID; binds uploaded files to this request)"
	if helpHint != "" {
		desc = helpHint
	}
	return Flag{Name: "transaction-id", Desc: desc, Required: required}
}

// EnsureTransactionID returns the user-supplied --transaction-id flag,
// generating a fresh UUID v4 if empty. Callers that want strict mode
// should use TransactionIDFlag with required=true instead.
func EnsureTransactionID(r *RuntimeContext) string {
	v := r.Str("transaction-id")
	if v != "" {
		return v
	}
	return uuid.NewString()
}

// itoa is a no-import shim for small ints (Default field is a string).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// RequireString returns the flag value or an ExitError if empty.
func RequireString(r *RuntimeContext, name string) (string, error) {
	v := r.Str(name)
	if v == "" {
		return "", output.ErrValidation("--%s is required", name)
	}
	return v, nil
}

// RequireBatchLimit returns a validation error when a slice exceeds the
// per-call cap (Jodoo enforces 100 per batch).
func RequireBatchLimit(flag string, cap, got int) error {
	return output.ErrValidation("%s holds %d items but the per-request cap is %d — split into smaller batches", flag, got, cap)
}

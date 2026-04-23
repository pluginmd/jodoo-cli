// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package common

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DryRunAPI is the printable preview returned by a Shortcut.DryRun hook.
//
// All Jodoo endpoints are POST, so we omit the method field; users get a
// preview of `POST <url>` plus the body that would be sent.
type DryRunAPI struct {
	URL  string                 `json:"url"`
	Body map[string]interface{} `json:"body,omitempty"`
}

// NewDryRunAPI returns a builder for a DryRun preview.
func NewDryRunAPI(path string) *DryRunAPI {
	if path != "" && !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "http") {
		path = "/" + path
	}
	return &DryRunAPI{URL: path, Body: map[string]interface{}{}}
}

// Set adds or overrides a body field. Convenience for shortcuts that
// progressively build the body.
func (d *DryRunAPI) Set(key string, val interface{}) *DryRunAPI {
	if d.Body == nil {
		d.Body = map[string]interface{}{}
	}
	d.Body[key] = val
	return d
}

// SetIf only adds the field when val is "truthy" (non-empty string,
// non-zero number, non-nil interface).
func (d *DryRunAPI) SetIf(key string, val interface{}) *DryRunAPI {
	switch v := val.(type) {
	case nil:
		return d
	case string:
		if v == "" {
			return d
		}
	case int:
		if v == 0 {
			return d
		}
	case bool:
		if !v {
			return d
		}
	}
	return d.Set(key, val)
}

// BodyJSON sets the entire body to a parsed JSON object (or merges if
// already a map).
func (d *DryRunAPI) BodyJSON(raw string) *DryRunAPI {
	if strings.TrimSpace(raw) == "" {
		return d
	}
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &v); err == nil {
		if d.Body == nil {
			d.Body = v
			return d
		}
		for k, val := range v {
			d.Body[k] = val
		}
	}
	return d
}

// Format renders the DryRunAPI as a human-readable string.
func (d *DryRunAPI) Format() string {
	b, _ := json.MarshalIndent(d.Body, "", "  ")
	return fmt.Sprintf("POST %s\n%s\n", d.URL, string(b))
}

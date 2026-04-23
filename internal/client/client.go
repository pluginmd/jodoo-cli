// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package client wraps the Jodoo HTTP API.
//
// Jodoo's API is small: every endpoint is HTTP POST, every body is JSON
// (except the multipart file upload which targets a per-token URL outside
// the standard base URL). The response envelope is always:
//
//	{ "code": <int>, "msg": <string>, ... data fields ... }
//
// `code == 0` means success; any other value is a structured error.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"jodoo-cli/internal/build"
	"jodoo-cli/internal/core"
	"jodoo-cli/internal/output"
)

// DefaultTimeout is the per-request HTTP timeout.
const DefaultTimeout = 30 * time.Second

// LongTimeout is used for batch / long-running operations.
const LongTimeout = 5 * time.Minute

// APIClient wraps net/http with auth + envelope parsing.
type APIClient struct {
	Config *core.CliConfig
	HTTP   *http.Client
	// UserAgent overrides the default UA string.
	UserAgent string
}

// New returns an APIClient with sensible defaults.
func New(cfg *core.CliConfig) *APIClient {
	return &APIClient{
		Config: cfg,
		HTTP: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// userAgent returns the HTTP User-Agent header value.
func (c *APIClient) userAgent() string {
	if c.UserAgent != "" {
		return c.UserAgent
	}
	return fmt.Sprintf("jodoo-cli/%s", build.Version)
}

// Request describes one Jodoo POST call.
type Request struct {
	// Path is the API path (with or without leading slash) appended to BaseURL.
	// All Jodoo endpoints are POST; method is hard-coded.
	Path string
	// Body is marshaled to JSON. Pass nil for an empty body.
	Body interface{}
	// Headers may add per-request headers (auth header is set automatically).
	Headers map[string]string
	// Timeout overrides the default per-request timeout.
	Timeout time.Duration
}

// RawResponse holds both the parsed envelope ("code", "msg") and the raw
// JSON map returned by the server.
type RawResponse struct {
	Code   int
	Msg    string
	Data   map[string]interface{} // entire decoded body (envelope + payload)
	Status int                    // HTTP status code
	Body   []byte                 // raw response body
}

// DryRun returns the prepared request without sending it.
type DryRun struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

// BuildDryRun constructs a DryRun preview for a Request.
func (c *APIClient) BuildDryRun(req Request) (*DryRun, error) {
	u, err := c.fullURL(req.Path)
	if err != nil {
		return nil, err
	}
	dry := &DryRun{
		Method:  http.MethodPost,
		URL:     u,
		Headers: redactedHeaders(c.Config.APIKey),
	}
	if req.Body != nil {
		// round-trip through JSON to normalize types
		b, err := json.Marshal(req.Body)
		if err != nil {
			return nil, err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err == nil {
			dry.Body = m
		} else {
			// not an object — surface raw payload under "_raw"
			var v interface{}
			_ = json.Unmarshal(b, &v)
			dry.Body = map[string]interface{}{"_raw": v}
		}
	}
	return dry, nil
}

// redactedHeaders returns the headers dry-run prints (token redacted).
func redactedHeaders(apiKey string) map[string]string {
	disp := "Bearer ***"
	if len(apiKey) > 4 {
		disp = "Bearer " + apiKey[:2] + "***" + apiKey[len(apiKey)-2:]
	}
	return map[string]string{
		"Authorization": disp,
		"Content-Type":  "application/json",
	}
}

// Do sends a Jodoo POST request and returns the parsed envelope.
//
// On code != 0 it returns an *output.ExitError describing the API failure.
// On a transport / decode failure it returns an ExitError with type "network".
func (c *APIClient) Do(ctx context.Context, req Request) (*RawResponse, error) {
	if c.Config == nil || c.Config.APIKey == "" {
		return nil, output.ErrAuth("no API key configured (run `jodoo-cli config init` or export JODOO_API_KEY)")
	}
	u, err := c.fullURL(req.Path)
	if err != nil {
		return nil, output.ErrValidation("invalid path %q: %v", req.Path, err)
	}

	var body io.Reader
	if req.Body != nil {
		b, err := json.Marshal(req.Body)
		if err != nil {
			return nil, output.ErrValidation("marshal body: %v", err)
		}
		body = bytes.NewReader(b)
	} else {
		body = bytes.NewReader([]byte(`{}`))
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, body)
	if err != nil {
		return nil, output.ErrNetwork("build request: %v", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.Config.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent())
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	httpClient := c.HTTP
	if req.Timeout > 0 && req.Timeout != httpClient.Timeout {
		clone := *httpClient
		clone.Timeout = req.Timeout
		httpClient = &clone
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, output.ErrNetwork("POST %s: %v", u, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, output.ErrNetwork("read response: %v", err)
	}

	out := &RawResponse{Status: resp.StatusCode, Body: raw}
	if len(raw) == 0 {
		if resp.StatusCode >= 400 {
			return out, output.ErrAPI(resp.StatusCode, fmt.Sprintf("HTTP %d (empty body)", resp.StatusCode), nil)
		}
		out.Data = map[string]interface{}{}
		return out, nil
	}

	if err := json.Unmarshal(raw, &out.Data); err != nil {
		return out, output.ErrAPI(resp.StatusCode,
			fmt.Sprintf("invalid JSON response (HTTP %d): %s", resp.StatusCode, snippet(raw)),
			nil)
	}
	if v, ok := out.Data["code"]; ok {
		if n, ok := toInt(v); ok {
			out.Code = n
		}
	}
	if v, ok := out.Data["msg"].(string); ok {
		out.Msg = v
	}

	if out.Code != 0 {
		// API-side error — surface code + msg via ErrAPI
		return out, output.ErrAPI(out.Code, msgOrFallback(out.Msg, out.Code), out.Data)
	}
	if resp.StatusCode >= 400 {
		return out, output.ErrAPI(resp.StatusCode,
			fmt.Sprintf("HTTP %d (%s)", resp.StatusCode, msgOrFallback(out.Msg, out.Code)),
			out.Data)
	}

	return out, nil
}

// fullURL prepends BaseURL when path is relative.
func (c *APIClient) fullURL(path string) (string, error) {
	if path == "" {
		return "", errors.New("empty path")
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path, nil
	}
	base := c.Config.BaseURL
	if base == "" {
		base = core.DefaultBaseURL
	}
	base = strings.TrimRight(base, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path, nil
}

// PayloadOnly returns the response body with the envelope keys ("code","msg")
// stripped — most consumers only care about the data fields.
func (r *RawResponse) PayloadOnly() map[string]interface{} {
	if r == nil || r.Data == nil {
		return nil
	}
	out := make(map[string]interface{}, len(r.Data))
	for k, v := range r.Data {
		if k == "code" || k == "msg" {
			continue
		}
		out[k] = v
	}
	return out
}

func snippet(raw []byte) string {
	const max = 200
	s := string(raw)
	s = strings.TrimSpace(s)
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}

func msgOrFallback(msg string, code int) string {
	if msg != "" {
		return msg
	}
	return fmt.Sprintf("api error %d", code)
}

func toInt(v interface{}) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int64:
		return int(t), true
	case float64:
		return int(t), true
	case json.Number:
		n, err := t.Int64()
		if err == nil {
			return int(n), true
		}
	case string:
		var n int
		_, err := fmt.Sscanf(t, "%d", &n)
		if err == nil {
			return n, true
		}
	}
	return 0, false
}

// ── File upload helpers ──

// FileUploadRequest describes a Jodoo file upload (Step 2 of the file flow).
type FileUploadRequest struct {
	URL      string // absolute URL returned by /v5/app/entry/file/get_upload_token
	Token    string // token returned alongside URL
	FilePath string // local file to upload (mutually exclusive with FileBytes)
	FileBytes []byte
	FileName string // optional; defaults to basename(FilePath)
	Timeout  time.Duration
}

// UploadFile performs the multipart POST to a per-token upload URL.
// Returns the parsed envelope (key + url, ...).
func (c *APIClient) UploadFile(ctx context.Context, req FileUploadRequest) (*RawResponse, error) {
	if req.URL == "" || req.Token == "" {
		return nil, output.ErrValidation("file upload requires both --url and --token")
	}

	var body []byte
	name := req.FileName
	if req.FilePath != "" {
		b, err := os.ReadFile(req.FilePath)
		if err != nil {
			return nil, output.ErrValidation("read file %s: %v", req.FilePath, err)
		}
		body = b
		if name == "" {
			name = filepath.Base(req.FilePath)
		}
	} else if len(req.FileBytes) > 0 {
		body = req.FileBytes
		if name == "" {
			name = "upload.bin"
		}
	} else {
		return nil, output.ErrValidation("file upload requires --file or in-memory bytes")
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField("token", req.Token); err != nil {
		return nil, output.ErrNetwork("multipart token: %v", err)
	}
	fw, err := mw.CreateFormFile("file", name)
	if err != nil {
		return nil, output.ErrNetwork("multipart file: %v", err)
	}
	if _, err := fw.Write(body); err != nil {
		return nil, output.ErrNetwork("multipart write: %v", err)
	}
	if err := mw.Close(); err != nil {
		return nil, output.ErrNetwork("multipart close: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, req.URL, &buf)
	if err != nil {
		return nil, output.ErrNetwork("build upload request: %v", err)
	}
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())
	httpReq.Header.Set("User-Agent", c.userAgent())

	timeout := req.Timeout
	if timeout == 0 {
		timeout = LongTimeout
	}
	hc := *c.HTTP
	hc.Timeout = timeout

	resp, err := hc.Do(httpReq)
	if err != nil {
		return nil, output.ErrNetwork("POST %s: %v", req.URL, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, output.ErrNetwork("read upload response: %v", err)
	}

	out := &RawResponse{Status: resp.StatusCode, Body: raw}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out.Data); err != nil {
			return out, output.ErrAPI(resp.StatusCode,
				fmt.Sprintf("invalid upload response (HTTP %d): %s", resp.StatusCode, snippet(raw)),
				nil)
		}
		if v, ok := out.Data["code"]; ok {
			if n, ok := toInt(v); ok {
				out.Code = n
			}
		}
		if v, ok := out.Data["msg"].(string); ok {
			out.Msg = v
		}
	}
	if out.Code != 0 {
		return out, output.ErrAPI(out.Code, msgOrFallback(out.Msg, out.Code), out.Data)
	}
	if resp.StatusCode >= 400 {
		return out, output.ErrAPI(resp.StatusCode,
			fmt.Sprintf("upload HTTP %d", resp.StatusCode), out.Data)
	}
	return out, nil
}

// PingURL returns a URL suitable for connectivity checks (the apps list
// endpoint with a tiny payload). The `doctor` command uses this.
func (c *APIClient) PingURL() string {
	u, _ := c.fullURL("/v5/app/list")
	return u
}

// ParseURL exposes net/url.Parse to consumers that just need a sanity check.
func ParseURL(s string) (*url.URL, error) { return url.Parse(s) }

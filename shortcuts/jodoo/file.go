// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package jodoo

import (
	"context"

	"jodoo-cli/internal/client"
	"jodoo-cli/shortcuts/common"
)

func fileShortcuts() []common.Shortcut {
	return []common.Shortcut{
		fileGetToken,
		fileUpload,
	}
}

// File Upload Credentials and URL Get — POST /v5/app/entry/file/get_upload_token
//
// Returns up to 100 (url, token) pairs scoped to a transaction_id. The
// uploaded files become reusable in any record-create / record-update
// call that carries the SAME transaction_id.
var fileGetToken = common.Shortcut{
	Service:     "jodoo",
	Command:     "+file-get-token",
	Description: "Step 1: get upload URLs/tokens (up to 100/req)",
	Risk:        "read",
	HasFormat:   true,
	Tips: []string{
		"Pass the SAME --transaction-id to +file-upload, then to +data-create / +data-update.",
		"If --transaction-id is omitted, a fresh UUID v4 is generated and printed in the response.",
	},
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.EntryIDFlag(true),
		common.TransactionIDFlag(false, "transaction_id (auto-generated if omitted)"),
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/app/entry/file/get_upload_token").
			Set("app_id", r.Str("app-id")).
			Set("entry_id", r.Str("entry-id")).
			Set("transaction_id", common.EnsureTransactionID(r))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		txn := common.EnsureTransactionID(r)
		body := map[string]interface{}{
			"app_id":         r.Str("app-id"),
			"entry_id":       r.Str("entry-id"),
			"transaction_id": txn,
		}
		data, err := r.CallAPI("/v5/app/entry/file/get_upload_token", body)
		if err != nil {
			return err
		}
		// always include the transaction_id so callers can reuse it
		if _, ok := data["transaction_id"]; !ok {
			data["transaction_id"] = txn
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// File Upload — POST <url returned by Step 1>
//
// This call goes to a per-token URL OUTSIDE api.jodoo.com, so we don't
// re-use the standard CallAPI plumbing.
var fileUpload = common.Shortcut{
	Service:     "jodoo",
	Command:     "+file-upload",
	Description: "Step 2: upload one file to the URL/token from +file-get-token",
	Risk:        "write",
	HasFormat:   true,
	Tips: []string{
		"Each (url, token) pair accepts exactly one file — no overwrite.",
		"Use the returned `key` value when filling image / attachment fields in +data-create.",
	},
	Flags: []common.Flag{
		{Name: "url", Required: true, Desc: "absolute upload URL from +file-get-token"},
		{Name: "token", Required: true, Desc: "upload token paired with --url"},
		{Name: "file", Required: true, Desc: "local path of the file to upload"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI(r.Str("url")).
			Set("token", r.Str("token")).
			Set("file", r.Str("file"))
	},
	Execute: func(ctx context.Context, r *common.RuntimeContext) error {
		c := client.New(r.Config)
		resp, err := c.UploadFile(ctx, client.FileUploadRequest{
			URL:      r.Str("url"),
			Token:    r.Str("token"),
			FilePath: r.Str("file"),
		})
		if err != nil {
			return err
		}
		// Strip the envelope; expose only the payload (key, url, ...).
		payload := map[string]interface{}{}
		for k, v := range resp.Data {
			if k == "code" || k == "msg" {
				continue
			}
			payload[k] = v
		}
		r.OutFormat(payload, nil, nil)
		return nil
	},
}

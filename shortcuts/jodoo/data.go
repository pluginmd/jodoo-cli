// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package jodoo

import (
	"context"

	"jodoo-cli/shortcuts/common"
)

func dataShortcuts() []common.Shortcut {
	return []common.Shortcut{
		widgetList,
		dataGet,
		dataList,
		dataCreate,
		dataBatchCreate,
		dataUpdate,
		dataBatchUpdate,
		dataDelete,
		dataBatchDelete,
	}
}

// Form Fields Query — POST /v5/app/entry/widget/list
var widgetList = common.Shortcut{
	Service:     "jodoo",
	Command:     "+widget-list",
	Description: "List fields (widgets) of a form",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.EntryIDFlag(true),
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/app/entry/widget/list").
			Set("app_id", r.Str("app-id")).
			Set("entry_id", r.Str("entry-id"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"app_id":   r.Str("app-id"),
			"entry_id": r.Str("entry-id"),
		}
		data, err := r.CallAPI("/v5/app/entry/widget/list", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Single Record Query — POST /v5/app/entry/data/get
var dataGet = common.Shortcut{
	Service:     "jodoo",
	Command:     "+data-get",
	Description: "Read a single record by data ID",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.EntryIDFlag(true),
		common.DataIDFlag(true),
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/app/entry/data/get").
			Set("app_id", r.Str("app-id")).
			Set("entry_id", r.Str("entry-id")).
			Set("data_id", r.Str("data-id"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"app_id":   r.Str("app-id"),
			"entry_id": r.Str("entry-id"),
			"data_id":  r.Str("data-id"),
		}
		data, err := r.CallAPI("/v5/app/entry/data/get", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Multiple Records Query — POST /v5/app/entry/data/list
var dataList = common.Shortcut{
	Service:     "jodoo",
	Command:     "+data-list",
	Description: "List records (cursor pagination via --data-id)",
	Risk:        "read",
	HasFormat:   true,
	Tips: []string{
		"records are returned ordered by data_id ASC",
		"loop with the last `_id` as the next --data-id until len(returned) < --limit",
		"--paginate-all walks every page automatically (be mindful of the 5 req/s rate limit)",
	},
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.EntryIDFlag(true),
		{Name: "data-id", Desc: "cursor (last _id from previous page)"},
		common.FieldsFlag(),
		common.FilterJSONFlag(),
		common.LimitFlag(10),
		{Name: "paginate-all", Type: "bool", Default: "false", Desc: "follow cursor until exhausted"},
	},
	Validate: func(_ context.Context, r *common.RuntimeContext) error {
		if raw := r.Str("filter"); raw != "" {
			if _, err := common.ParseJSONObject(raw, "filter"); err != nil {
				return err
			}
		}
		return nil
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/app/entry/data/list").
			Set("app_id", r.Str("app-id")).
			Set("entry_id", r.Str("entry-id")).
			SetIf("data_id", r.Str("data-id")).
			SetIf("limit", common.ParseIntBounded(r, "limit", 1, 100)).
			SetIf("fields", r.StrArray("fields")).
			BodyJSON(`{"filter":` + emptyOr(r.Str("filter")) + `}`)
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		limit := common.ParseIntBounded(r, "limit", 1, 100)
		buildBody := func(cursor string) map[string]interface{} {
			body := map[string]interface{}{
				"app_id":   r.Str("app-id"),
				"entry_id": r.Str("entry-id"),
				"limit":    limit,
			}
			if cursor != "" {
				body["data_id"] = cursor
			}
			if fields := r.StrArray("fields"); len(fields) > 0 {
				body["fields"] = fields
			}
			if raw := r.Str("filter"); raw != "" {
				if m, err := common.ParseJSONObject(raw, "filter"); err == nil {
					body["filter"] = m
				}
			}
			return body
		}

		if r.Bool("paginate-all") {
			items, err := r.PaginateAll("/v5/app/entry/data/list", "data", limit, buildBody)
			if err != nil {
				return err
			}
			r.OutFormat(map[string]interface{}{"data": items, "count": len(items)}, nil, nil)
			return nil
		}
		body := buildBody(r.Str("data-id"))
		data, err := r.CallAPI("/v5/app/entry/data/list", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// emptyOr is a tiny helper for DryRun JSON building.
func emptyOr(s string) string {
	if s == "" {
		return "null"
	}
	return s
}

// Single Record Creation — POST /v5/app/entry/data/create
var dataCreate = common.Shortcut{
	Service:     "jodoo",
	Command:     "+data-create",
	Description: "Create a single record",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.EntryIDFlag(true),
		common.DataJSONFlag(true),
		{Name: "data-creator", Desc: "submitter username (defaults to business owner)"},
		{Name: "is-start-workflow", Type: "bool", Default: "false", Desc: "trigger workflow"},
		{Name: "is-start-trigger", Type: "bool", Default: "false", Desc: "trigger Automations"},
		common.TransactionIDFlag(false, "transaction_id (UUID; required for file uploads)"),
	},
	Validate: func(_ context.Context, r *common.RuntimeContext) error {
		_, err := common.ParseJSONObject(r.Str("data"), "data")
		return err
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		body, _ := common.ParseJSONObject(r.Str("data"), "data")
		dry := common.NewDryRunAPI("/v5/app/entry/data/create").
			Set("app_id", r.Str("app-id")).
			Set("entry_id", r.Str("entry-id")).
			Set("data", body).
			SetIf("data_creator", r.Str("data-creator")).
			Set("is_start_workflow", r.Bool("is-start-workflow")).
			Set("is_start_trigger", r.Bool("is-start-trigger"))
		if t := r.Str("transaction-id"); t != "" {
			dry.Set("transaction_id", t)
		}
		return dry
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body, err := common.ParseJSONObject(r.Str("data"), "data")
		if err != nil {
			return err
		}
		req := map[string]interface{}{
			"app_id":            r.Str("app-id"),
			"entry_id":          r.Str("entry-id"),
			"data":              body,
			"is_start_workflow": r.Bool("is-start-workflow"),
			"is_start_trigger":  r.Bool("is-start-trigger"),
		}
		if v := r.Str("data-creator"); v != "" {
			req["data_creator"] = v
		}
		if v := r.Str("transaction-id"); v != "" {
			req["transaction_id"] = v
		}
		data, err := r.CallAPI("/v5/app/entry/data/create", req)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Multiple Records Creation — POST /v5/app/entry/data/batch_create
var dataBatchCreate = common.Shortcut{
	Service:     "jodoo",
	Command:     "+data-batch-create",
	Description: "Batch create records (max 100/req)",
	Risk:        "write",
	HasFormat:   true,
	Tips: []string{
		"On partial failure, retry with the SAME --transaction-id; previously inserted rows are skipped.",
		"Pass --transaction-id explicitly when retrying; otherwise a fresh UUID is generated each call.",
	},
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.EntryIDFlag(true),
		common.DataListJSONFlag(true),
		{Name: "data-creator", Desc: "submitter username"},
		{Name: "is-start-workflow", Type: "bool", Default: "false", Desc: "trigger workflow per row"},
		common.TransactionIDFlag(false, "transaction_id (auto-generated if omitted)"),
	},
	Validate: func(_ context.Context, r *common.RuntimeContext) error {
		arr, err := common.ParseJSONArray(r.Str("data-list"), "data-list")
		if err != nil {
			return err
		}
		if len(arr) > 100 {
			return common.RequireBatchLimit("--data-list", 100, len(arr))
		}
		return nil
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		arr, _ := common.ParseJSONArray(r.Str("data-list"), "data-list")
		return common.NewDryRunAPI("/v5/app/entry/data/batch_create").
			Set("app_id", r.Str("app-id")).
			Set("entry_id", r.Str("entry-id")).
			Set("data_list", arr).
			SetIf("data_creator", r.Str("data-creator")).
			Set("is_start_workflow", r.Bool("is-start-workflow")).
			Set("transaction_id", common.EnsureTransactionID(r))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		arr, err := common.ParseJSONArray(r.Str("data-list"), "data-list")
		if err != nil {
			return err
		}
		req := map[string]interface{}{
			"app_id":            r.Str("app-id"),
			"entry_id":          r.Str("entry-id"),
			"data_list":         arr,
			"is_start_workflow": r.Bool("is-start-workflow"),
			"transaction_id":    common.EnsureTransactionID(r),
		}
		if v := r.Str("data-creator"); v != "" {
			req["data_creator"] = v
		}
		data, err := r.CallAPI("/v5/app/entry/data/batch_create", req)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Single Record Update — POST /v5/app/entry/data/update
var dataUpdate = common.Shortcut{
	Service:     "jodoo",
	Command:     "+data-update",
	Description: "Update a single record (subforms must be passed in full)",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.EntryIDFlag(true),
		common.DataIDFlag(true),
		common.DataJSONFlag(true),
		{Name: "is-start-trigger", Type: "bool", Default: "false", Desc: "trigger Automations"},
		common.TransactionIDFlag(false, "transaction_id (UUID; required for file uploads)"),
	},
	Validate: func(_ context.Context, r *common.RuntimeContext) error {
		_, err := common.ParseJSONObject(r.Str("data"), "data")
		return err
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		body, _ := common.ParseJSONObject(r.Str("data"), "data")
		dry := common.NewDryRunAPI("/v5/app/entry/data/update").
			Set("app_id", r.Str("app-id")).
			Set("entry_id", r.Str("entry-id")).
			Set("data_id", r.Str("data-id")).
			Set("data", body).
			Set("is_start_trigger", r.Bool("is-start-trigger"))
		if t := r.Str("transaction-id"); t != "" {
			dry.Set("transaction_id", t)
		}
		return dry
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body, err := common.ParseJSONObject(r.Str("data"), "data")
		if err != nil {
			return err
		}
		req := map[string]interface{}{
			"app_id":           r.Str("app-id"),
			"entry_id":         r.Str("entry-id"),
			"data_id":          r.Str("data-id"),
			"data":             body,
			"is_start_trigger": r.Bool("is-start-trigger"),
		}
		if v := r.Str("transaction-id"); v != "" {
			req["transaction_id"] = v
		}
		data, err := r.CallAPI("/v5/app/entry/data/update", req)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Multiple Records Update — POST /v5/app/entry/data/batch_update
var dataBatchUpdate = common.Shortcut{
	Service:     "jodoo",
	Command:     "+data-batch-update",
	Description: "Apply the same patch to many records (no subform fields)",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.EntryIDFlag(true),
		{Name: "data-ids", Type: "string_array", Required: true, Desc: "data IDs to update (repeatable)"},
		common.DataJSONFlag(true),
		common.TransactionIDFlag(false, "transaction_id (UUID; required for file uploads)"),
	},
	Validate: func(_ context.Context, r *common.RuntimeContext) error {
		_, err := common.ParseJSONObject(r.Str("data"), "data")
		return err
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		body, _ := common.ParseJSONObject(r.Str("data"), "data")
		return common.NewDryRunAPI("/v5/app/entry/data/batch_update").
			Set("app_id", r.Str("app-id")).
			Set("entry_id", r.Str("entry-id")).
			Set("data_ids", r.StrArray("data-ids")).
			Set("data", body).
			SetIf("transaction_id", r.Str("transaction-id"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body, err := common.ParseJSONObject(r.Str("data"), "data")
		if err != nil {
			return err
		}
		req := map[string]interface{}{
			"app_id":   r.Str("app-id"),
			"entry_id": r.Str("entry-id"),
			"data_ids": r.StrArray("data-ids"),
			"data":     body,
		}
		if v := r.Str("transaction-id"); v != "" {
			req["transaction_id"] = v
		}
		data, err := r.CallAPI("/v5/app/entry/data/batch_update", req)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Single Record Deletion — POST /v5/app/entry/data/delete
var dataDelete = common.Shortcut{
	Service:     "jodoo",
	Command:     "+data-delete",
	Description: "Delete a single record by data ID",
	Risk:        "high-risk-write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.EntryIDFlag(true),
		common.DataIDFlag(true),
		{Name: "is-start-trigger", Type: "bool", Default: "false", Desc: "trigger Automations"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/app/entry/data/delete").
			Set("app_id", r.Str("app-id")).
			Set("entry_id", r.Str("entry-id")).
			Set("data_id", r.Str("data-id")).
			Set("is_start_trigger", r.Bool("is-start-trigger"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"app_id":           r.Str("app-id"),
			"entry_id":         r.Str("entry-id"),
			"data_id":          r.Str("data-id"),
			"is_start_trigger": r.Bool("is-start-trigger"),
		}
		data, err := r.CallAPI("/v5/app/entry/data/delete", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Multiple Records Deletion — POST /v5/app/entry/data/batch_delete
var dataBatchDelete = common.Shortcut{
	Service:     "jodoo",
	Command:     "+data-batch-delete",
	Description: "Delete many records by data IDs",
	Risk:        "high-risk-write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.EntryIDFlag(true),
		{Name: "data-ids", Type: "string_array", Required: true, Desc: "data IDs to delete (repeatable)"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/app/entry/data/batch_delete").
			Set("app_id", r.Str("app-id")).
			Set("entry_id", r.Str("entry-id")).
			Set("data_ids", r.StrArray("data-ids"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"app_id":   r.Str("app-id"),
			"entry_id": r.Str("entry-id"),
			"data_ids": r.StrArray("data-ids"),
		}
		data, err := r.CallAPI("/v5/app/entry/data/batch_delete", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

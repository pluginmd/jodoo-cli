// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package jodoo

import (
	"context"

	"jodoo-cli/shortcuts/common"
)

func appShortcuts() []common.Shortcut {
	return []common.Shortcut{appList, formList}
}

// User App Query — POST /v5/app/list
var appList = common.Shortcut{
	Service:     "jodoo",
	Command:     "+app-list",
	Description: "List apps reachable by the API key",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		common.LimitFlag(100),
		common.SkipFlag(),
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/app/list").
			Set("limit", common.ParseIntBounded(r, "limit", 1, 100)).
			Set("skip", r.Int("skip"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"limit": common.ParseIntBounded(r, "limit", 1, 100),
			"skip":  r.Int("skip"),
		}
		data, err := r.CallAPI("/v5/app/list", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// User Form Query — POST /v5/app/entry/list
var formList = common.Shortcut{
	Service:     "jodoo",
	Command:     "+form-list",
	Description: "List forms inside an app",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.LimitFlag(100),
		common.SkipFlag(),
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/app/entry/list").
			Set("app_id", r.Str("app-id")).
			Set("limit", common.ParseIntBounded(r, "limit", 1, 100)).
			Set("skip", r.Int("skip"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"app_id": r.Str("app-id"),
			"limit":  common.ParseIntBounded(r, "limit", 1, 100),
			"skip":   r.Int("skip"),
		}
		data, err := r.CallAPI("/v5/app/entry/list", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

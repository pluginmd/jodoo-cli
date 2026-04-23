// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package jodoo

import (
	"context"
	"fmt"

	"jodoo-cli/shortcuts/common"
)

func workflowShortcuts() []common.Shortcut {
	return []common.Shortcut{
		workflowInstanceGet,
		workflowInstanceLogs,
		workflowInstanceActivate,
		workflowInstanceClose,
		workflowTaskList,
		workflowTaskForward,
		workflowTaskReject,
		workflowTaskBack,
		workflowTaskTransfer,
		workflowTaskRevoke,
		workflowTaskAddSign,
		workflowCCList,
		workflowApprovalComments,
	}
}

// instanceIDFlag is the recurring --instance-id flag.
func instanceIDFlag() common.Flag {
	return common.Flag{Name: "instance-id", Required: true, Desc: "workflow instance ID (= data_id)"}
}

func taskIDFlag(required bool) common.Flag {
	return common.Flag{Name: "task-id", Required: required, Desc: "workflow task ID"}
}

// Workflow Instance Get — POST /v5/workflow/instance/get
var workflowInstanceGet = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-instance-get",
	Description: "Get a workflow instance (and optionally its tasks)",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		instanceIDFlag(),
		{Name: "tasks-type", Type: "int", Default: "0", Desc: "0 = no tasks, 1 = include tasks"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/workflow/instance/get").
			Set("instance_id", r.Str("instance-id")).
			Set("tasks_type", r.Int("tasks-type"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"instance_id": r.Str("instance-id"),
			"tasks_type":  r.Int("tasks-type"),
		}
		data, err := r.CallAPI("/v5/workflow/instance/get", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Workflow Instance Logs — POST /v5/workflow/instance/logs
var workflowInstanceLogs = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-instance-logs",
	Description: "List workflow logs (currently only \"comment\" type)",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		instanceIDFlag(),
		{Name: "types", Type: "string_array", Desc: "log types (defaults to [\"comment\"])"},
		common.LimitFlag(100),
		common.SkipFlag(),
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		types := r.StrArray("types")
		if len(types) == 0 {
			types = []string{"comment"}
		}
		return common.NewDryRunAPI("/v5/workflow/instance/logs").
			Set("instance_id", r.Str("instance-id")).
			Set("types", types).
			Set("limit", common.ParseIntBounded(r, "limit", 1, 100)).
			Set("skip", r.Int("skip"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		types := r.StrArray("types")
		if len(types) == 0 {
			types = []string{"comment"}
		}
		body := map[string]interface{}{
			"instance_id": r.Str("instance-id"),
			"types":       types,
			"limit":       common.ParseIntBounded(r, "limit", 1, 100),
			"skip":        r.Int("skip"),
		}
		data, err := r.CallAPI("/v5/workflow/instance/logs", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Workflow Instance Activate — POST /v5/workflow/instance/activate
var workflowInstanceActivate = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-instance-activate",
	Description: "Reactivate a completed workflow instance at a given node",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		instanceIDFlag(),
		{Name: "flow-id", Type: "int", Required: true, Desc: "node ID to reactivate"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/workflow/instance/activate").
			Set("instance_id", r.Str("instance-id")).
			Set("flow_id", r.Int("flow-id"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		data, err := r.CallAPIRaw("/v5/workflow/instance/activate", map[string]interface{}{
			"instance_id": r.Str("instance-id"),
			"flow_id":     r.Int("flow-id"),
		})
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Workflow Instance Close — POST /v5/workflow/instance/close (admin only)
var workflowInstanceClose = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-instance-close",
	Description: "End a workflow instance (admin only)",
	Risk:        "high-risk-write",
	HasFormat:   true,
	Flags:       []common.Flag{instanceIDFlag()},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/workflow/instance/close").
			Set("instance_id", r.Str("instance-id"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		data, err := r.CallAPIRaw("/v5/workflow/instance/close", map[string]interface{}{
			"instance_id": r.Str("instance-id"),
		})
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Workflow Task List — POST /v5/workflow/task/list
var workflowTaskList = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-task-list",
	Description: "List a user's pending workflow tasks",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		common.UsernameFlag(true),
		common.LimitFlag(10),
		common.SkipFlag(),
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/workflow/task/list").
			Set("username", r.Str("username")).
			Set("limit", common.ParseIntBounded(r, "limit", 1, 100)).
			Set("skip", r.Int("skip"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"username": r.Str("username"),
			"limit":    common.ParseIntBounded(r, "limit", 1, 100),
			"skip":     r.Int("skip"),
		}
		data, err := r.CallAPI("/v5/workflow/task/list", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Workflow Task Forward (approve) — POST /v5/workflow/task/forward
var workflowTaskForward = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-task-forward",
	Description: "Approve / submit a workflow task",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.UsernameFlag(true),
		instanceIDFlag(),
		taskIDFlag(true),
		{Name: "comment", Desc: "approval comment (optional)"},
	},
	DryRun:  workflowActionDry("/v5/workflow/task/forward"),
	Execute: workflowActionExec("/v5/workflow/task/forward"),
}

// Workflow Task Reject — POST /v5/workflow/task/reject
var workflowTaskReject = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-task-reject",
	Description: "Reject a workflow task",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.UsernameFlag(true),
		instanceIDFlag(),
		taskIDFlag(true),
		{Name: "comment", Desc: "approval comment (optional)"},
	},
	DryRun:  workflowActionDry("/v5/workflow/task/reject"),
	Execute: workflowActionExec("/v5/workflow/task/reject"),
}

// Workflow Task Back (return) — POST /v5/workflow/task/back
var workflowTaskBack = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-task-back",
	Description: "Return a workflow task to a previous node",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.UsernameFlag(true),
		instanceIDFlag(),
		taskIDFlag(true),
		{Name: "flow-id", Type: "int", Default: "0", Desc: "target node ID (defaults to previous)"},
		{Name: "comment", Desc: "approval comment (optional)"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		dry := workflowActionDry("/v5/workflow/task/back")(nil, r)
		if id := r.Int("flow-id"); id > 0 {
			dry.Set("flow_id", id)
		}
		return dry
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := workflowActionBody(r)
		if id := r.Int("flow-id"); id > 0 {
			body["flow_id"] = id
		}
		data, err := r.CallAPIRaw("/v5/workflow/task/back", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Workflow Task Transfer — POST /v5/workflow/task/transfer
var workflowTaskTransfer = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-task-transfer",
	Description: "Hand over a task to another user",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.UsernameFlag(true),
		instanceIDFlag(),
		taskIDFlag(true),
		{Name: "transfer-username", Required: true, Desc: "user to transfer the task to"},
		{Name: "comment", Desc: "approval comment (optional)"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		dry := workflowActionDry("/v5/workflow/task/transfer")(nil, r)
		dry.Set("transfer_username", r.Str("transfer-username"))
		return dry
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := workflowActionBody(r)
		body["transfer_username"] = r.Str("transfer-username")
		data, err := r.CallAPIRaw("/v5/workflow/task/transfer", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Workflow Task Revoke (withdraw) — POST /v5/workflow/task/revoke
var workflowTaskRevoke = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-task-revoke",
	Description: "Withdraw a previously submitted task",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.UsernameFlag(true),
		instanceIDFlag(),
		taskIDFlag(false),
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/workflow/task/revoke").
			Set("instance_id", r.Str("instance-id")).
			Set("username", r.Str("username")).
			SetIf("task_id", r.Str("task-id"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"instance_id": r.Str("instance-id"),
			"username":    r.Str("username"),
		}
		if v := r.Str("task-id"); v != "" {
			body["task_id"] = v
		}
		data, err := r.CallAPIRaw("/v5/workflow/task/revoke", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Node Approvers Add — POST /v5/workflow/task/add_sign
var workflowTaskAddSign = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-task-add-sign",
	Description: "Add an approver to a workflow node",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.UsernameFlag(true),
		instanceIDFlag(),
		taskIDFlag(true),
		{Name: "add-sign-type", Type: "int", Required: true,
			Desc: "0 = pre, 1 = post, 2 = parallel"},
		{Name: "add-sign-username", Required: true, Desc: "approver to add"},
		{Name: "comment", Desc: "approval comment (optional)"},
	},
	Validate: func(_ context.Context, r *common.RuntimeContext) error {
		t := r.Int("add-sign-type")
		if t < 0 || t > 2 {
			return fmt.Errorf("--add-sign-type must be 0, 1 or 2")
		}
		return nil
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		dry := common.NewDryRunAPI("/v5/workflow/task/add_sign").
			Set("instance_id", r.Str("instance-id")).
			Set("task_id", r.Str("task-id")).
			Set("username", r.Str("username")).
			Set("add_sign_type", r.Int("add-sign-type")).
			Set("add_sign_username", r.Str("add-sign-username"))
		if v := r.Str("comment"); v != "" {
			dry.Set("comment", v)
		}
		return dry
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"instance_id":       r.Str("instance-id"),
			"task_id":           r.Str("task-id"),
			"username":          r.Str("username"),
			"add_sign_type":     r.Int("add-sign-type"),
			"add_sign_username": r.Str("add-sign-username"),
		}
		if v := r.Str("comment"); v != "" {
			body["comment"] = v
		}
		data, err := r.CallAPIRaw("/v5/workflow/task/add_sign", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// CC List — POST /v5/workflow/cc/list
var workflowCCList = common.Shortcut{
	Service:     "jodoo",
	Command:     "+workflow-cc-list",
	Description: "List CC notifications for a user (within 90 days)",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		common.UsernameFlag(true),
		common.LimitFlag(10),
		common.SkipFlag(),
		{Name: "read-status", Default: "all", Enum: []string{"read", "unread", "all"},
			Desc: "filter by read status"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/workflow/cc/list").
			Set("username", r.Str("username")).
			Set("limit", common.ParseIntBounded(r, "limit", 1, 100)).
			Set("skip", r.Int("skip")).
			Set("read_status", r.Str("read-status"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"username":    r.Str("username"),
			"limit":       common.ParseIntBounded(r, "limit", 1, 100),
			"skip":        r.Int("skip"),
			"read_status": r.Str("read-status"),
		}
		data, err := r.CallAPI("/v5/workflow/cc/list", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Approval Comments — POST /v5/app/{appId}/entry/{entryId}/data/{dataId}/approval_comments
//
// Note: this endpoint embeds IDs in the URL — distinct from the
// "/v5/workflow/instance/logs" route.
var workflowApprovalComments = common.Shortcut{
	Service:     "jodoo",
	Command:     "+approval-comments",
	Description: "List approval comments for a workflow form record",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		common.AppIDFlag(true),
		common.EntryIDFlag(true),
		common.DataIDFlag(true),
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI(approvalCommentsPath(r))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		data, err := r.CallAPI(approvalCommentsPath(r), map[string]interface{}{})
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

func approvalCommentsPath(r *common.RuntimeContext) string {
	return fmt.Sprintf("/v5/app/%s/entry/%s/data/%s/approval_comments",
		r.Str("app-id"), r.Str("entry-id"), r.Str("data-id"))
}

// ── Helpers shared across workflow action shortcuts ──

// workflowActionBody builds the body shared by forward/reject/back/transfer.
func workflowActionBody(r *common.RuntimeContext) map[string]interface{} {
	b := map[string]interface{}{
		"username":    r.Str("username"),
		"instance_id": r.Str("instance-id"),
		"task_id":     r.Str("task-id"),
	}
	if v := r.Str("comment"); v != "" {
		b["comment"] = v
	}
	return b
}

func workflowActionDry(path string) func(context.Context, *common.RuntimeContext) *common.DryRunAPI {
	return func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		dry := common.NewDryRunAPI(path).
			Set("username", r.Str("username")).
			Set("instance_id", r.Str("instance-id")).
			Set("task_id", r.Str("task-id"))
		if v := r.Str("comment"); v != "" {
			dry.Set("comment", v)
		}
		return dry
	}
}

func workflowActionExec(path string) func(context.Context, *common.RuntimeContext) error {
	return func(_ context.Context, r *common.RuntimeContext) error {
		data, err := r.CallAPIRaw(path, workflowActionBody(r))
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	}
}

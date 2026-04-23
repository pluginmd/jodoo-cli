// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package jodoo

import (
	"context"
	"fmt"

	"jodoo-cli/shortcuts/common"
)

func contactShortcuts() []common.Shortcut {
	return []common.Shortcut{
		memberList,
		memberGet,
		memberCreate,
		memberUpdate,
		memberDelete,
		memberBatchImport,
		departmentList,
		departmentCreate,
		departmentUpdate,
		departmentDelete,
		departmentBatchImport,
		roleList,
		roleCreate,
		roleMemberList,
	}
}

// ── Member APIs ──

// Member List (recursive) — POST /v5/corp/member/list
var memberList = common.Shortcut{
	Service:     "jodoo",
	Command:     "+member-list",
	Description: "List members under a department (optionally recursive)",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "dept-no", Type: "int", Required: true, Desc: "department number (root = 1)"},
		{Name: "has-child", Type: "bool", Default: "false", Desc: "include sub-departments"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/corp/member/list").
			Set("dept_no", r.Int("dept-no")).
			Set("has_child", r.Bool("has-child"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"dept_no":   r.Int("dept-no"),
			"has_child": r.Bool("has-child"),
		}
		data, err := r.CallAPI("/v5/corp/member/list", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Member Get — POST /v5/corp/member/get
var memberGet = common.Shortcut{
	Service:     "jodoo",
	Command:     "+member-get",
	Description: "Get a member by username",
	Risk:        "read",
	HasFormat:   true,
	Flags:       []common.Flag{common.UsernameFlag(true)},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/corp/member/get").Set("username", r.Str("username"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		data, err := r.CallAPI("/v5/corp/member/get", map[string]interface{}{
			"username": r.Str("username"),
		})
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Member Create — POST /v5/corp/member/create
var memberCreate = common.Shortcut{
	Service:     "jodoo",
	Command:     "+member-create",
	Description: "Create a member (auto-activated)",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "name", Required: true, Desc: "nickname"},
		{Name: "username", Required: true, Desc: "username (letters, digits, underscores)"},
		{Name: "departments", Type: "string_array", Desc: "department numbers (repeatable; ints)"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/corp/member/create").
			Set("name", r.Str("name")).
			Set("username", r.Str("username")).
			SetIf("departments", deptsAsInts(r.StrArray("departments")))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"name":     r.Str("name"),
			"username": r.Str("username"),
		}
		if d := deptsAsInts(r.StrArray("departments")); len(d) > 0 {
			body["departments"] = d
		}
		data, err := r.CallAPI("/v5/corp/member/create", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Member Update — POST /v5/corp/member/update
var memberUpdate = common.Shortcut{
	Service:     "jodoo",
	Command:     "+member-update",
	Description: "Update a member's nickname / departments",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		common.UsernameFlag(true),
		{Name: "name", Desc: "new nickname"},
		{Name: "departments", Type: "string_array", Desc: "new department numbers (repeatable)"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		dry := common.NewDryRunAPI("/v5/corp/member/update").
			Set("username", r.Str("username")).
			SetIf("name", r.Str("name"))
		if d := deptsAsInts(r.StrArray("departments")); len(d) > 0 {
			dry.Set("departments", d)
		}
		return dry
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{"username": r.Str("username")}
		if v := r.Str("name"); v != "" {
			body["name"] = v
		}
		if d := deptsAsInts(r.StrArray("departments")); len(d) > 0 {
			body["departments"] = d
		}
		data, err := r.CallAPI("/v5/corp/member/update", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Member Delete (deactivate) — POST /v5/corp/member/delete
var memberDelete = common.Shortcut{
	Service:     "jodoo",
	Command:     "+member-delete",
	Description: "Deactivate a member (not permanently removed)",
	Risk:        "high-risk-write",
	HasFormat:   true,
	Flags:       []common.Flag{common.UsernameFlag(true)},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/corp/member/delete").Set("username", r.Str("username"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		data, err := r.CallAPI("/v5/corp/member/delete", map[string]interface{}{
			"username": r.Str("username"),
		})
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Member Batch Import — POST /v5/corp/member/batch_import
var memberBatchImport = common.Shortcut{
	Service:     "jodoo",
	Command:     "+member-batch-import",
	Description: "Upsert many members keyed by username",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "users", Required: true,
			Desc:  "JSON array of {username,name,departments?}",
			Input: []string{common.File, common.Stdin}},
	},
	Validate: func(_ context.Context, r *common.RuntimeContext) error {
		_, err := common.ParseJSONArray(r.Str("users"), "users")
		return err
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		arr, _ := common.ParseJSONArray(r.Str("users"), "users")
		return common.NewDryRunAPI("/v5/corp/member/batch_import").Set("users", arr)
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		arr, err := common.ParseJSONArray(r.Str("users"), "users")
		if err != nil {
			return err
		}
		data, err := r.CallAPI("/v5/corp/member/batch_import", map[string]interface{}{
			"users": arr,
		})
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// ── Department APIs ──

// Department List (recursive) — POST /v5/corp/department/list
var departmentList = common.Shortcut{
	Service:     "jodoo",
	Command:     "+department-list",
	Description: "List sub-departments (optionally recursive)",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "dept-no", Type: "int", Required: true, Desc: "department number (root = 1)"},
		{Name: "has-child", Type: "bool", Default: "false", Desc: "include all descendants"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/corp/department/list").
			Set("dept_no", r.Int("dept-no")).
			Set("has_child", r.Bool("has-child"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		data, err := r.CallAPI("/v5/corp/department/list", map[string]interface{}{
			"dept_no":   r.Int("dept-no"),
			"has_child": r.Bool("has-child"),
		})
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Department Create — POST /v5/corp/department/create
var departmentCreate = common.Shortcut{
	Service:     "jodoo",
	Command:     "+department-create",
	Description: "Create a department (max name 32 chars)",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "name", Required: true, Desc: "department name"},
		{Name: "parent-no", Type: "int", Default: "0", Desc: "parent department (defaults to root)"},
		{Name: "dept-no", Type: "int", Default: "0", Desc: "custom dept ID (auto if 0)"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		dry := common.NewDryRunAPI("/v5/corp/department/create").Set("name", r.Str("name"))
		if v := r.Int("parent-no"); v > 0 {
			dry.Set("parent_no", v)
		}
		if v := r.Int("dept-no"); v > 0 {
			dry.Set("dept_no", v)
		}
		return dry
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{"name": r.Str("name")}
		if v := r.Int("parent-no"); v > 0 {
			body["parent_no"] = v
		}
		if v := r.Int("dept-no"); v > 0 {
			body["dept_no"] = v
		}
		data, err := r.CallAPI("/v5/corp/department/create", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Department Update — POST /v5/corp/department/update
var departmentUpdate = common.Shortcut{
	Service:     "jodoo",
	Command:     "+department-update",
	Description: "Update a department's name / parent",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "dept-no", Type: "int", Required: true, Desc: "department ID"},
		{Name: "name", Desc: "new name"},
		{Name: "parent-no", Type: "int", Default: "0", Desc: "new parent department ID"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		dry := common.NewDryRunAPI("/v5/corp/department/update").Set("dept_no", r.Int("dept-no"))
		if v := r.Str("name"); v != "" {
			dry.Set("name", v)
		}
		if v := r.Int("parent-no"); v > 0 {
			dry.Set("parent_no", v)
		}
		return dry
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{"dept_no": r.Int("dept-no")}
		if v := r.Str("name"); v != "" {
			body["name"] = v
		}
		if v := r.Int("parent-no"); v > 0 {
			body["parent_no"] = v
		}
		data, err := r.CallAPI("/v5/corp/department/update", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Department Delete — POST /v5/corp/department/delete
var departmentDelete = common.Shortcut{
	Service:     "jodoo",
	Command:     "+department-delete",
	Description: "Delete a department (must be empty)",
	Risk:        "high-risk-write",
	HasFormat:   true,
	Flags:       []common.Flag{{Name: "dept-no", Type: "int", Required: true, Desc: "department ID"}},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/corp/department/delete").Set("dept_no", r.Int("dept-no"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		data, err := r.CallAPI("/v5/corp/department/delete", map[string]interface{}{
			"dept_no": r.Int("dept-no"),
		})
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Department Batch Import — POST /v5/corp/department/batch_import
var departmentBatchImport = common.Shortcut{
	Service:     "jodoo",
	Command:     "+department-batch-import",
	Description: "Replace the department tree (keyed by dept_no)",
	Risk:        "high-risk-write",
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "departments", Required: true,
			Desc:  "JSON array of {dept_no,name,parent_no?}",
			Input: []string{common.File, common.Stdin}},
	},
	Validate: func(_ context.Context, r *common.RuntimeContext) error {
		_, err := common.ParseJSONArray(r.Str("departments"), "departments")
		return err
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		arr, _ := common.ParseJSONArray(r.Str("departments"), "departments")
		return common.NewDryRunAPI("/v5/corp/department/batch_import").Set("departments", arr)
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		arr, err := common.ParseJSONArray(r.Str("departments"), "departments")
		if err != nil {
			return err
		}
		data, err := r.CallAPI("/v5/corp/department/batch_import", map[string]interface{}{
			"departments": arr,
		})
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// ── Role APIs ──

// Role List — POST /v5/corp/role/list
var roleList = common.Shortcut{
	Service:     "jodoo",
	Command:     "+role-list",
	Description: "List enterprise roles",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		common.SkipFlag(),
		common.LimitFlag(50),
		{Name: "has-internal", Type: "bool", Default: "true", Desc: "include built-in roles"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/corp/role/list").
			Set("skip", r.Int("skip")).
			Set("limit", common.ParseIntBounded(r, "limit", 1, 100)).
			Set("has_internal", r.Bool("has-internal"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"skip":         r.Int("skip"),
			"limit":        common.ParseIntBounded(r, "limit", 1, 100),
			"has_internal": r.Bool("has-internal"),
		}
		data, err := r.CallAPI("/v5/corp/role/list", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Role Create — POST /v5/corp/role/create
var roleCreate = common.Shortcut{
	Service:     "jodoo",
	Command:     "+role-create",
	Description: "Create a custom role under a role group",
	Risk:        "write",
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "name", Required: true, Desc: "role name"},
		{Name: "group-no", Type: "int", Required: true, Desc: "role group number"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/corp/role/create").
			Set("name", r.Str("name")).
			Set("group_no", r.Int("group-no"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		data, err := r.CallAPI("/v5/corp/role/create", map[string]interface{}{
			"name":     r.Str("name"),
			"group_no": r.Int("group-no"),
		})
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// Role Member List — POST /v5/corp/role/member/list
var roleMemberList = common.Shortcut{
	Service:     "jodoo",
	Command:     "+role-member-list",
	Description: "List members under a role",
	Risk:        "read",
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "role-no", Type: "int", Required: true, Desc: "role number"},
		common.SkipFlag(),
		common.LimitFlag(100),
		{Name: "has-manage-range", Type: "bool", Default: "false",
			Desc: "include department-range info"},
	},
	DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI("/v5/corp/role/member/list").
			Set("role_no", r.Int("role-no")).
			Set("skip", r.Int("skip")).
			Set("limit", common.ParseIntBounded(r, "limit", 1, 100)).
			Set("has_manage_range", r.Bool("has-manage-range"))
	},
	Execute: func(_ context.Context, r *common.RuntimeContext) error {
		body := map[string]interface{}{
			"role_no":          r.Int("role-no"),
			"skip":             r.Int("skip"),
			"limit":            common.ParseIntBounded(r, "limit", 1, 100),
			"has_manage_range": r.Bool("has-manage-range"),
		}
		data, err := r.CallAPI("/v5/corp/role/member/list", body)
		if err != nil {
			return err
		}
		r.OutFormat(data, nil, nil)
		return nil
	},
}

// deptsAsInts parses a string-array of department numbers into []int.
// Non-integer entries are silently skipped.
func deptsAsInts(in []string) []int {
	if len(in) == 0 {
		return nil
	}
	out := make([]int, 0, len(in))
	for _, s := range in {
		var n int
		if _, err := fmt.Sscanf(s, "%d", &n); err == nil {
			out = append(out, n)
		}
	}
	return out
}

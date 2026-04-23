# Jodoo API - Workflow APIs

Workflow APIs enable you to manage workflow instances and tasks (approval processes).

---

## Workflow Instances Query API

Query workflow instance details.

- **URL**: `POST /v5/workflow/instance/get`
- **Rate Limit**: 30 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `instance_id` | String | Yes | Instance ID (same as `data_id`) |
| `tasks_type` | Number | No | 0 = Don't return tasks, 1 = Return all tasks |

### Response Parameters

| Parameter | Type | Description |
|---|---|---|
| `url` | String | Instance access link |
| `instance_id` | String | Instance ID |
| `app_id` | String | App ID |
| `form_id` | String | Form ID |
| `form_title` | String | Form Name |
| `update_time` | String | Modification time |
| `create_time` | String | Created time |
| `creator` | Object | Creator info (member entity) |
| `status` | Number | 0=In Progress, 1=Completed, 2=End Manually |
| `tasks` | Object[] | Task list |
| `tasks[].task_id` | String | Task ID |
| `tasks[].flow_id` | Number | Node ID |
| `tasks[].flow_name` | String | Node name |
| `tasks[].url` | String | Task access link |
| `tasks[].assignee` | Object[] | Assignee info (member entity) |
| `tasks[].creator` | Object | Instance creator info |
| `tasks[].create_time` | String | Task start time |
| `tasks[].create_action` | String | Task creation action (see Action Table) |
| `tasks[].finish_time` | String | Task end time |
| `tasks[].finish_action` | String | Task completion action |
| `tasks[].status` | Number | 0=In Progress, 1=Completed, 2=End Manually, 4=Activated, 5=Paused |

### Workflow Action Values

| Action | Description |
|---|---|
| `auto_approve` | Remove duplicate approvers, keep highest-level |
| `forward` | Submit |
| `back` | Return |
| `close` | Close |
| `transfer` | Hand over |
| `revoke` | Withdraw |
| `activate` | Activate |
| `auto_forward` | Auto submit overdue tasks |
| `auto_back` | Auto return overdue tasks |
| `batch_forward` | Batch submit |
| `batch_transfer` | Batch change approvers |
| `sign_before` | Add pre-approver |
| `sign_after` | Add post-approver |
| `sign_parallel` | Add parallel approver |
| `invoke_plugin` | Run a plugin |

---

## Workflow Instance Logs Query API

Query workflow processing logs (currently supports approval comments only).

- **URL**: `POST /v5/workflow/instance/logs`
- **Rate Limit**: 30 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `instance_id` | String | Yes | Instance ID (same as data_id) |
| `types` | String[] | Yes | Log types: `["comment"]` |
| `limit` | Number | No | Page size (max 100). Default: 100 |
| `skip` | Number | No | Records to skip. Default: 0 |

### Response Parameters

| Parameter | Type | Description |
|---|---|---|
| `logs[].flow_id` | Number | Node ID |
| `logs[].flow_name` | String | Node name |
| `logs[].create_action` | String | Create action |
| `logs[].create_time` | Date | Start time |
| `logs[].finish_action` | String | Completion action |
| `logs[].finish_time` | Date | End time |
| `logs[].comment` | String | Approval comment |
| `logs[].signature.url` | String | Signature URL (valid 6 days) |
| `logs[].attachments[]` | Object[] | Attachment list |
| `logs[].operator` | Object | Approver info (member entity) |

---

## Workflow Instances Reactivate API

Reactivate a completed/ended workflow instance.

- **URL**: `POST /v5/workflow/instance/activate`
- **Rate Limit**: 20 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `instance_id` | String | Yes | Instance ID to reactivate |
| `flow_id` | Number | Yes | Node ID to reactivate |

---

## Workflow Instances End API

End (terminate) a workflow instance. Admin only.

- **URL**: `POST /v5/workflow/instance/close`
- **Rate Limit**: 20 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `instance_id` | String | Yes | Instance ID to end |

---

## Workflow Tasks Query API

Query the workflow tasks of a specific user.

- **URL**: `POST /v5/workflow/task/list`
- **Rate Limit**: 30 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `username` | String | Yes | User No. in Contacts |
| `limit` | Number | No | Page size (max 100). Default: 10 |
| `skip` | Number | No | Records to skip. Default: 0 |

### Response Parameters

| Parameter | Type | Description |
|---|---|---|
| `has_more` | Boolean | Whether more data is available |
| `tasks[].app_id` | String | App ID |
| `tasks[].form_id` | String | Form ID |
| `tasks[].task_id` | String | Task ID |
| `tasks[].instance_id` | String | Instance ID (= data_id) |
| `tasks[].form_title` | String | Form name |
| `tasks[].title` | String | Task name |
| `tasks[].flow_id` | Number | Node ID |
| `tasks[].flow_name` | String | Node name |
| `tasks[].url` | String | Task access link |
| `tasks[].assignee` | Object | Assignee (member entity) |
| `tasks[].creator` | Object | Creator (member entity) |
| `tasks[].create_time` | String | Start time |
| `tasks[].create_action` | String | Creation action |
| `tasks[].finish_time` | String | End time |
| `tasks[].finish_action` | String | Completion action |
| `tasks[].status` | Number | Task status |

---

## Workflow Tasks Submit API (Approve)

Submit/approve a workflow task.

- **URL**: `POST /v5/workflow/task/forward`
- **Rate Limit**: 20 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `username` | String | Yes | User No. in Contacts |
| `instance_id` | String | Yes | Instance ID (= data_id) |
| `task_id` | String | Yes | Task ID (corresponds to username) |
| `comment` | String | No | Approval comment |

---

## Workflow Task Rejection API

Reject a workflow task.

- **URL**: `POST /v5/workflow/task/reject`
- **Rate Limit**: 20 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `instance_id` | String | Yes | Instance ID |
| `task_id` | String | Yes | Task ID |
| `username` | String | Yes | Username associated with the task |
| `comment` | String | No | Approval comment |

---

## Workflow Tasks Return API

Return a workflow task to a previous node.

- **URL**: `POST /v5/workflow/task/back`
- **Rate Limit**: 20 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `username` | String | Yes | User No. in Contacts |
| `instance_id` | String | Yes | Instance ID (= data_id) |
| `task_id` | String | Yes | Task ID |
| `flow_id` | Number | No | Target node to return to. If empty, returns to previous node |
| `comment` | String | No | Approval comment |

---

## Workflow Tasks Transfer API

Transfer a workflow task to another user.

- **URL**: `POST /v5/workflow/task/transfer`
- **Rate Limit**: 20 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `username` | String | Yes | Current assignee User No. |
| `instance_id` | String | Yes | Instance ID (= data_id) |
| `task_id` | String | Yes | Task ID |
| `transfer_username` | String | Yes | Target user to transfer to |
| `comment` | String | No | Approval comment |

---

## Workflow Tasks Withdraw API

Withdraw a previously submitted workflow task.

- **URL**: `POST /v5/workflow/task/revoke`
- **Rate Limit**: 20 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `instance_id` | String | Yes | Instance ID |
| `task_id` | String | No | Task ID to withdraw (if empty, withdraws start node) |
| `username` | String | Yes | Username of the task owner |

---

## Node Approvers Add API

Add approvers to a workflow node.

- **URL**: `POST /v5/workflow/task/add_sign`
- **Rate Limit**: 20 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `instance_id` | String | Yes | Instance ID |
| `task_id` | String | Yes | Task ID |
| `username` | String | Yes | Current task assignee username |
| `comment` | String | No | Approval comment |
| `add_sign_type` | Number | Yes | 0=Pre-approver, 1=Post-approver, 2=Parallel approver |
| `add_sign_username` | String | Yes | Username of approver to add |

---

## Query CC List API

Query CC (carbon copy) notifications list. Only supports CC data within 90 days.

- **URL**: `POST /v5/workflow/cc/list`
- **Rate Limit**: 5 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `username` | String | Yes | User No. in Contacts |
| `skip` | Number | No | Records to skip. Default: 0 |
| `limit` | Number | No | Page size (max 100). Default: 10 |
| `read_status` | String | No | `"read"`, `"unread"`, `"all"` (default) |

### Response Parameters

| Parameter | Type | Description |
|---|---|---|
| `has_more` | Boolean | More data available |
| `cc_list[].task_id` | String | CC ID |
| `cc_list[].instance_id` | String | Instance ID (= data_id) |
| `cc_list[].status` | Number | 0=Unread, 1=Read |

---

## Approval Comments Query API

Query approval comments for a workflow form record.

- **URL**: `POST /v5/app/{appId}/entry/{entryId}/data/{dataId}/approval_comments`
- **Rate Limit**: 30 times/second

### URL Parameters

| Parameter | Type | Description |
|---|---|---|
| `appId` | String | App ID |
| `entryId` | String | Form ID |
| `dataId` | String | Workflow form data ID |

### Response

| Parameter | Type | Description |
|---|---|---|
| `approveCommentList[].flowNodeName` | String | Workflow node name |
| `approveCommentList[].flowAction` | String | Action: `forward`, `transfer`, `back`, `close`, `sign_before`, `sign_after`, `sign_parallel` |
| `approveCommentList[].comment` | String | Approval comment |
| `approveCommentList[].signature_url` | String | Optional signature URL (valid 15 days) |

---

## Standard Workflow Response

All workflow action APIs return:

| Parameter | Type | Description |
|---|---|---|
| `status` | String | `"success"` or `"failure"` |
| `code` | Number | Error code (only on failure) |
| `message` | String | Error description (only on failure) |

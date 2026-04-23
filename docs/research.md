Tôi đã thu thập đủ dữ liệu. Bây giờ tôi sẽ tổng hợp toàn bộ thành hệ thống file .md hoàn chỉnh cho Jodoo SDK. Dưới đây là kết quả:

---

# Hệ thống tài liệu API Jodoo - Toàn bộ

Dưới đây là toàn bộ hệ thống tài liệu API được tổ chức thành **7 file .md** riêng biệt, sẵn sàng cho việc build Jodoo SDK.

---

## 📁 Cấu trúc thư mục đề xuất

```
jodoo-sdk-docs/
├── 01-developer-guide.md
├── 02-app-apis.md
├── 03-form-data-apis.md
├── 04-file-api.md
├── 05-workflow-apis.md
├── 06-contact-apis.md
├── 07-webhook.md
└── 08-error-codes.md
```

---

## 📄 File 1: `01-developer-guide.md`

```markdown
# Jodoo API - Developer's Guide

## Base URL

```
https://api.jodoo.com/api
```

> **Note**: This is not a direct URL. It should be used in conjunction with parameters such as `app_id`, `entry_id`, etc.

## API Versioning

Starting from V2, Jodoo will introduce (never reduce) new input and output parameters while maintaining compatibility with previous versions. Developers should anticipate the addition of new parameters and ensure the stability of their codes.

Current API version paths: `/v1/`, `/v2/`, `/v3/`, `/v4/`, `/v5/`

## Global Rate Limit

- **Maximum global request acceptance rate**: 50 requests per second
- Per-API rate limits are specified in each API's documentation

## Authentication

Jodoo uses **simple token-based authentication** (Bearer Token).

### Setting Up API Key

1. Go to **Open Platform** > **API Key** > **Create API Key**
2. One company can generate up to **500 API keys**
3. You can enable, disable, or delete existing API keys

### Authentication Header

```http
Authorization: Bearer YOUR_APIKEY
```

## Request Requirements

| Requirement | Details |
|---|---|
| Protocol | All requests must be via **HTTPS** |
| Method | All requests use **POST** method |
| Encoding | Data transmission encoding: **UTF-8** |
| Format | Use `JSON` for all APIs except file upload (use `form_data`) |

## Response Format

All APIs follow a response format of `code + error message` to indicate errors.

### Success Response

HTTP status code: **2xx**

### Error Response

HTTP status code: **400**

```json
{
  "code": 8301,
  "msg": "Failed to verify the API key for authorization."
}
```

## API Key Permission Scope

The API Key is scoped to specific apps. You configure which apps an API key can access in **Open Platform** > **API Key**.

## Field Name Convention

Once a form field is added, it is represented by a fixed **field ID** with the prefix `widget` (e.g., `_widget_1529400746031`). The field ID will remain the same no matter how you modify the field.

Users can set **field aliases** via **Extension** > **Webhook** > **Set Field Alias**. If aliases are set, they will be used as field names in all API requests/responses.

## Time Format Support

| Format | Example | Note |
|---|---|---|
| ISO Date Format | `'2018-11-09T10:00:00Z'`, `'2018-11-09T10:00:00'` | `Z` suffix = UTC timezone |
| Milliseconds Timestamp | `1639106951523` | Must be milliseconds, not seconds |
| RFC 3339 / yyyy-MM-dd HH:mm:ss | `'2021-10-10 10:10:10'`, `'2021-10-10'` | - |
```

---

## 📄 File 2: `02-app-apis.md`

```markdown
# Jodoo API - App APIs

App APIs enable you to query apps and forms that are within the API Key permission scope.

## User App Query API

Query the apps within the API key permission scope.

- **URL**: `POST /v5/app/list`
- **Rate Limit**: 30 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `limit` | Number | No | Number of records per request (1-100). Default: 100 |
| `skip` | Number | No | Number of records to skip. Default: 0 |

### Request Example

```json
{
  "limit": 100,
  "skip": 0
}
```

### Response

Returns a list of apps with `app_id` and app information.

---

## User Form Query API

Query all forms under a specific app.

- **URL**: `POST /v5/app/entry/list`
- **Rate Limit**: 30 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `app_id` | String | Yes | App ID |
| `limit` | Number | No | Number of records per request (0-100). Default: 100 |
| `skip` | Number | No | Number of records to skip. Default: 0 |

### Request Example

```json
{
  "app_id": "5b237267b22ab14884086c49",
  "limit": 100,
  "skip": 0
}
```

### Response

Returns a list of forms with `entry_id`, form name, and other metadata.
```

---

## 📄 File 3: `03-form-data-apis.md`

```markdown
# Jodoo API - Form & Data APIs

All API paths include `app_id` and `entry_id`, which represent the application ID and form ID. Together `app_id + entry_id` represents a globally unique form ID.

---

## Reference Table: Field & Data Types

### Form Fields

| Field Name | Field Type | Data Type | Data Sample | Notes |
|---|---|---|---|---|
| Single Line | `text` | String | `"Martin"` | - |
| Multi Line | `textarea` | String | `"I love Jodoo"` | - |
| Serial No. | `sn` | String | `"00001"` | - |
| Number | `number` | Number | `10` | - |
| Date&Time | `datetime` | String | `"2018-01-01T10:10:10.000Z"` | UTC format |
| Radio | `radiogroup` | String | `"First grade"` | - |
| Checkbox | `checkboxgroup` | Array | `["Option 1", "Option 2"]` | - |
| Single Select | `combo` | String | `"Female"` | - |
| Multi Select | `combocheck` | Array | `["Option 1", "Option 2"]` | - |
| Image | `image` | Array | `[{"name":"img.png","size":262144,"mime":"image/png","url":"..."}]` | URL valid 15 days |
| Attachment | `upload` | Array | `[{"name":"doc.pdf","size":524288,"mime":"application/pdf","url":"..."}]` | URL valid 15 days |
| Subform | `subform` | Array | `[{"_id":"...","_widget_xxx":"..."}]` | `_id` = subform data ID |
| Select Data | `linkdata` | JSON | `{"id":"...","key":"Jodoo"}` | `id`=associated data ID |
| Signature | `signature` | JSON | `{"name":"sig.png","size":1024,"mime":"image/png","url":"..."}` | URL valid 15 days |
| Member | `user` | JSON | `{"name":"Peach","username":"Martin","status":1,"type":0,"departments":[1,3]}` | status: -1=Inactive, 0=Not joined, 1=Joined |
| Members | `usergroup` | Array | Array of member JSON objects | Same as Member |
| Department | `dept` | JSON | `{"name":"Finance","dept_no":1,"type":0,"parent_no":2,"status":1}` | - |
| Departments | `deptgroup` | Array | Array of department JSON objects | - |

### System Fields

| System Field | Field Name | Data Type | Description |
|---|---|---|---|
| App ID | `appId` | String | Global unique ID |
| Form ID | `entryId` | String | `appId + entryId` = unique form |
| Data ID | `_id` | String | Global unique data ID |
| URL Parameter | `ext` | String | - |
| Created Time | `createTime` | String | UTC format |
| Created User | `creator` | JSON | Member entity |
| Updated Time | `updateTime` | String | UTC format |
| Updated User | `updater` | JSON | Member entity |
| Deleted User | `deleter` | JSON | Member entity |
| Workflow Status | `flowState` | Number | 0=In Progress, 1=Completed, 2=Manual Close |

---

## Form Fields Query API

Query the field list of a specified form.

- **URL**: `POST /v5/app/entry/widget/list`
- **Rate Limit**: 30 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `app_id` | String | Yes | App ID |
| `entry_id` | String | Yes | Form ID |

### Response

| Parameter | Description |
|---|---|
| `widgets` | Field information array |
| `widgets[].label` | Field title |
| `widgets[].name` | Field name (alias if set; otherwise field ID) |
| `widgets[].type` | Field type |
| `widgets[].items` | Subform only: array of subfield info |
| `sysWidgets` | List of system fields |
| `sysWidgets[].name` | System field name |
| `dataModifyTime` | Latest data update time of the form |

---

## Single Record Query API

Query one record from a form by data ID.

- **URL**: `POST /v5/app/entry/data/get`
- **Rate Limit**: 30 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `app_id` | String | Yes | App ID |
| `entry_id` | String | Yes | Form ID |
| `data_id` | String | Yes | Data ID |

### Request Example

```json
{
  "app_id": "5b237267b22ab14884086c49",
  "entry_id": "5b237267b22ab14884086cc9",
  "data_id": "5b237267b22ab14884086c50"
}
```

---

## Multiple Records Query API

Query multiple records from a form with filtering support.

- **URL**: `POST /v5/app/entry/data/list`
- **Rate Limit**: 5 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `app_id` | String | Yes | App ID |
| `entry_id` | String | Yes | Form ID |
| `data_id` | String | No | Last record ID from previous query (for pagination) |
| `fields` | Array | No | Fields to return |
| `filter` | JSON | No | Data filter conditions |
| `limit` | Number | No | Records per query (1-100). Default: 10 |

### Filter Structure

```json
{
  "filter": {
    "rel": "and",
    "cond": [
      {
        "field": "widget_xxx",
        "type": "text",
        "method": "eq",
        "value": ["some value"]
      }
    ]
  }
}
```

#### Filter `rel` Values

| Value | Description |
|---|---|
| `and` | Meet ALL filter conditions |
| `or` | Meet ANY filter condition |

#### Filter `method` Values

| Method | Description |
|---|---|
| `not_empty` | Not empty |
| `empty` | Empty |
| `eq` | Equal |
| `in` | Equal to any (max 200 values) |
| `range` | Between x and y (inclusive) |
| `nin` | Not equal to any (max 200 values) |
| `ne` | Not equal |
| `like` | Text included |
| `gt` | Greater than (Number only) |
| `lt` | Less than (Number only) |
| `all` | Contains all (Checkbox/Multi Select) |
| `verified` | Email filled and verified |
| `unverified` | Email filled but not verified |

#### Supported Filter Fields

| Field Type | Supported Methods |
|---|---|
| `flowState` | `eq`, `ne`, `in`, `nin`, `empty`, `not_empty` |
| `data_id` | `eq`, `in`, `empty`, `not_empty` |
| Submitter | `eq`, `ne`, `in`, `nin`, `empty`, `not_empty` |
| Date&Time | `eq`, `ne`, `range`, `empty`, `not_empty` |
| Number | `eq`, `ne`, `range`, `empty`, `not_empty`, `gt`, `lt` |
| Text (Single Line, Radio, Single Select, Serial No.) | `eq`, `ne`, `in`, `nin`, `empty`, `not_empty` |
| Checkbox / Multi Select | `in`, `all`, `empty`, `not_empty` |
| Phone | `like`, `verified`, `unverified`, `empty`, `not_empty` |
| Member / Department | `eq`, `ne`, `in`, `nin`, `empty`, `not_empty` |
| Members / Departments | `in`, `all`, `empty`, `not_empty` |
| Serial No. | `like`, `empty`, `not_empty` |
| Lookup | `eq`, `ne`, `in`, `nin`, `empty`, `not_empty` |

### Pagination (Looped Retrieval)

Use the `data_id` parameter as a cursor marker. Records are always returned sorted in ascending order by data ID. When the returned count is less than `limit`, all data has been retrieved.

---

## Single Record Creation API

Add one record to a form.

- **URL**: `POST /v5/app/entry/data/create`
- **Rate Limit**: 20 times/second

### Request Parameters

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `app_id` | String | Yes | - | App ID |
| `entry_id` | String | Yes | - | Form ID |
| `data` | JSON | Yes | - | Record data |
| `data_creator` | String | No | Business owner | Member NO. of the submitter |
| `is_start_workflow` | Bool | No | `false` | Whether to initiate workflows |
| `is_start_trigger` | Bool | No | `false` | Whether to trigger Automations |
| `transaction_id` | String | No | - | For binding uploaded files |

### Request Example

```json
{
  "app_id": "5b237267b22ab14884086c49",
  "entry_id": "5b237267b22ab14884086cc9",
  "data": {
    "_widget_1529400746031": "Hello Jodoo",
    "_widget_1529400746032": 100
  },
  "is_start_workflow": false,
  "is_start_trigger": true
}
```

---

## Multiple Records Creation API

Add multiple records to a form (max 100 per request).

- **URL**: `POST /v5/app/entry/data/batch_create`
- **Rate Limit**: 10 times/second

### Request Parameters

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `app_id` | String | Yes | - | App ID |
| `entry_id` | String | Yes | - | Form ID |
| `data_list` | Array | Yes | - | Array of record data objects |
| `data_creator` | String | No | Business owner | Member NO. of the submitter |
| `transaction_id` | String | No | - | For idempotency and file binding |
| `is_start_workflow` | Bool | No | `false` | Whether to initiate workflows |

### Response

| Parameter | Type | Description |
|---|---|---|
| `status` | String | Request result |
| `success_count` | Number | Number of records successfully added |
| `success_ids` | Array | List of IDs of successfully added records |

### Retry Logic

If batch creation partially fails, use the **same `transaction_id`** to retry. Previously successful records will NOT be duplicated; only failed records will be added.

---

## Single Record Update API

Update one record in a form by data ID.

- **URL**: `POST /v5/app/entry/data/update`
- **Rate Limit**: 20 times/second

### Request Parameters

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `app_id` | String | Yes | - | App ID |
| `entry_id` | String | Yes | - | Form ID |
| `data_id` | String | Yes | - | Data ID |
| `data` | JSON | Yes | - | Updated data content |
| `is_start_trigger` | Bool | No | `false` | Whether to trigger Automations |
| `transaction_id` | String | No | - | For file binding |

### Important: Subform Update Behavior

- If subform data ID is **not passed or incorrect**: Original subform data is cleared and replaced with new data; new IDs are generated.
- If subform data ID is **correct**: The data ID remains unchanged.
- You **cannot** update a single row in a subform; you must pass the **entire** subform data structure.

---

## Multiple Records Update API

Batch update multiple records.

- **URL**: `POST /v5/app/entry/data/batch_update`
- **Rate Limit**: 10 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `app_id` | String | Yes | App ID |
| `entry_id` | String | Yes | Form ID |
| `data_ids` | Array | Yes | Array of data IDs to update |
| `data` | JSON | Yes | Data to apply to all records |
| `transaction_id` | String | No | For file binding |

> **Note**: Cannot be used with subform fields.

---

## Single Record Deletion API

Delete one record by data ID.

- **URL**: `POST /v5/app/entry/data/delete`
- **Rate Limit**: 20 times/second

### Request Parameters

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `app_id` | String | Yes | - | App ID |
| `entry_id` | String | Yes | - | Form ID |
| `data_id` | String | Yes | - | Data ID |
| `is_start_trigger` | Bool | No | `false` | Whether to trigger Automations |

---

## Multiple Records Deletion API

Delete multiple records by data IDs.

- **URL**: `POST /v5/app/entry/data/batch_delete`
- **Rate Limit**: 10 times/second

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `app_id` | String | Yes | App ID |
| `entry_id` | String | Yes | Form ID |
| `data_ids` | String[] | Yes | Array of data IDs to delete |

---

## API Operation Trigger Table

| Feature | create | update | delete | batch_create | batch_update |
|---|---|---|---|---|---|
| Delayed computation (data factory) | ✅ | ✅ | ✅ | ✅ | ✅ |
| Data message push | ✅ | ✅ | ❌ | ✅ | ✅ |
| Aggregate table | ✅ | ✅ | ✅ | ✅ | ✅ |
| Data action history | ✅ | ✅ | ✅ | ✅ | ✅ |
| Webhook | ❌ | ❌ | ❌ | ❌ | ❌ |
| Automations | ✅ | ✅ | ✅ | ❌ | ❌ |
| Duplicate values validation | ❌ | ❌ | - | ❌ | ❌ |
| Form validations | ❌ | ❌ | - | ❌ | ❌ |
| Required field validation | ❌ | ❌ | - | ❌ | ❌ |
| Workflow node validation | ❌ | ❌ | - | ❌ | ❌ |
| Trigger workflow | ✅ | ❌ | - | ✅ | ❌ |
| Aggregate table validation | ✅ | ✅ | ✅ | ❌ | ❌ |
| Data linkage / formula | ❌ | ❌ | - | ❌ | ❌ |
| Form notifications | ✅ | ✅ | - | ❌ | ❌ |
```

---

## 📄 File 4: `04-file-api.md`

```markdown
# Jodoo API - File API

The File API allows you to upload files (images, attachments) to be used in form records.

The upload process has 2 steps:
1. **Get upload credentials and URL** → returns `token` + `url`
2. **Upload the file** using the token and url

---

## Step 1: File Upload Credentials and URL Get API

Get credentials and URL for file upload.

- **URL**: `POST /v5/app/entry/file/get_upload_token`
- **Rate Limit**: 20 times/second

### Description

Each request returns **100** file upload credentials and URLs. Uploaded files are associated with the `transaction_id` and can only be used in creation/update requests with the **same** `transaction_id`.

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `app_id` | String | Yes | App ID |
| `entry_id` | String | Yes | Form ID |
| `transaction_id` | String | Yes | Transaction ID (UUID recommended) |

### Request Example

```json
{
  "app_id": "5b237267b22ab14884086c49",
  "entry_id": "5b237267b22ab14884086cc9",
  "transaction_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

### Response

| Parameter | Type | Description |
|---|---|---|
| `token_and_url_list` | JSON | File upload credentials and URL list |
| `token_and_url_list[].url` | String | File upload URL |
| `token_and_url_list[].token` | String | File upload credential |

---

## Step 2: File Upload API

Upload a single file using the credentials from Step 1.

- **URL**: `POST {url}` (the `url` returned from Step 1)
- **Rate Limit**: 20 times/second
- **Content-Type**: `multipart/form-data`

### Important

- Only **one file** per token
- Cannot overwrite
- The returned `key` should be used to fill in the attachment/image field value when creating or updating records

### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `token` | String | Yes | File upload credential (from Step 1) |
| `file` | File | Yes | The file to upload |

### Response

Returns the `key` to use in the data field value for attachment/image.

---

## Usage in Create/Update APIs

When creating or updating records with file fields, the `transaction_id` in the data API must match the `transaction_id` used in the File Upload Credentials API.

```json
{
  "app_id": "...",
  "entry_id": "...",
  "transaction_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "data": {
    "_widget_file_field": [
      { "key": "returned_key_from_upload" }
    ]
  }
}
```
```

---

## 📄 File 5: `05-workflow-apis.md`

```markdown
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
```

---

## 📄 File 6: `06-contact-apis.md`

```markdown
# Jodoo API - Contact APIs

Contact APIs enable you to manage members, departments, and roles in your organization.

---

## Entity Structures

### Department Entity

| Property | Type | Description | Note |
|---|---|---|---|
| `dept_no` | Number | Department number (unique within enterprise) | - |
| `parent_no` | Number | Parent department number | - |
| `type` | Number | Department type | 0 = Regular department |
| `status` | Number | Department status | 1 = In use |
| `seq` | Number | Sort order (ascending within parent) | - |
| `name` | String | Department name | - |

### Member Entity

| Property | Type | Description | Note |
|---|---|---|---|
| `username` | String | Member number (unique within company) | - |
| `name` | String | Nickname | - |
| `departments` | Number[] | List of department numbers | - |
| `type` | Number | Member type | 0 = Regular member |
| `status` | Number | Status | 0=Unconfirmed, 1=Joined |
| `integrate_id` | String | Integration ID | - |

### Role Entity

| Property | Type | Description | Note |
|---|---|---|---|
| `role_no` | Number | Role ID (unique within enterprise) | - |
| `group_no` | Number | Role group number | - |
| `type` | Number | Role type | 0 = Regular role |
| `status` | Number | Status | 1 = In use |
| `name` | String | Role name | - |

### Role Group Entity

| Property | Type | Description |
|---|---|---|
| `group_no` | Number | Role group number (unique within enterprise) |
| `type` | Number | 0 = Regular group |
| `status` | Number | 1 = In use |
| `name` | String | Role group name |

> **Note**: The root department number is always `1`.

---

## Member APIs

### Member Retrieval (Recursively) API

Recursively get all members under a department.

- **URL**: `POST /v5/corp/member/list`
- **Rate Limit**: 10 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `dept_no` | Number | Yes | Department number |
| `has_child` | Boolean | No | Whether to recursively get sub-department members. Default: false |

**Response**: Array of member entities.

---

### Member Information Query API

Get member info by username.

- **URL**: `POST /v5/corp/member/get`
- **Rate Limit**: 30 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `username` | String | Yes | User No. in Contacts |

**Response**:
```json
{
  "user": {
    "username": "jodoo",
    "name": "Harry",
    "departments": [1, 3],
    "type": 0,
    "status": 1,
    "integrate_id": "jodoo"
  }
}
```

---

### Member Adding API

Create a member under a department.

- **URL**: `POST /v5/corp/member/create`
- **Rate Limit**: 20 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `name` | String | Yes | Nickname |
| `username` | String | Yes | Member number (letters, digits, underscores only) |
| `departments` | Number[] | No | Department list |

> Members created via API are auto-activated and can access Jodoo through SSO.

---

### Member Update API

Update member information (department, nickname).

- **URL**: `POST /v5/corp/member/update`
- **Rate Limit**: 20 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `username` | String | Yes | Member ID |
| `name` | String | No | New nickname |
| `departments` | Number[] | No | New department list |

---

### Member Deletion API

Delete (deactivate) a member.

- **URL**: `POST /v5/corp/member/delete`
- **Rate Limit**: 20 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `username` | String | Yes | Member No. |

> Deleted members are **deactivated**, not permanently removed. They can be managed on the **Inactive Members** page.

---

### Batch Member Import API

Batch create/update members using `username` as primary key.

- **URL**: `POST /v5/corp/member/batch_import`
- **Rate Limit**: 10 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `users` | JSON[] | Yes | User list |
| `users[].username` | String | Yes | User ID |
| `users[].name` | String | Yes | Nickname |
| `users[].departments` | Number[] | No | Department numbers |

---

## Department APIs

### Department Lists Retrieval (Recursively) API

Retrieve all sub-departments recursively.

- **URL**: `POST /v5/corp/department/list`
- **Rate Limit**: 10 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `dept_no` | Number | Yes | Department number (use `1` for root) |
| `has_child` | Boolean | No | Whether to recurse. Default: false (immediate children only) |

---

### Department Creation API

Create a new department.

- **URL**: `POST /v5/corp/department/create`
- **Rate Limit**: 20 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `name` | String | Yes | Department name (max 32 chars) |
| `parent_no` | Number | No | Parent dept ID (defaults to root) |
| `dept_no` | Number | No | Custom dept ID (auto-generated if empty) |

**Response**: Created department entity.

---

### Department Update API

Update department information.

- **URL**: `POST /v5/corp/department/update`
- **Rate Limit**: 20 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `dept_no` | Number | Yes | Department ID |
| `name` | String | No | New name |
| `parent_no` | Number | No | New parent department ID |

---

### Department Deletion API

Delete a department.

- **URL**: `POST /v5/corp/department/delete`
- **Rate Limit**: 20 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `dept_no` | Number | Yes | Department ID |

---

### Department Batch Import API

Replace entire department tree using `dept_no` as primary key.

- **URL**: `POST /v5/corp/department/batch_import`
- **Rate Limit**: 10 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `departments` | Array | Yes | Department list |
| `departments[].dept_no` | Number | Yes | Department ID (max: 9007199254740991) |
| `departments[].name` | String | Yes | Department name |
| `departments[].parent_no` | Number | No | Parent department ID (defaults to root) |

---

## Role APIs

### Role List Query API

List all roles.

- **URL**: `POST /v5/corp/role/list`
- **Rate Limit**: 30 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `skip` | Number | No | Offset |
| `limit` | Number | No | Limit (default: 1) |
| `has_internal` | Boolean | No | Include built-in roles. Default: true |

---

### Custom Role Creation API

Create an internal custom role.

- **URL**: `POST /v5/corp/role/create`
- **Rate Limit**: 20 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `name` | String | Yes | Role name |
| `group_no` | Number | Yes | Role group number |

---

### Member List Query (Role) API

List members under a specific role.

- **URL**: `POST /v5/corp/role/member/list`
- **Rate Limit**: 30 times/second

| Parameter | Type | Required | Description |
|---|---|---|---|
| `role_no` | Number | Yes | Role number |
| `skip` | Number | No | Offset |
| `limit` | Number | No | Page size |
| `has_manage_range` | Boolean | No | Include assigned department info. Default: false |

**Response**: Array of member entities with optional `departments_range` and `has_child`.
```

---

## 📄 File 7: `07-webhook.md`

```markdown
# Jodoo API - Webhook

Webhook pushes data changes from Jodoo forms to your server via HTTP POST requests.

## Configuration

Go to **Extension** > **Webhook** to configure:

1. **Target Server**: Select "Custom Server"
2. **Server Address**: Your server URL (data pushed via HTTP POST)
3. **Secret**: Click "Generate Secret" for authentication
4. **Push Events**: Select which events trigger pushes

## Push Events

| Event | Description | Notes |
|---|---|---|
| When data is created | Triggered on form submission | Required. Not triggered on batch import |
| When data is updated | Triggered on data update (workflow, editing, admin update) | Optional. Not triggered on batch update |
| When data is deleted | Triggered on data deletion | Optional. Not triggered on batch delete |
| When data is recovered | Triggered on data recovery | Optional. Not triggered on batch recover |
| When form structure changes | Triggered on form structure changes | Optional |

## Push Data Structure

### Data Create / Update / Recover

Full record data is pushed with all field values. See Field Type mapping below.

### Data Delete

| Field | Type | Description |
|---|---|---|
| `formName` | String | Form name |
| `_id` | String | Deleted record ID |
| `deleter` | JSON | Deleted user (member entity) |
| `deleteTime` | String | Deletion time |

## Webhook Field Types

| Field Type | Data Type | Notes |
|---|---|---|
| Single Line | string | - |
| Multi Line | string | - |
| Serial No. | string | - |
| Number | number | - |
| Date&Time | string | - |
| Radio | string | - |
| Checkbox | array | - |
| Single Select | string | - |
| Multi Select | array | - |
| Image | array | URL valid 15 days |
| Attachment | array | URL valid 15 days |
| Signature | json | URL valid 15 days |
| SubForm | array | - |
| Select Data | json | - |
| Member | json | status: -1=Inactive, 0=Pending, 1=Active |
| Members | array | - |
| Department | json | - |
| Departments | array | - |
| Workflow Status | number | 0=In Progress, 1=Done, 2=Manual Close |
| Created User | json | Member entity |
| Updated User | json | Member entity |
| Deleted User | json | Member entity |
| Created Time | string | UTC timestamp |
| Updated Time | string | UTC timestamp |

## Field Alias

Set field aliases in **Extension** > **Webhook** > **Set Field Alias** to replace default `widget_xxx` field names with meaningful names in both webhook pushes and API responses.

## Failed Push Handling

- Admin receives notifications on push failures
- After **100 failed pushes**, the webhook is auto-disabled
- Admin must manually restart via webhook settings
- Options: **Retry** failed cache or **Abandon** cache

## Push History

View push history on the webhook configuration page to see successful/failed pushes with details.
```

---

## 📄 File 8: `08-error-codes.md`

```markdown
# Jodoo API - Error Codes

## General Error Codes

| Code | Message | Description |
|---|---|---|
| 1 | Invalid Request | The request is invalid |
| 8017 | Invalid Company/Team | Company/Team doesn't exist or has been closed |
| 8301 | Invalid Authorization | Failed to verify the API key |
| 8302 | No Permission | No permission for API calls |
| 8303 | Company/Team Request Limit Exceeded | Call frequency exceeds limit |
| 8304 | Request Limit Exceeded | Call frequency exceeds limit |
| 9004 | Creation Failed | Failed to create the task |
| 9007 | Lock Obtain Failed | Failed to obtain the lock |
| 17017 | Invalid Parameter(s) | Invalid parameter(s) |
| 17018 | Invalid API Key | The API key is invalid |
| 17025 | Invalid transaction_id | Invalid transaction_id parameter(s) |
| 17026 | Duplicate transaction_id | The transaction_id already exists |
| 17027 | Uploading Failed | Failed to upload API file |
| 17032 | Invalid Field Type | Field type isn't supported |
| 17034 | Invalid Sub-Field Type | Sub-field type isn't supported |

## Member Error Codes

| Code | Message | Description |
|---|---|---|
| 1005 | Duplicate Email | Email already exists |
| 1010 | Member Not Found | Member does not exist |
| 1017 | Invalid Member Name | Invalid name format |
| 1018 | Invalid Email | Invalid email format |
| 1019 | Invalid Member Nickname | Invalid nickname format |
| 1022 | Invalid Company/Team | Member's company/team doesn't exist |
| 1024 | Invalid Mobile Number | Invalid mobile number |
| 1027 | Duplicate Phone Number | Phone number already exists |
| 1058 | No Permission | No permission for task list |
| 1065 | Duplicate Member | Member already in team |
| 1082 | Phone Number/Email Required | Required field missing |
| 1085 | Nickname Required | Nickname is required |
| 1087 | Duplicate Unique Fields | Duplicate unique fields |
| 1092 | Username Limit Exceeded | Username too long |
| 1096 | Invalid Request Parameter(s) | Invalid member parameter(s) |

## Role Error Codes

| Code | Message | Description |
|---|---|---|
| 1201 | Invalid Role | Role does not exist |
| 1203 | Invalid Role Group/Role | Invalid role group/role info |
| 1205 | Role Group Required | Role group is required |
| 1206 | Invalid Role Group | Role group does not exist |
| 1207 | Role Group Name Limit Exceeded | Name too long |
| 1208 | Role Group with Member(s) | Can't delete group with members |

## App & Form Error Codes

| Code | Message | Description |
|---|---|---|
| 2004 | Invalid App | App does not exist |
| 3000 | Invalid Form | Form does not exist |
| 3001 | Name Required | Name is required |
| 3005 | Invalid Parameter(s) | Invalid parameter(s) |
| 3041 | Serial No. Limit Exceeded | Too many Serial No. fields |
| 3042 | Invalid Field Alias | Failed to verify field alias |
| 3083 | Widget Limit Exceeded | Too many widgets in form |
| 3091 | Form Name Limit Exceeded | Form name exceeds 100 chars |
| 3092 | Title Limit Exceeded | Title exceeds 100 chars |

## Data Error Codes

| Code | Message | Description |
|---|---|---|
| 4000 | Data Submission Failed | Failed to submit data |
| 4001 | Invalid Data | Data does not exist |
| 4042 | Data Deletion Failed | Failed to delete data |
| 4402 | Aggregation Validation Failed | Failed aggregation calculation |
| 4815 | Invalid Filter | Invalid filter condition |
| 17023 | Batch Update Limit Exceeded | Batch update limit exceeded |
| 17024 | Batch Creation Limit Exceeded | Batch creation limit exceeded |

## Workflow Error Codes

| Code | Message | Description |
|---|---|---|
| 4007 | No Approver | No approver for workflow |
| 4008 | Closed Workflow | Workflow was closed |
| 4009 | No Permission | No permission to approve |
| 4015 | Invalid Approver | Transferred node approver invalid |
| 4016 | Action Failed | Can't transfer to oneself |
| 4025 | No Permission | No permission for workflow |
| 5003 | Invalid Node | Workflow node doesn't exist |
| 5004 | No Approval Comment | No comment for approval |
| 5006 | Transfer Error | Can't transfer to CC node |
| 5008 | Approver Not Found | Can't find the approver |
| 5009 | Invalid Approver | Returned node approver invalid |
| 5011 | Invalid Node | Target node is invalid |
| 5012 | No Signature | No signature for approval |
| 5024 | No Permission | No permission at this node |
| 5025 | Node Completed | Current node has been completed |
| 5026 | Batch Approve Failed | Batch approve failed at task node |
| 5034 | Child Workflow(s) Found | Node has child workflows/plugin nodes |
| 5044 | Child Workflow Error | Child workflow wrongly configured |
| 5045 | Child Workflow Limit | Child workflows exceed 200 |
| 5049 | Transfer Disabled | Transfer not enabled at node |
| 5053 | Action Failed | Invalid initial department for multi-level approval |
| 5056 | Invalid Form | Can't perform workflow ops on non-workflow forms |
| 50004 | Invalid Task | Task doesn't exist |
| 50008 | Action Failed | Workflow processing error |
| 50011 | Invalid Task | Task doesn't exist |
| 50013 | Approver Count Mismatch | Approvers before/after must be same count |
| 50014 | Invalid Approver | No permission to close task |
| 50015 | Approver Not Changed | Approver unchanged |
| 50016 | Invalid instance_id | Instance_id is invalid |
| 50019 | Invalid Action | Action not supported |
| 50021 | Invalid data_id | data_id is invalid |
| 50024 | Empty Approver | Approver can't be empty after adjustment |
| 50025 | Batch Limit Exceeded | Single batch approval limit exceeded |
| 50031 | Invalid data_id | data_id is invalid |
| 50040 | Return Failed | This node can't be returned |
| 50041 | Return Failed | Can't return to current node |
| 50046 | Too Many Pending Tasks | Too many pending sub-workflows |
| 50047 | Workflow Being Migrated | Try again later |
| 50049 | Return Failed | Can't return to Start Node without initiator |
| 50051 | Duplicate Request | Workflow already approved/transferred/returned |
| 50052 | Add Approver Failed | Can't add approver for this node |
| 50053 | No Approver | No approver to add |
| 50054 | No Nesting | Can't add nested approval |
| 50055 | Action Failed | No pre/post-approver added |
| 50056 | Invalid Parent Task | Parent task is lost |
| 50057 | Add Approver Disabled | Feature not enabled |
| 50059 | Approver Added | Member is added as approver |
| 50060 | Invalid Approver | Approver is invalid |
| 50061 | Approver Limit | Can't add more than one approver |
| 50062 | Added Approver Reviewing | Added approver is reviewing |
| 50063 | Batch Submit Failed | Unfinished tasks with added approvers |
| 50064 | Loop Detected | Async sub-workflow forms a loop |
| 50070 | Return Failed | Can't return to plugin node |
| 50071 | Plugin Unavailable | Plugin nodes unavailable in current version |
| 50073 | Return Failed | Can't return to in-progress node |

## Department Error Codes

| Code | Message | Description |
|---|---|---|
| 6000 | Duplicate Sub-Department | Same-name sub-department exists |
| 6001 | Invalid Parent Department | Parent department doesn't exist |
| 6002 | Invalid Department | Department doesn't exist |
| 6003 | Department with Sub-Department(s) | Can't delete with sub-departments |
| 6004 | Department Update Failed | Failed to update |
| 6005 | Department Creation Failed | Failed to create |
| 6006 | Department with Member(s) | Can't delete with members |
| 6010 | Invalid Department ID | Invalid ID format |
| 6011 | Circular Department Relationship | Circular relationship detected |
| 6012 | Invalid Department Name | Name is invalid |
| 6013 | Duplicate Department ID | ID already exists |
| 6014 | No Sub-Department Under Root | Must have at least one sub-dept |
| 6015 | Root Department | Can't delete root department |
| 6017 | Cascading Levels Limit | Cascade levels exceed limit |
| 6019 | Member List Required | Member list is required |
| 6020 | Imported Departments Limit | Single import departments exceeded |
| 6021 | Imported Members Limit | Single import members exceeded |
| 6064 | Duplicate Department | Parent department already exists |

## Plan/Quota Error Codes

| Code | Message | Description |
|---|---|---|
| 7103 | System Limit Exceeded | System suspended, upgrade required |
| 7212 | Data Limit Exceeded | Monthly data limit reached |
| 7216 | Attachment Upload Limit | Contact link creator |
| 7217 | Attachment Upload Limit | Contact business owner |
| 7218 | Attachment Upload Limit | Contact link creator |
| 7219 | Attachment Upload Limit | Contact business owner |
```

---

## Tóm tắt

Trên đây là **8 file markdown** hệ thống hóa toàn bộ Jodoo API documentation, bao gồm:

| # | File | Nội dung |
|---|---|---|
| 1 | `01-developer-guide.md` | Base URL, authentication, rate limits, request format, field naming, time formats |
| 2 | `02-app-apis.md` | User App Query, User Form Query |
| 3 | `03-form-data-apis.md` | Field types, Form Query, Single/Multiple Record CRUD, Filter, Pagination, Trigger table |
| 4 | `04-file-api.md` | File upload credentials, file upload, integration with data APIs |
| 5 | `05-workflow-apis.md` | Instance query/logs/reactivate/end, Task query/submit/reject/return/transfer/withdraw, Add approver, CC list, Approval comments |
| 6 | `06-contact-apis.md` | Entity structures, Member CRUD/batch, Department CRUD/batch, Role CRUD/query |
| 7 | `07-webhook.md` | Webhook config, push events, data structure, field alias, error handling |
| 8 | `08-error-codes.md` | Toàn bộ ~100+ error codes phân loại theo module |

Toàn bộ tài liệu này đủ để bạn build SDK với đầy đủ type definitions, API client methods, error handling và request/response models. Chúc bạn build SDK thành công!
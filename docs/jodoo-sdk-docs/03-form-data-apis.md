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

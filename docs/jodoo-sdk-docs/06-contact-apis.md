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

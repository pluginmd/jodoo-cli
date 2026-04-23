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

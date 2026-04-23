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

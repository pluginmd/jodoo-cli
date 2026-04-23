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

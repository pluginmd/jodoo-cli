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

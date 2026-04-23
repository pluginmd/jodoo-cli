# Jodoo API Documentation System

The full API reference has been split into **8 separate markdown files** under [`jodoo-sdk-docs/`](./jodoo-sdk-docs/), ready for building the Jodoo SDK.

## 📁 Directory Structure

```
docs/jodoo-sdk-docs/
├── 01-developer-guide.md
├── 02-app-apis.md
├── 03-form-data-apis.md
├── 04-file-api.md
├── 05-workflow-apis.md
├── 06-contact-apis.md
├── 07-webhook.md
└── 08-error-codes.md
```

## 📄 File Index

| # | File | Contents |
|---|---|---|
| 1 | [01-developer-guide.md](./jodoo-sdk-docs/01-developer-guide.md) | Base URL, authentication, rate limits, request format, field naming, time formats |
| 2 | [02-app-apis.md](./jodoo-sdk-docs/02-app-apis.md) | User App Query, User Form Query |
| 3 | [03-form-data-apis.md](./jodoo-sdk-docs/03-form-data-apis.md) | Field types, Form Query, Single/Multiple Record CRUD, Filter, Pagination, Trigger table |
| 4 | [04-file-api.md](./jodoo-sdk-docs/04-file-api.md) | File upload credentials, file upload, integration with data APIs |
| 5 | [05-workflow-apis.md](./jodoo-sdk-docs/05-workflow-apis.md) | Instance query/logs/reactivate/end, Task query/submit/reject/return/transfer/withdraw, Add approver, CC list, Approval comments |
| 6 | [06-contact-apis.md](./jodoo-sdk-docs/06-contact-apis.md) | Entity structures, Member CRUD/batch, Department CRUD/batch, Role CRUD/query |
| 7 | [07-webhook.md](./jodoo-sdk-docs/07-webhook.md) | Webhook config, push events, data structure, field alias, error handling |
| 8 | [08-error-codes.md](./jodoo-sdk-docs/08-error-codes.md) | All ~100+ error codes categorized by module |

This set is sufficient to build an SDK with complete type definitions, API client methods, error handling, and request/response models.

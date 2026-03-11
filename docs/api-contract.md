# Tracking API Contract (Client <-> Server)

Base path: `/v1`

Authentication:
- Required for history endpoints when server `TRACKER_API_KEY` is configured.
- Preferred header: `Authorization: Bearer <token>`.
- Fallback header: `X-API-Key: <token>`.

## Health Check

- Method: `GET`
- Path: `/healthz`
- Auth: none
- Success response:

```json
{
  "status": "ok"
}
```

## Fetch History

- Method: `GET`
- Path: `/v1/history`
- Auth: required (if enabled)
- Success response:

```json
{
  "entries": [
    {
      "gameName": "CS2.exe",
      "date": "2026-03-11",
      "totalTimeSecs": 1200,
      "lastPlayedDate": "2026-03-11T14:31:22Z"
    }
  ]
}
```

## Append History Deltas

- Method: `POST`
- Path: `/v1/history/append`
- Auth: required (if enabled)
- Request body:

```json
{
  "entries": [
    {
      "gameName": "CS2.exe",
      "date": "2026-03-11",
      "totalTimeSecs": 60,
      "lastPlayedDate": "2026-03-11T14:32:22Z"
    }
  ]
}
```

Validation rules:
- `entries` must not be empty.
- Each `gameName` must be non-empty.
- Each `totalTimeSecs` must be greater than `0`.

Success response:
- Status: `204 No Content`

Error responses:
- `400 Bad Request` invalid payload or validation failure.
- `401 Unauthorized` missing/invalid credentials.
- `405 Method Not Allowed` wrong HTTP verb.
- `500 Internal Server Error` persistence failure.

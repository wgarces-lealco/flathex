# Postman Collections — Conventions

> This file is documentation only. Do not import it into Postman.

## File naming

| Pattern | Purpose |
|---|---|
| `<aggregate>.collection.json` | One collection per core aggregate |
| `environment.<env>.json` | One environment per deployment target |

## Folder structure inside every collection

| Folder | Contents |
|---|---|
| `CRUD` | Basic create / read / list / update / delete |
| `Lifecycle` | State-machine transitions specific to the aggregate |
| `Cross-Domain` | Requests that exercise Rule 5 (handler-level orchestration) |
| `Error Cases` | Requests prefixed with ❌ that validate domain errors → HTTP codes |

## Environment variables

| Variable | Convention |
|---|---|
| `base_url` | Always read from the active environment. Never hardcode. |
| `<entity>_id` | Auto-captured by the test script in `Create <Entity>` via `pm.environment.set`. All downstream requests reference `{{entity_id}}`. |
| `notify_email` | Used for cross-aggregate notification scenarios. |

## Test script contract

- Every request must assert **at minimum** the expected HTTP status code.
- `Create` requests must capture the returned `id` into `pm.environment.set('<entity>_id', body.id)`.
- `❌` error requests must assert the status code **and** that the response body contains the sentinel error message.

## Adding a new aggregate

1. Create `.postman/<aggregate>.collection.json` following the folder structure above.
2. Add any new shared variables to `environment.local.json` (and other environment files).
3. If the aggregate interacts with existing ones, add a `Cross-Domain` folder to **both** collections showing the handler-level orchestration.
4. Name error-case requests with a `❌` prefix for quick visual scanning.

## Environments

| Name | `base_url` | Notes |
|---|---|---|
| Local | `http://localhost:8080` | In-memory adapters, NoOpSender |
| Staging | `https://api-staging.taskhex.dev` | Postgres + real SMTP |
| Prod | `https://api.taskhex.dev` | Read-only smoke tests only |

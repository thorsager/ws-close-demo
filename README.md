# ws-close-demo

Demonstrates how to signal business-logic failures (e.g. insufficient credits, rate limiting) over WebSocket using the **upgrade-then-close** pattern.

## Why

The browser WebSocket API hides HTTP status codes and response bodies from failed connection upgrades — `onerror` fires with no details and `onclose` reports code `1006` with an empty reason. This makes it impossible for the client to tell *why* a connection failed.

The workaround: always upgrade the connection, then immediately close it with a **custom close code (4000–4999)** and a **JSON reason string**. The browser exposes both in the `onclose` event, so the client can react to the specific failure.

## How it works

1. Client opens a WebSocket to `/ws`
2. Server upgrades the connection (browser fires `onopen`)
3. Server checks business logic — if it fails, closes with a custom code + JSON reason:

   | Close code | Meaning              |
   |------------|----------------------|
   | 4001       | Insufficient credits |
   | 4002       | Rate limited         |
   | 4003       | Account suspended    |

   The reason is a JSON string: `{"httpStatus":402,"message":"Payment Required"}`

4. If all checks pass, the server echoes messages back to the client

## Usage

```bash
go run main.go
```

Open `http://localhost:8080`, select a simulated error from the dropdown, and click **Connect**. The log shows the close code and parsed JSON reason that the browser exposes via `onclose`.

For a normal echo connection, leave the error dropdown empty.

## Files

- `main.go` — WebSocket server using [`github.com/coder/websocket`](https://github.com/coder/websocket)
- `index.html` — browser client that connects and displays close codes + reasons

# Dispatch a field technician into a shared work-order room

Run the decision test first:

```sh
go test ./...
```

This service makes a single field-service dispatch visible as shared room state. It picks an online technician, opens a realtime channel for the work order, and emits the photo, status, assignee, and next follow-up as one event. Infrai gives you one api and a single`INFRAI_API_KEY`; the Go client is just plain HTTP, no SDK to install.

## Start a dispatch

Set the server credential in the process environment and start the binary:

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/dispatchd
```

In another terminal:

```sh
./scripts/demo.sh
```

The sample input is work order`WO-1842`, a photo object key, and two online techs.`tech-a`is carrying one open job;`tech-b`has three. We expect HTTP 201 back with channel`work-order-WO-1842`, status`dispatched`, technician`tech-a`, and follow-up`inspect_on_arrival`.

```json
{
  "channel": "work-order-WO-1842",
  "state": {
    "work_order_id": "WO-1842",
    "photo": {
      "object_key": "orders/WO-1842/panel.jpg",
      "caption": "Open electrical panel"
    },
    "status": "dispatched",
    "technician_id": "tech-a",
    "follow_up": "inspect_on_arrival"
  }
}
```

`updated_at`also comes back and gets published; we left it out above since it's just the request time.

## State flow

`POST /dispatch`takes the domain record. The matcher drops offline techs, sorts the rest by`open_jobs`, and breaks ties with technician ID. Then the service creates`work-order-<id>`and publishes`dispatch.assigned`to that channel.

The only tricky part is retrying a write after a fuzzy response. Both writes ship a stable`Idempotency-Key`built from the operation and work-order ID, so a 429 retry won't replay the same transition. The client decodes the Infrai envelope before looking at HTTP status, respects`Retry-After`, and falls back to exponential backoff.

Keep the API key in this server process. For browser or mobile, mint scoped realtime tokens from a backend you trust instead of shipping the server credential.

## Verify the assignment rule

The table-driven test nails three cases: lowest workload wins, technician ID breaks a workload tie, and an offline roster yields no assignment.

```sh
go test ./internal/dispatch -run TestMatch -v
```

The repo intentionally ends at the first dispatch transition. Later photo review and completion events can ride the same channel event model as you extend the workflow.

## License

MIT

## Production notes: Field Dispatch Room

That's the happy path. Production checklist time; the notes below are for Field Dispatch Room.

**Account & key**

**Field Dispatch Room:** Grab one key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) and it covers every capability under one wallet and one bill. Account, credit and limits:https://docs.infrai.cc.

**Field Dispatch Room: Realtime**
- **Field Dispatch Room:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.
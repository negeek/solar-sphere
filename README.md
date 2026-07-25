# Solar-Sphere

Solar-Sphere is a self-hostable API for collecting and retrieving readings
from low-cost solar-irradiance sensors.

## The Device

![Solar meter device](https://github.com/negeek/solar-sphere/blob/main/solarmeterproject.png)

Solar-Sphere is the backend for a low-cost solar-irradiance meter built
around five photodiodes whose outputs are averaged to reach an accuracy
comparable to much more expensive thermopile-based pyranometers, at roughly
8% of the cost. The design and validation are written up in:

> Adebisi J. I., Adebowale D. A., Adegoke O. M., "Development of a Low-Cost
> Irradiance Meter with Remote Data Logger," *Advanced Engineering Forum*,
> Vol. 54, pp. 91–103, 2025. https://www.scientific.net/AEF.54.91

The paper's own prototype streamed readings to ThingSpeak over Wi-Fi. This
repository is a self-hosted alternative to that: your device publishes
readings over MQTT, and this API stores and serves them back to you — no
third-party IoT platform involved, and no dependency on a research-lab
Arduino serial monitor either, which is what originally motivated writing
this codebase.

## Architecture

Four independent Go modules, each deployable on its own:

```
                     ┌──────────────┐
   your device  ───▶ │  MQTT broker │
  (publishes readings)└─────┬────────┘
                             │ subscribes
                             ▼
┌──────────┐   HTTP    ┌─────────────┐        ┌──────────────┐
│  client  │◀─────────▶│ solar-galaxy│◀──────▶│  solar-auth  │
│(you /    │  :8080    │  (gateway)  │  :3000 │ (sign-up,    │
│ your app)│           └──────┬──────┘        │  access keys)│
└──────────┘                  │                └──────┬───────┘
                               │ :5000                 │
                               ▼                        │
                        ┌──────────────┐                │
                        │solar-sentinel│                │
                        │ (devices +   │                │
                        │  readings)   │                │
                        └──────┬───────┘                │
                               │                         │
                               ▼                         ▼
                          ┌─────────────────────────────────┐
                          │             MongoDB              │
                          └───────────────────────────────────┘

solar-spectrum: shared library (types, env/log/http/JWT/Mongo helpers,
                migration runner) imported by the three services above —
                it isn't a running service.
```

- **solar-auth** — sign-up and access-key issuance/rotation.
- **solar-sentinel** — MQTT ingestion of readings, device registration, CSV
  export, all gated by access-key authentication.
- **solar-galaxy** — a hand-rolled API gateway/reverse proxy in front of the
  two services above, so clients only need to talk to one host/port.
- **solar-spectrum** — the shared library the other three import (not a
  service you run).

## Self-hosting (Docker)

Requires Docker and Docker Compose.

```bash
cp .env.example .env
go run ./solar-spectrum/cmd/keygen   # prints SIGNING_KEY and VERIFICATION_KEY — paste both into .env
docker compose up --build
```

This starts MongoDB, an MQTT broker (Mosquitto), and all three services,
applying database migrations automatically before each service starts.

| Service        | Port |
|----------------|------|
| solar-galaxy   | 8080 |
| solar-auth     | 3000 |
| solar-sentinel | 5000 |
| MQTT (mosquitto) | 1883 |

Everything below talks to the gateway on `:8080` — that's the one port you
need exposed to the outside world; `auth`/`sentinel` are only reachable
through it unless you choose to publish their ports too.

### 1. Sign up

```bash
curl -X POST 'http://localhost:8080/auth/v1/join/' \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com"}'
```

Returns an `access_key` — a bearer token, used on every authenticated
request below. By default it never expires; pass `"expires_in_days": 30` in
the request body if you'd rather it did. Either way you can invalidate it at
any time via `/auth/v1/new_key/`, which revokes the old key and issues a new
one (same optional `expires_in_days` field):

```bash
curl -X POST 'http://localhost:8080/auth/v1/new_key/' \
  -H 'Content-Type: application/json' \
  -d '{"key":"<your current access_key>","email":"you@example.com"}'
```

### 2. Register a device

A user can register as many devices as they like — there's no admin step.

```bash
curl -X POST 'http://localhost:8080/sentinel/v1/device/' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <your access_key>' \
  -d '{"name":"backyard sensor"}'
```

Returns the new `device_id`. The device is owned by whoever's access key
created it — only that account can later download its data.

### 3. Publish readings from the device

Point your device's MQTT client at the broker and publish JSON readings to:

```
Broker:  tcp://localhost:1883
Topic:   solar-sphere/solar-sentinel/sensor/<device_id>/solar-irradiance
```

### 4. Download your data

```bash
curl -X GET 'http://localhost:8080/sentinel/v1/download/<device_id>' \
  -H 'Authorization: Bearer <your access_key>'
```

Streams a CSV of that device's readings. Only the device's owner can
download it — anyone else's key gets a 403.

## Local development (without Docker)

Each service is its own Go module; `go.work` at the repo root ties them
together so they resolve `solar-spectrum` locally without needing it
published anywhere.

```bash
# Infra only, via Docker — or point at your own local Mongo/Mosquitto instead.
docker compose up mongo mosquitto

# Per service (repeat for solar-sentinel, solar-galaxy):
cd solar-auth
cp .env.example .env   # then fill in SIGNING_KEY/VERIFICATION_KEY
go run ./db/v1          # apply migrations
go run .                 # start the service
```

Run the test suite from the repo root:

```bash
go build ./... && go vet ./...
go test ./...
```

Most tests (service logic, handlers, the gateway) run with no setup. The
repository-layer integration tests talk to a real MongoDB and are skipped
unless `TEST_DATABASE_URL` is set:

```bash
TEST_DATABASE_URL=mongodb://localhost:27017 go test ./...
```

Each of those tests uses its own throwaway database and drops it in
cleanup.

## Notes on the security model

- Access keys are bearer tokens (EdDSA-signed JWTs) with **opt-in** expiry —
  set `expires_in_days` at sign-up/rotation if you want one, otherwise the
  key is valid until you explicitly revoke it via `/auth/v1/new_key/`.
- The Mosquitto config shipped in `mosquitto/mosquitto.conf` allows
  anonymous publish/subscribe, which is fine for a private/personal
  deployment. If you expose the broker beyond your own network, add a
  password file and TLS first.
- MQTT ingestion isn't authenticated at the application layer (there's no
  per-device broker ACL wired up) — it does check that the device ID in the
  topic was actually registered before storing a reading under it, but
  anyone who can reach your broker can still publish to a known topic. Keep
  the broker on a network you trust, or add MQTT-level ACLs if you don't.

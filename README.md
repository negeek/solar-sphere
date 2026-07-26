# Solar-Sphere

## Why I built this

A few years ago, in my final year, a colleague and I built a low-cost
solar-irradiance meter: five photodiodes averaged together to reach an
accuracy comparable to much more expensive thermopile-based pyranometers,
at roughly 8% of the cost. The design and validation are written up in:

> Adebisi J. I., Adebowale D. A., Adegoke O. M., "Development of a Low-Cost
> Irradiance Meter with Remote Data Logger," *Advanced Engineering Forum*,
> Vol. 54, pp. 91–103, 2025. https://www.scientific.net/AEF.54.91

![Solar meter device](https://github.com/negeek/solar-sphere/blob/main/solarmeterproject.png)

Once the meter itself worked, I didn't want to be stuck reading numbers off
an Arduino serial monitor — I wanted the readings somewhere I could get to
in real time, store, and export for analysis. That's what this codebase
was originally for. Since then I've spent time growing as a software
engineer, and I've come back to rebuild this the way I'd actually build a
backend today: a proper service layer instead of logic-in-handlers,
real tests, structured logging, and something you can self-host with one
command instead of a research-project script.

## What it actually does

Strip away the "solar" branding and this is a small, generic pattern:

**device → publishes JSON over MQTT → gets authenticated, stored, and made
downloadable as CSV, scoped to whoever owns that device.**

A reading is stored as `{device_id, data: {...arbitrary JSON...}, timestamps}`
— the `data` field isn't tied to solar-irradiance fields at all. Any device
that can publish JSON to an MQTT topic can use this exactly as-is: a
temperature logger, a soil-moisture sensor, a home-brew air-quality
monitor, whatever you've got that just needs "publish readings under my own
device ID, authenticate as the owner, pull them back out as CSV later."
The service/collection names (`solar-sentinel`, `solar-irradiance`) still
reflect what it was originally built for — I haven't renamed them — but
nothing about the data model or auth model assumes solar irradiance
specifically.

## Design principles

- **Lean on dependencies.** Outside of the Mongo driver, the MQTT client,
  and a router, this avoids reaching for libraries where a small amount of
  standard-library code does the job — env-file loading and DB migrations
  are both hand-rolled rather than pulled in as dependencies.
- **Memory-efficient, small footprint.** Each service compiles to a static
  binary running in a distroless image (no shell, no package manager, just
  the binary) — the whole stack (3 services + MongoDB + an MQTT broker) is
  light enough to self-host on a small VPS or something like a Raspberry
  Pi, not something that needs a cluster.
- **A handful of independent services, not a monolith.** Each piece
  (`solar-auth`, `solar-sentinel`, `solar-galaxy`) is its own Go module you
  could run, redeploy, or scale on its own.

## Architecture

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

## Self-hosting

Requires Docker and Docker Compose. Grab `docker-compose.sample.yml`,
`mosquitto/mosquitto.conf`, and `.env.example` from this repo (cloning it
is the easiest way):

```bash
git clone https://github.com/negeek/solar-sphere.git && cd solar-sphere
cp .env.example .env
go run ./solar-spectrum/cmd/keygen   # prints SIGNING_KEY and VERIFICATION_KEY — paste both into .env
docker compose -f docker-compose.sample.yml up
```

This pulls the published images (no local build step), starts MongoDB, an
MQTT broker (Mosquitto), and all three services, and applies database
migrations automatically before each service starts.

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

The JSON body can be whatever key/value readings your device produces —
it's stored as-is.

### 4. Download your data

```bash
curl -X GET 'http://localhost:8080/sentinel/v1/download/<device_id>' \
  -H 'Authorization: Bearer <your access_key>'
```

Streams a CSV of that device's readings. Only the device's owner can
download it — anyone else's key gets a 403.

## Local development

Each service is its own Go module; `go.work` at the repo root ties them
together so they resolve `solar-spectrum` locally without needing it
published anywhere. `docker-compose.yml` (as opposed to the sample compose
file above) builds all three images from source, if you'd rather do that
than run them with `go run`.

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

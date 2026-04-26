# mini-brimble

A self-hosted deployment platform that takes a Git repository, builds it into a container image using Railpack, runs it with Docker, and routes traffic to it through Caddy — all driven from a single UI.

---

## What this is

This is a scoped-down PaaS pipeline. You push a Git URL, the system clones it, builds a container image without a Dockerfile, starts the container, and registers a live URL — all visible in real time through a log stream in the browser.

---

## Architecture

```
Browser
↓
Caddy :80  (single point of ingress)
├── /api/*       → backend:8080
├── /deploys/*   → user containers (routes added dynamically at runtime)
└── /*           → frontend:80 (nginx, static files)

Backend (Go)
├── SQLite       → deployment state + log persistence
├── Railpack     → builds container images via BuildKit
├── Docker CLI   → starts/stops user containers
└── Caddy API    → registers routes at runtime via Admin API :2019
```

### How a deployment works

1. User submits a Git URL and a name
2. Backend creates a deployment record (`status: pending`) and returns immediately
3. A background goroutine runs the pipeline:
   - Clones the repo into a temp directory
   - Updates status → `building`, runs Railpack via BuildKit
   - Updates status → `deploying`, starts the container on the `brimble` Docker network
   - Calls Caddy's Admin API to register `/deploys/<id>/*` → container IP
   - Updates status → `running`, stores the live URL
4. Every log line is written to SQLite and broadcast to any open SSE connections in real time
5. If any step fails, status → `failed` with the error message in the log stream

### Key design decisions

**Caddy as single ingress**
All traffic flows through Caddy on port 80. User container ports are never exposed publicly. Caddy reaches containers via their internal Docker network IP — not via `host.docker.internal` which doesn't resolve on Linux.

**Dynamic routing via Caddy Admin API**
Routes are added at runtime by POST-ing to Caddy's Admin API on port 2019. No Caddyfile edits, no restarts. Each deployment inserts its route at index 0 so it matches before the frontend catch-all `/*`.

**SSE over WebSocket for log streaming**
Logs are unidirectional — server pushes, browser listens. SSE is simpler than WebSocket for this pattern. Each log line is written to SQLite first (so replays work on reconnect), then broadcast via an in-memory pub/sub broadcaster to all open SSE connections for that deployment.

**SQLite with WAL mode**
WAL (Write-Ahead Logging) allows concurrent reads during writes. Without it, the SSE reader would block whenever the build goroutine writes a log line. `SetMaxOpenConns(1)` prevents "database is locked" errors since SQLite allows only one writer at a time.

**Docker-out-of-Docker**
The backend container has the Docker CLI installed and the host Docker socket mounted at `/var/run/docker.sock`. It talks to the host Docker daemon directly — no Docker daemon inside the container. This is lighter and safer than Docker-in-Docker.

**Railpack via BuildKit**
Railpack detects the runtime (Node, Go, Python) from source code and builds a container image without a Dockerfile. It uses BuildKit as its build engine. BuildKit runs as a privileged container in docker-compose and is referenced via `BUILDKIT_HOST`.

**Async pipeline**
The `POST /deployments` endpoint returns 201 immediately after creating the DB record. The build pipeline runs in a background goroutine. The frontend tracks progress via SSE — it never polls.

**Layered backend structure**

```
handlers/    → HTTP only, no SQL
pipeline/    → infrastructure only, no HTTP
db/          → SQL only, no business logic
models/      → shared types
config/      → environment config with sensible defaults
```

---

## Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.22, stdlib HTTP router, modernc/sqlite |
| Frontend | Vite, React, TanStack Router, TanStack Query |
| Build | Railpack + BuildKit |
| Runtime | Docker |
| Ingress | Caddy with dynamic Admin API routing |
| Database | SQLite (WAL mode) |

---

## Prerequisites

- Docker and Docker Compose installed on the host
- Port 80 free (Caddy)
- Port 8080 free (backend)

---

## Running it

```bash
git clone https://github.com/favxlaw/mini-brimble
cd mini-brimble
docker compose up --build
```

Open `http://localhost` in your browser.

---

## Testing the pipeline

A sample Node.js app is available for testing the full pipeline:
https://github.com/favxlaw/sample-app_brimble

Paste this URL into the UI, give it a name, and hit Deploy. The build takes 2–5 minutes on first run (Railpack pulls base images). Subsequent builds are faster due to BuildKit layer caching.

Once the status shows `running`, click the live URL to see the deployed app.

---

## Environment variables

All variables have sensible defaults and work out of the box with `docker compose up`.

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Backend server port |
| `DB_PATH` | `/data/brimble.db` | SQLite database path |
| `CADDY_ADMIN_URL` | `http://caddy:2019` | Caddy Admin API endpoint |
| `PUBLIC_URL` | `http://localhost` | Base URL for live deployment links |
| `CONTAINER_PORT` | `3000` | Port user containers listen on |
| `DOCKER_NETWORK` | `mini-brimble_brimble` | Docker network for user containers |
| `BUILDKIT_HOST` | `docker-container://buildkit` | BuildKit daemon address |

---

## API

```
POST   /deployments              create a deployment, start pipeline
GET    /deployments              list all deployments
GET    /deployments/{id}         get one deployment
GET    /deployments/{id}/logs    SSE stream of build/deploy logs
DELETE /deployments/{id}         stop a running deployment
```

---

## What I'd do with more time

**Reconciliation loop**
Currently routing state lives in two places — SQLite says `running`, Caddy holds the route in memory. If Caddy restarts, routes are gone but the DB still shows running deployments. The fix is a reconciler: a goroutine that runs every 30 seconds, fetches Caddy's actual route table, diffs it against the DB, and re-registers missing routes. This is the control plane pattern — desired state (SQLite) reconciled against observed state (Caddy).

**Docker SDK instead of CLI**
We use `exec.Command("docker", ...)` because the Docker SDK requires Go 1.24+ and we're on 1.22. The SDK would give us typed return values instead of string parsing, and a smaller final image (no docker.io package needed — just SDK code compiled into the binary).

**Build cache across deploys**
Railpack supports BuildKit cache mounts. Right now each build starts cold. Persisting the BuildKit cache between builds would cut build times significantly for repos with heavy dependencies.

**Graceful zero-downtime redeploys**
Current flow: stop old container, start new one — there's a gap. The fix is blue/green: start the new container, wait for it to pass a health check, update the Caddy route, then stop the old container.

**Upload support**
The UI had a zip upload option that we removed to keep scope tight. The backend pipeline already handles `SourceTypeUpload` — it just needs a multipart handler to receive the file and extract it to the temp directory.

---

## What I'd rip out

The `host_port` column in the deployments table. We store it but Caddy now reaches containers via their internal IP — the host port is never used for routing. It's leftover from an earlier approach where we bound ports to the host. Worth cleaning up to avoid confusion.

---

## Time spent

Approximately 2 days. First day on backend architecture, pipeline, and Docker/Caddy integration. Second day debugging Linux-specific networking issues (`host.docker.internal` not resolving, BuildKit internet access inside custom Docker networks) and frontend wiring.

---

## Brimble deploy + feedback

*[to be added after deploying on Brimble]*
# Elysium Architecture

This document describes the technical architecture of Elysium — the API app store consisting of a **CLI tool** (`ely`), a **Registry Server**, and a **Supabase-backed database**.

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [CLI Architecture](#2-cli-architecture)
3. [Server Architecture](#3-server-architecture)
4. [Database Schema](#4-database-schema)
5. [Authentication Flows](#5-authentication-flows)
6. [Key Design Decisions](#6-key-design-decisions)
7. [Directory Structure](#7-directory-structure)

---

## 1. System Overview

Elysium has three main components:

| Component | Technology | Responsibility |
|-----------|-----------|----------------|
| **CLI** (`ely`) | Go + Cobra | User-facing tool: login, search, pull, execute emblems |
| **Registry Server** | Python + FastAPI | REST API for managing and serving emblems |
| **Database** | Supabase (PostgreSQL) | Persistent storage for users, emblems, keys, and device codes |

### System Overview Diagram

```mermaid
flowchart TD
    User["👤 Developer"]

    subgraph CLI["CLI Tool (ely)"]
        Commands["Cobra Commands\n(login, pull, search, execute…)"]
        EmblemParser["Emblem Parser & Validator"]
        Executor["HTTP Executor"]
        Config["Config / State\n(~/.elysium/)"]
    end

    subgraph Server["Registry Server (FastAPI)"]
        Routes["API Routes\n(/api/auth, /api/emblems, /api/keys)"]
        Services["Service Layer\n(AuthService, EmblemService, KeyService)"]
        Middleware["Middleware\n(CORS, Rate Limit, Security Headers)"]
    end

    subgraph DB["Supabase (PostgreSQL)"]
        AuthSchema["auth.users"]
        PublicSchema["public schema\n(profiles, emblems,\nemblem_versions, device_codes,\napi_keys)"]
    end

    subgraph ExtAPI["Third-party API"]
        TargetAPI["Target REST API\n(e.g. clothing-shop)"]
    end

    User -->|"runs ely commands"| Commands
    Commands --> EmblemParser
    Commands --> Config
    EmblemParser --> Executor
    Executor -->|"HTTP request"| TargetAPI
    Commands -->|"HTTPS REST calls"| Middleware
    Middleware --> Routes
    Routes --> Services
    Services -->|"Supabase client"| DB
    DB --> AuthSchema
    DB --> PublicSchema
```

### Data Flow Summary

1. The **developer** runs `ely` commands.
2. The CLI authenticates with the Registry Server and stores tokens locally (`~/.elysium/config.yaml`).
3. To discover or retrieve emblems, the CLI calls the Registry's REST API.
4. To execute an emblem action, the CLI parses the local YAML, builds an HTTP request, and calls the **third-party API** directly — no traffic passes through the Registry at execution time.

---

## 2. CLI Architecture

The CLI is built with [Cobra](https://github.com/spf13/cobra) and organised into commands and internal packages.

### Command Structure

```mermaid
flowchart TD
    Root["ely (root)"]

    Root --> login["login\n(password / device / OAuth)"]
    Root --> logout["logout"]
    Root --> whoami["whoami"]

    Root --> pull["pull <name>[@version]"]
    Root --> update["update &lt;name&gt;"]
    Root --> list["list"]
    Root --> outdated["outdated"]

    Root --> search["search &lt;query&gt;"]
    Root --> info["info &lt;name&gt;"]

    Root --> execute["execute &lt;name&gt; &lt;action&gt;"]
    Root --> validate["validate &lt;file&gt;"]
    Root --> init["init"]

    Root --> keys["keys (list/create/delete)"]
    Root --> config["config"]
    Root --> self_update["self-update"]
    Root --> check_updates["check-updates"]
    Root --> completion["completion"]
```

### Package Organisation

```mermaid
flowchart LR
    subgraph cmd["cli/cmd/"]
        CmdFiles["*.go — one file per command\n(login.go, pull.go, search.go…)"]
    end

    subgraph internal["cli/internal/"]
        API["api/\nclient.go — Registry HTTP client"]
        Config["config/\nconfig.go — read/write\n~/.elysium/config.yaml"]
        Emblem["emblem/\nparser.go — YAML loader & validator"]
        Executor["executor/\nrunner.go — HTTP request executor"]
        Validator["validator/\nvalidator.go — schema validation"]
        ErrFmt["errfmt/\nstructured error formatting"]
        HTTPClient["httpclient/\nshared http.Client with timeout"]
        Scaffold["scaffold/\nemblem init scaffolding"]
        SelfUpdate["selfupdate/\nbinary self-update logic"]
    end

    CmdFiles --> API
    CmdFiles --> Config
    CmdFiles --> Emblem
    CmdFiles --> Executor
    Emblem --> Validator
    Executor --> HTTPClient
    API --> HTTPClient
```

### CLI Data Flow for `ely execute`

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Cache
    participant API as Third-Party API

    User->>CLI: ely stripe create-customer --email a@b.com
    CLI->>Cache: Load stripe/emblem.yaml
    Cache-->>CLI: Emblem definition
    CLI->>CLI: Match action "create-customer"
    CLI->>CLI: Build HTTP request from action + flags
    CLI->>CLI: Inject auth (API key / Bearer token)
    CLI->>API: POST /customers {"email": "a@b.com"}
    API-->>CLI: 200 {"id": "cus_xyz"}
    CLI->>User: Pretty-print response
```

### Local State

Local state is stored in `~/.elysium/`:

```
~/.elysium/
├── config.yaml        # Auth token, server URL, user info
└── cache/
    ├── clothing-shop/
    │   └── emblem.yaml
    └── stripe/
        └── emblem.yaml
```

---

## 3. Server Architecture

The server is a **FastAPI** application deployed on Vercel (serverless) or as a standard ASGI server via Uvicorn/Gunicorn.

### Request Lifecycle

```mermaid
flowchart TD
    Client["Client\n(CLI / Browser)"]

    subgraph FastAPI["FastAPI Application"]
        SecHdr["SecurityHeadersMiddleware\n(X-Frame-Options, nosniff…)\nadded first → innermost"]
        CORS["CORSMiddleware\n(origin allowlist)\nadded second → outermost"]
        Router["APIRouter\n/api/auth  /api/emblems  /api/keys"]
        AuthDep["get_current_user() dependency\n(Bearer JWT → Supabase validation)"]
        RateLimitNote["Rate limiting applied\nper-route via @limiter.limit()"]

        subgraph Services["Service Layer"]
            AuthSvc["AuthService\nauth_service.py"]
            EmblemSvc["EmblemService\nemblem_service.py"]
            KeySvc["KeyService\nkey_service.py"]
        end
    end

    Supabase["Supabase Client\n(run_sync → asyncio.to_thread)"]

    Client -->|"HTTPS request"| SecHdr
    SecHdr --> CORS
    CORS --> Router
    Router -->|"rate-limited endpoints"| RateLimitNote
    RateLimitNote --> AuthDep
    Router -->|"protected routes"| AuthDep
    AuthDep --> Services
    Router -->|"public routes"| Services
    Services --> Supabase
```

### Route Map

| Prefix | File | Key Endpoints |
|--------|------|---------------|
| `/api/auth` | `routes/auth.py` | `POST /register`, `POST /login`, `POST /logout`, `POST /refresh`, `GET /me`, `PATCH /profile`, `GET /oauth/{provider}/start`, `POST /device/code`, `POST /device/verify`, `POST /device/token` |
| `/api/emblems` | `routes/emblems.py` | `GET /`, `POST /`, `GET /{name}`, `GET /{name}/versions`, `GET /{name}/{version}`, `PUT /{name}`, `DELETE /{name}` |
| `/api/keys` | `routes/keys.py` | `GET /`, `POST /`, `DELETE /{id}` |

### Key Files

| File | Purpose |
|------|---------|
| `app/main.py` | FastAPI app factory, middleware registration |
| `app/config.py` | Environment-based settings via Pydantic |
| `app/models.py` | Pydantic request/response models |
| `app/database.py` | Supabase client singleton and `run_sync()` helper |
| `app/limiter.py` | Rate-limit configuration (slowapi) |
| `app/services/auth_service.py` | Registration, login, OAuth, device-code logic |
| `app/services/emblem_service.py` | CRUD and search for emblems |
| `app/services/key_service.py` | API key lifecycle management |

### Pull & Publish Flows

```mermaid
sequenceDiagram
    participant Developer
    participant CLI
    participant Registry
    participant Supabase

    Note over Developer,Supabase: Pull flow
    CLI->>Registry: GET /api/emblems/stripe
    Registry->>Supabase: Query emblems + emblem_versions
    Supabase-->>Registry: Emblem record
    Registry-->>CLI: 200 {yaml_content, ...}
    CLI->>CLI: Write ~/.elysium/cache/stripe/emblem.yaml

    Note over Developer,Supabase: Publish flow
    Developer->>CLI: ely validate ./stripe/
    CLI->>CLI: Validate against emblem.schema.json
    CLI-->>Developer: ✓ Valid
    Developer->>CLI: ely publish ./stripe/
    CLI->>Registry: POST /api/emblems (Bearer token)
    Registry->>Supabase: Insert emblem + emblem_version
    Supabase-->>Registry: 201 Created
    Registry-->>CLI: 201 Created
    CLI-->>Developer: ✓ Published stripe@1.0.0
```

---

## 4. Database Schema

The database is hosted on **Supabase** (PostgreSQL). Row Level Security (RLS) is enabled on all public tables. `auth.users` is managed entirely by Supabase (passwords, JWT issuance) — application code only references it via foreign keys. The `public.profiles` table holds application-level user data and is linked 1-to-1 with `auth.users`.

> **Migration status:** `profiles` is created by `001_profiles_and_search.sql` and `device_codes` by `002_device_codes.sql`. The `emblems`, `emblem_versions`, and `api_keys` tables are referenced in grants and foreign-key constraints within those migrations but their own `CREATE TABLE` DDL is not yet in the migrations directory (pending migration). The schema below reflects the intended complete structure used by the application code.

```mermaid
erDiagram
    USERS["auth.users (Supabase-managed)"] {
        uuid id PK
        text email
        timestamptz created_at
    }

    PROFILES {
        uuid id PK
        text username UK
        text email
        text avatar_url
        text bio
        timestamptz created_at
        timestamptz updated_at
    }

    EMBLEMS {
        uuid id PK
        text name UK
        text description
        uuid author_id FK
        text author_name
        text category
        text_array tags
        text license
        text repository_url
        text homepage_url
        text latest_version
        int downloads_count
        tsvector search_vector
        text security_advisory
        text security_severity
        timestamptz created_at
        timestamptz updated_at
    }

    EMBLEM_VERSIONS {
        uuid id PK
        uuid emblem_id FK
        text version
        text yaml_content
        text changelog
        uuid published_by FK
        timestamptz published_at
    }

    API_KEYS {
        uuid id PK
        uuid user_id FK
        text name
        text key_hash
        timestamptz created_at
        timestamptz expires_at
    }

    DEVICE_CODES {
        uuid id PK
        text device_code UK
        text user_code UK
        uuid user_id FK
        text access_token
        text refresh_token
        text client_name
        timestamptz verified_at
        timestamptz expires_at
        timestamptz created_at
    }

    USERS ||--|| PROFILES : "has one"
    PROFILES ||--o{ EMBLEMS : "authors"
    EMBLEMS ||--o{ EMBLEM_VERSIONS : "has many"
    PROFILES ||--o{ API_KEYS : "owns"
    USERS ||--o{ DEVICE_CODES : "authenticates via"
    PROFILES ||--o{ EMBLEM_VERSIONS : "published_by"
```

### Key Indexes

| Table | Index | Type | Purpose |
|-------|-------|------|---------|
| `profiles` | `idx_profiles_username` | B-tree | Fast username lookup |
| `profiles` | `idx_profiles_id` | B-tree | Join optimisation |
| `emblems` | `idx_emblems_search` | GIN | Full-text search (`tsvector`) |
| `emblems` | `idx_emblems_author_id` | B-tree | Author queries |
| `device_codes` | `idx_device_codes_device_code` | B-tree | CLI polling |
| `device_codes` | `idx_device_codes_user_code` | B-tree | Browser verification |
| `device_codes` | `idx_device_codes_expires_at` | B-tree | Cleanup queries |

### Full-Text Search

The `search_vector` column on `emblems` is maintained by a PostgreSQL trigger (`emblems_search_trigger`) that combines the emblem `name` (weight A), `description` (weight B), and `tags` (weight C) into a `tsvector`. Search queries use the `search_emblems_fts` RPC function.

---

## 5. Authentication Flows

Elysium supports three authentication mechanisms.

### 5a. Email / Password Flow

```mermaid
sequenceDiagram
    actor User
    participant CLI as ely CLI
    participant Server as Registry Server
    participant Supabase

    User->>CLI: ely login --password
    CLI->>User: Prompt for email & password
    User->>CLI: Enters credentials
    CLI->>Server: POST /api/auth/login {email, password}
    Server->>Supabase: supabase.auth.sign_in_with_password()
    Supabase-->>Server: access_token + refresh_token
    Server-->>CLI: AuthResponse {access_token, refresh_token, user}
    CLI->>CLI: Save tokens to ~/.elysium/config.yaml
    CLI-->>User: "Logged in as username"
```

### 5b. Device Code Flow (browser-based CLI login)

```mermaid
sequenceDiagram
    actor User
    participant CLI as ely CLI
    participant Server as Registry Server
    participant Browser
    participant Supabase

    User->>CLI: ely login (default)
    CLI->>Server: POST /api/auth/device/code
    Server->>Supabase: Insert device_codes record
    Server-->>CLI: {device_code, user_code, verification_uri}
    CLI-->>User: "Open https://... and enter code XXXX-YYYY"

    User->>Browser: Opens verification_uri
    Browser->>Server: GET /api/auth/device/status?user_code=XXXX-YYYY
    Server-->>Browser: {verified: false}

    User->>Browser: Logs in via email/password or OAuth
    Browser->>Server: POST /api/auth/device/verify {user_code}
    Server->>Supabase: Mark device_code verified, store tokens
    Server-->>Browser: 200 OK

    loop Poll every 5 s (until verified or expired)
        CLI->>Server: POST /api/auth/device/token {device_code}
        Server-->>CLI: 202 Pending
    end

    CLI->>Server: POST /api/auth/device/token {device_code}
    Server-->>CLI: DeviceTokenResponse {access_token, refresh_token, user}
    CLI->>CLI: Save tokens to ~/.elysium/config.yaml
    CLI-->>User: "Logged in as username"
```

### 5c. OAuth Flow (GitHub / Google)

```mermaid
sequenceDiagram
    actor User
    participant CLI as ely CLI
    participant Server as Registry Server
    participant Supabase
    participant OAuthProvider as GitHub or Google

    User->>CLI: ely login --oauth github
    CLI->>Server: GET /api/auth/oauth/github/start?redirect_uri=...
    Server->>Supabase: Generate OAuth URL + CSRF state
    Server-->>CLI: {url: "https://github.com/login/oauth/authorize?..."}
    CLI-->>User: "Open URL in browser"

    User->>OAuthProvider: Authorises application
    OAuthProvider->>Server: GET /api/auth/oauth/github/callback?code=...&state=...
    Server->>Supabase: Exchange code for session
    Supabase-->>Server: access_token + refresh_token
    Server-->>User: Redirect to frontend with tokens
    CLI->>CLI: Save tokens to ~/.elysium/config.yaml
    CLI-->>User: "Logged in as username"
```

### Token Management

```mermaid
flowchart TD
    A["Token stored in\n~/.elysium/config.yaml"] -->|"attached to every request"| B["Authorization: Bearer token"]
    B --> C{Valid?}
    C -->|"Yes"| D["Request proceeds"]
    C -->|"401 Unauthorized"| E["CLI calls\nPOST /api/auth/refresh"]
    E --> F{Refresh OK?}
    F -->|"Yes"| G["New access_token saved\nRequest retried"]
    F -->|"No"| H["User prompted to\nrun ely login again"]
```

---

## 6. Key Design Decisions

### Why Go for the CLI?
- Compiles to a single static binary — no runtime required.
- Starts in under 100 ms — important for tight feedback loops.
- Excellent cross-platform support (Linux, macOS, Windows, ARM).
- Strong stdlib for HTTP, file I/O, and concurrency.

### Why FastAPI for the Registry?
- Auto-generated OpenAPI docs at `/docs`.
- Async endpoints allow future real-time features (emblem change notifications).
- Pydantic gives strict validation at the request boundary.
- Python's ecosystem makes it easy to add ML-based search ranking later.

### Why Supabase?
- Managed PostgreSQL removes operational burden.
- Built-in auth with JWT and Row Level Security.
- Free tier is sufficient for the current scale.
- Easy to self-host with the open-source Supabase stack if needed.

### Why YAML for Emblems?
- Human-readable and writeable — authors edit these by hand.
- Comments are supported (unlike JSON).
- Widely used in DevOps tooling (Kubernetes, GitHub Actions) — familiar to developers.
- JSON Schema validation can be applied to the parsed representation.

### Security Architecture

Key security properties:

- JWT tokens are stored in `~/.elysium/config.yaml` (file permissions: user-only).
- API credentials (e.g. `STRIPE_API_KEY`) are **never** stored by Elysium; they are read from environment variables at execution time.
- The executor validates all URLs before making requests (http/https only, no `file://` or internal addresses).
- The registry enforces authentication on write operations via Supabase Row Level Security.
- All services catch exceptions and return `500 Internal server error` without exposing internal details.
- Rate limiting is applied per-IP on all public endpoints.

---

## 7. Directory Structure

```
elysium/
├── cli/                    # Go CLI
│   ├── cmd/               # One file per Cobra command
│   ├── internal/
│   │   ├── api/          # Registry HTTP client
│   │   ├── config/       # ~/.elysium state
│   │   ├── emblem/       # YAML parser, validator, cache
│   │   ├── executor/     # HTTP request runner
│   │   ├── validator/    # Schema validation
│   │   ├── errfmt/       # Structured error formatting
│   │   ├── httpclient/   # Shared HTTP client
│   │   ├── scaffold/     # Emblem init scaffolding
│   │   └── selfupdate/   # Binary self-update logic
│   └── go.mod
│
├── server/                # FastAPI registry
│   ├── app/
│   │   ├── routes/       # auth.py, emblems.py, keys.py
│   │   ├── services/     # auth_service.py, emblem_service.py, key_service.py
│   │   ├── models.py     # Pydantic schemas
│   │   ├── database.py   # Supabase client + run_sync()
│   │   ├── config.py     # Settings
│   │   └── limiter.py    # Rate-limit config
│   ├── migrations/       # SQL migration scripts
│   └── tests/
│
├── schemas/
│   └── emblem.schema.json # JSON Schema — source of truth for emblems
│
├── examples/
│   └── clothing-shop/     # Example emblem + API
│
├── docs/
│   ├── ARCHITECTURE.md    # This file
│   ├── EMBLEM_SPEC.md     # Full emblem YAML specification
│   ├── GETTING_STARTED.md # User quick-start guide
│   └── SERVER_SETUP.md    # Deploying the registry server
│
└── scripts/
    ├── install.sh         # One-line installer
    └── build-all.sh       # Cross-platform binary builder
```

Full emblem specification: [EMBLEM_SPEC.md](EMBLEM_SPEC.md)

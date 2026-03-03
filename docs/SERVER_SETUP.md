# Server Setup Guide

This guide explains how to set up the Elysium registry server with Supabase.

## Prerequisites

- Python 3.11 or higher
- A [Supabase](https://supabase.com) account
- PostgreSQL (managed by Supabase)

## Architecture Overview

The Elysium registry consists of:

1. **Supabase Auth** - Handles user authentication
2. **Supabase PostgreSQL** - Stores emblem definitions and versions
3. **FastAPI Backend** - REST API for emblem management
4. **Go CLI** - Client tool for interacting with the registry

## Supabase Setup

### 1. Create a Supabase Project

1. Go to [supabase.com](https://supabase.com) and sign in
2. Click "New Project"
3. Name your project (e.g., "elysium-registry")
4. Set a secure database password
5. Choose a region close to your users
6. Click "Create new project"

### 2. Get API Keys

Once your project is created:

1. Go to Settings > API
2. Note down:
   - `Project URL` (e.g., `https://xxx.supabase.co`)
   - `anon public` key
   - `service_role` key (keep this secret!)

### 3. Create Database Schema

Go to SQL Editor and run:

```sql
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Profiles table (extends auth.users)
CREATE TABLE profiles (
    id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    username TEXT UNIQUE NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Emblems table
CREATE TABLE emblems (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    author_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    category TEXT,
    tags TEXT[],
    license TEXT DEFAULT 'MIT',
    repository_url TEXT,
    homepage_url TEXT,
    latest_version TEXT,
    downloads_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_name CHECK (name ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$'),
    CONSTRAINT valid_version CHECK (latest_version ~ '^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$')
);

-- Emblem versions table
CREATE TABLE emblem_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    emblem_id UUID REFERENCES emblems(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    yaml_content TEXT NOT NULL,
    changelog TEXT,
    published_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    published_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT unique_version UNIQUE (emblem_id, version),
    CONSTRAINT valid_semver CHECK (version ~ '^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$')
);

-- Pulls tracking (optional analytics)
CREATE TABLE emblem_pulls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    emblem_id UUID REFERENCES emblems(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    pulled_at TIMESTAMPTZ DEFAULT NOW(),
    pulled_by UUID REFERENCES auth.users(id) ON DELETE SET NULL
);

-- Create indexes
CREATE INDEX idx_emblems_name ON emblems(name);
CREATE INDEX idx_emblems_category ON emblems(category);
CREATE INDEX idx_emblems_author ON emblems(author_id);
CREATE INDEX idx_emblem_versions_emblem ON emblem_versions(emblem_id);
CREATE INDEX idx_emblem_pulls_emblem ON emblem_pulls(emblem_id);

-- Enable Row Level Security
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE emblems ENABLE ROW LEVEL SECURITY;
ALTER TABLE emblem_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE emblem_pulls ENABLE ROW LEVEL SECURITY;

-- Policies for profiles
CREATE POLICY "Public profiles are viewable by everyone"
    ON profiles FOR SELECT
    USING (true);

CREATE POLICY "Users can update own profile"
    ON profiles FOR UPDATE
    USING (auth.uid() = id);

-- Policies for emblems
CREATE POLICY "Emblems are viewable by everyone"
    ON emblems FOR SELECT
    USING (true);

CREATE POLICY "Authenticated users can create emblems"
    ON emblems FOR INSERT
    WITH CHECK (auth.uid() IS NOT NULL);

CREATE POLICY "Users can update own emblems"
    ON emblems FOR UPDATE
    USING (auth.uid() = author_id);

CREATE POLICY "Users can delete own emblems"
    ON emblems FOR DELETE
    USING (auth.uid() = author_id);

-- Policies for emblem_versions
CREATE POLICY "Emblem versions are viewable by everyone"
    ON emblem_versions FOR SELECT
    USING (true);

CREATE POLICY "Authenticated users can create versions"
    ON emblem_versions FOR INSERT
    WITH CHECK (auth.uid() IS NOT NULL);

-- Policies for emblem_pulls
CREATE POLICY "Authenticated users can insert pulls"
    ON emblem_pulls FOR INSERT
    WITH CHECK (auth.uid() IS NOT NULL);

-- Function to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers
CREATE TRIGGER update_profiles_updated_at
    BEFORE UPDATE ON profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_emblems_updated_at
    BEFORE UPDATE ON emblems
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Function to increment download count
CREATE OR REPLACE FUNCTION increment_download_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE emblems
    SET downloads_count = downloads_count + 1
    WHERE id = NEW.emblem_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER increment_download
    AFTER INSERT ON emblem_pulls
    FOR EACH ROW
    EXECUTE FUNCTION increment_download_count();
```

### 4. Set up Authentication

1. Go to Authentication > Providers
2. Enable Email provider (default)
3. Configure email templates for:
   - Confirmation email
   - Password reset
   - Magic link (optional)
4. (Optional) Enable OAuth providers:
   - Google
   - GitHub
   - GitLab

### 5. Configure Storage (Optional)

For storing emblem files (if needed):

1. Go to Storage
2. Create a bucket named `emblems`
3. Set public access policy:

```sql
-- Allow public read access
CREATE POLICY "Public read access"
    ON storage.objects FOR SELECT
    USING (bucket_id = 'emblems');
```

## Backend Setup

### 1. Clone and Install

```bash
cd elysium/server
python -m venv venv
source venv/bin/activate  # Linux/Mac
pip install -r requirements.txt
```

### 2. Environment Configuration

Create `.env`:

```env
# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_KEY=your-service-role-key

# App Configuration
APP_NAME=Elysium Registry
APP_VERSION=1.0.0
DEBUG=false

# Server Configuration
HOST=0.0.0.0
PORT=8000

# CORS
CORS_ORIGINS=http://localhost:3000,https://elysium.dev

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60
```

### 3. Run the Server

```bash
# Development
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

# Production
uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 4
```

### 4. Run with Docker

Create `Dockerfile`:

```dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

Build and run:

```bash
docker build -t elysium-server .
docker run -p 8000:8000 --env-file .env elysium-server
```

### 5. Deploy to Production

#### Option 1: Docker on VPS

```bash
# Build
docker build -t elysium-server .

# Run with environment variables
docker run -d \
  -p 8000:8000 \
  -e SUPABASE_URL=https://your-project.supabase.co \
  -e SUPABASE_ANON_KEY=your-anon-key \
  -e SUPABASE_SERVICE_KEY=your-service-key \
  --name elysium-server \
  elysium-server
```

#### Option 2: Railway

```yaml
# railway.toml
[build]
builder = "nixpacks"

[deploy]
startCommand = "uvicorn app.main:app --host 0.0.0.0 --port $PORT"
```

#### Option 3: Fly.io

```toml
# fly.toml
app = "elysium-registry"

[build]
  builder = "heroku/buildpacks:20"

[env]
  PORT = "8080"

[[services]]
  internal_port = 8080
  protocol = "tcp"

  [[services.ports]]
    handlers = ["http"]
    port = 80

  [[services.ports]]
    handlers = ["tls", "http"]
    port = 443
```

#### Option 4: Render

```yaml
# render.yaml
services:
  - type: web
    name: elysium-registry
    env: python
    buildCommand: pip install -r requirements.txt
    startCommand: uvicorn app.main:app --host 0.0.0.0 --port $PORT
    envVars:
      - key: SUPABASE_URL
        sync: false
      - key: SUPABASE_ANON_KEY
        sync: false
      - key: SUPABASE_SERVICE_KEY
        sync: false
```

## API Endpoints

### Authentication

```
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
POST /api/auth/refresh
GET  /api/auth/me
```

### Emblems

```
GET    /api/emblems                  # List all emblems
GET    /api/emblems/:name            # Get emblem metadata
GET    /api/emblems/:name/:version   # Download specific version
POST   /api/emblems                  # Publish new emblem
PUT    /api/emblems/:name            # Publish new version
DELETE /api/emblems/:name            # Delete emblem (owner only)
```

### Search

```
GET /api/search?q=query&category=ecommerce&sort=downloads
```

### User

```
GET /api/user/emblems    # List user's published emblems
GET /api/user/pulls      # List user's download history
```

## Security Considerations

### API Key Storage

Supabase service role key should never be exposed to clients:

```python
# Server-side only
SUPABASE_SERVICE_KEY = os.getenv('SUPABASE_SERVICE_KEY')
```

### Rate Limiting

Implement rate limiting to prevent abuse:

```python
from slowapi import Limiter
from slowapi.util import get_remote_address

limiter = Limiter(key_func=get_remote_address)

@app.route('/api/emblems')
@limiter.limit("100/minute")
def list_emblems():
    ...
```

### Input Validation

All inputs are validated using Pydantic:

```python
from pydantic import BaseModel, constr

class EmblemCreate(BaseModel):
    name: constr(regex=r'^[a-z0-9][a-z0-9-]*[a-z0-9]$', min_length=1, max_length=64)
    description: constr(min_length=10, max_length=500)
    version: constr(regex=r'^\d+\.\d+\.\d+')
    ...
```

### CORS

Configure allowed origins:

```python
from fastapi.middleware.cors import CORSMiddleware

app.add_middleware(
    CORSMiddleware,
    allow_origins=os.getenv('CORS_ORIGINS', '').split(','),
    allow_credentials=True,
    allow_methods=['*'],
    allow_headers=['*'],
)
```

## Monitoring

### Logging

```python
import logging

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
```

### Health Check

```
GET /health
```

Returns:
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

### Metrics (optional)

Use Prometheus:

```python
from prometheus_fastapi_instrumentator import Instrumentator

instrumentator = Instrumentator()
instrumentator.instrument(app).expose(app)
```

## Backup Strategy

### Database Backups

Supabase provides automatic daily backups. For additional backups:

```bash
# Using pg_dump
pg_dump $DATABASE_URL > backup.sql

# Restore
psql $DATABASE_URL < backup.sql
```

### Emblem Content

Since emblem YAML is stored in the database, regular database backups are sufficient.

## Troubleshooting

### Connection Issues

```bash
# Check Supabase status
curl https://your-project.supabase.co/rest/v1/

# Check logs
docker logs elysium-server
```

### Authentication Errors

1. Verify `SUPABASE_ANON_KEY` is correct
2. Check JWT expiration
3. Ensure user exists in Supabase Auth

### Database Errors

1. Check RLS policies
2. Verify user permissions
3. Check constraint violations
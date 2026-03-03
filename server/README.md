# Elysium Registry Server

FastAPI-based registry server for the Elysium API App Store.

## Quick Start

### Local Development

```bash
# Install dependencies
pip install -r requirements.txt

# Set up environment
cp .env.example .env
# Edit .env with your Supabase credentials

# Run server
uvicorn app.main:app --reload
```

### Production Deployment

See [VERCEL_DEPLOYMENT.md](VERCEL_DEPLOYMENT.md) for complete deployment guide.

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `POST /api/auth/logout` - Logout
- `POST /api/auth/refresh` - Refresh token
- `GET /api/auth/me` - Get current user

### Emblems (APIs)
- `GET /api/emblems` - List all emblems
- `GET /api/emblems/{name}` - Get emblem details
- `GET /api/emblems/{name}/{version}` - Get specific version
- `POST /api/emblems` - Create emblem (auth required)
- `PUT /api/emblems/{name}` - Update emblem (auth required)
- `DELETE /api/emblems/{name}` - Delete emblem (auth required)

### API Keys
- `GET /api/keys` - List user's keys (auth required)
- `POST /api/keys` - Create new key (auth required)
- `GET /api/keys/{id}` - Get key details (auth required)
- `DELETE /api/keys/{id}` - Delete key (auth required)

### Health
- `GET /health` - Health check
- `GET /` - API info

## Environment Variables

See `.env.production.example` for all available variables.

Required:
- `SUPABASE_URL` - Supabase project URL
- `SUPABASE_ANON_KEY` - Public anonymous key
- `SUPABASE_SERVICE_KEY` - Service role key (secret!)

Optional:
- `CORS_ORIGINS` - Allowed origins (default: `*`)
- `APP_NAME` - Application name
- `DEBUG` - Enable debug mode
- `RATE_LIMIT_REQUESTS` - Max requests per window
- `DOMAIN` - Custom domain for CORS

## Tech Stack

- **Framework**: FastAPI
- **Backend**: Supabase (PostgreSQL + Auth)
- **Docs**: Swagger UI at `/docs`
- **Deployment**: Vercel Serverless

## License

MIT# Build cache: 2

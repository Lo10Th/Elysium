# Environment Variables

This file documents all environment variables used by the Elysium Registry server.

## Required Variables

### `SUPABASE_URL`
- **Description:** Your Supabase project URL
- **Format:** `https://<project-id>.supabase.co`
- **Where to find:** Supabase Dashboard > Settings > API > Configuration
- **Example:** `https://abcdefghijklmnop.supabase.co`

### `SUPABASE_KEY`
- **Description:** Anonymous public key (safe for client-side)
- **Formerly called:** `SUPABASE_ANON_KEY`
- **Format:** JWT token starting with `eyJ...`
- **Where to find:** Supabase Dashboard > Settings > API > Project API keys > anon public
- **Security:** Safe to expose in client-side code
- **Example:** `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImFiY2RlZmdoaWprbG1ub3AiLCJyb2xlIjoiYW5vbiIsImlhdCI6MTYxMjU0MzYwMCwiZXhwIjoxOTI3ODA0MDAwfQ...`

### `SUPABASE_SERVICE_ROLE_KEY`
- **Description:** Service role key (admin access)
- **Formerly called:** `SUPABASE_SERVICE_KEY`
- **Format:** JWT token starting with `eyJ...`
- **Where to find:** Supabase Dashboard > Settings > API > Project API keys > service_role
- **Security:** ⚠️ **MUST KEEP SECRET** - Never use in client-side code
- **Example:** `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImFiY2RlZmdoaWprbG1ub3AiLCJyb2xlIjoic2VydmljZV9yb2xlIiwiaWF0IjoxNjEyNTQzNjAwLCJleHAiOjE5Mjc4MDQwMDB9...`

## Optional Variables

### `APP_NAME`
- **Description:** Application name shown in API docs
- **Default:** `Elysium Registry`
- **Example:** `My API Registry`

### `APP_VERSION`
- **Description:** API version displayed in health endpoint
- **Default:** `1.0.0`
- **Example:** `2.1.0`

### `DEBUG`
- **Description:** Enable debug mode (more verbose logging)
- **Default:** `false`
- **Values:** `true` or `false`
- **Security:** Enable only in development, never in production

### `CORS_ORIGINS`
- **Description:** Comma-separated list of allowed origins for CORS
- **Default:** `*` (all origins)
- **Example:** `https://example.com,https://app.example.com`
- **Security:** Use specific domains in production

### `RATE_LIMIT_REQUESTS`
- **Description:** Maximum requests per IP per time window
- **Default:** `100`
- **Example:** `1000`

### `RATE_LIMIT_WINDOW`
- **Description:** Time window for rate limiting (in seconds)
- **Default:** `60` (1 minute)
- **Example:** `300` (5 minutes)

### `DOMAIN`
- **Description:** Custom domain for automatic CORS configuration
- **Example:** `ely.karlharrenga.com`
- **Note:** When set, `CORS_ORIGINS` is automatically configured

## Local Development Setup

1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Fill in your Supabase credentials

3. Run the server:
   ```bash
   uvicorn app.main:app --reload
   ```

## Production Setup (Vercel)

1. Go to Vercel Dashboard > Your Project > Settings > Environment Variables

2. Add all required variables:
   - `SUPABASE_URL`
   - `SUPABASE_KEY`
   - `SUPABASE_SERVICE_ROLE_KEY`

3. (Optional) Add custom settings:
   - `CORS_ORIGINS`
   - `DOMAIN`
   - etc.

4. Deploy or Redeploy

## Security Checklist

- [ ] `SUPABASE_SERVICE_ROLE_KEY` is not in git repository
- [ ] `SUPABASE_SERVICE_ROLE_KEY` is encrypted in Vercel
- [ ] `.env` is in `.gitignore`
- [ ] Different keys for development/staging/production
- [ ] Keys have appropriate permissions (not over-privileged)

## Common Issues

### Error: "Invalid API key"
- Check key format (should start with `eyJ`)
- Verify you copied the entire key
- Ensure using new naming: `SUPABASE_KEY` not `SUPABASE_ANON_KEY`

### Error: "Permission denied"
- Verify `SUPABASE_SERVICE_ROLE_KEY` is correct
- Check RLS policies in Supabase
- Ensure migrations have been run

### Error: "CORS"
- Set `CORS_ORIGINS=*` for testing
- Use specific domain in production
- Set `DOMAIN` for automatic CORS configuration
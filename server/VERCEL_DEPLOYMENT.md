# Vercel Deployment Guide for Elysium Registry

This guide walks you through deploying the Elysium Registry server to Vercel with a Supabase backend.

## Prerequisites

- [Vercel account](https://vercel.com/signup) (free tier works)
- [Supabase account](https://supabase.com) (free tier works)
- [GitHub account](https://github.com) with this repository

## Step 1: Set Up Supabase

### 1.1 Create a Supabase Project

1. Go to [supabase.com](https://supabase.com)
2. Click "New Project"
3. Choose organization and enter project name: `elysium-registry`
4. Set a secure database password (save this!)
5. Choose a region close to your users
6. Click "Create new project"
7. Wait for project to be provisioned (~2 minutes)

### 1.2 Get Your Credentials

1. In your Supabase project dashboard, go to **Settings** > **API**
2. Copy the following values:
   - **Project URL** → This is your `SUPABASE_URL`
   - **anon public** → This is your `SUPABASE_ANON_KEY`
   - **service_role** → This is your `SUPABASE_SERVICE_KEY` (⚠️ Keep secret!)

### 1.3 Run Database Migrations

1. In Supabase dashboard, go to **SQL Editor**
2. Click "New query"
3. Copy the entire contents of `supabase/migrations/20240303_initial_schema.sql`
4. Paste into the editor
5. Click "Run" to execute

**This creates:**
- `profiles` table (user profiles)
- `emblems` table (API definitions)
- `emblem_versions` table (version history)
- `api_keys` table (key management)
- `emblem_pulls` table (analytics)
- Proper indexes, RLS policies, and triggers

**Verify success:**
- Go to **Table Editor** → You should see all tables
- Check each table has columns and indexes

## Step 2: Deploy to Vercel

### 2.1 Push to GitHub

Ensure your repository is pushed to GitHub:

```bash
cd elysium
git add server/
git commit -m "feat: add Vercel deployment configuration"
git push origin main
```

### 2.2 Create Vercel Project

1. Go to [vercel.com](https://vercel.com)
2. Click "Add New..." → "Project"
3. Import your GitHub repository
4. Configure the project:
   - **Framework Preset**: Other
   - **Root Directory**: `server`
   - **Build Command**: (leave empty)
   - **Output Directory**: (leave empty)
   - **Install Command**: `pip install -r requirements.txt`

### 2.3 Add Environment Variables

In Vercel project settings:

1. Go to **Settings** → **Environment Variables**
2. Add the following:

   | Name | Value | Environment |
   |------|-------|-------------|
   | `SUPABASE_URL` | `https://xxx.supabase.co` | Production, Preview, Development |
   | `SUPABASE_ANON_KEY` | `your-anon-key` | Production, Preview, Development |
   | `SUPABASE_SERVICE_KEY` | `your-service-role-key` | Production, Preview, Development |
   | `CORS_ORIGINS` | `*` | Production, Preview, Development |
   
   (Optional):
   | `APP_NAME` | `Elysium Registry` | Production |
   | `DEBUG` | `false` | Production |
   | `RATE_LIMIT_REQUESTS` | `100` | Production |
   | `RATE_LIMIT_WINDOW` | `60` | Production |

### 2.4 Deploy

1. Click "Deploy"
2. Wait for build to complete (~1-2 minutes)
3. You'll get a URL like: `elysium-registry.vercel.app`

## Step 3: Configure Custom Domain

### 3.1 Add Custom Domain in Vercel

1. Go to your Vercel project → **Settings** → **Domains**
2. Enter your domain: `ely.karlharrenga.com`
3. Click "Add"

### 3.2 Configure DNS

Add these DNS records to your domain provider (e.g., Cloudflare, Namecheap):

**Option A: Using A Record**
```
Type: A
Name: ely
Value: 76.76.21.21
TTL: 3600
```

**Option B: Using CNAME (recommended for subdomains)**
```
Type: CNAME
Name: ely
Value: cname.vercel-dns.com
TTL: 3600
```

### 3.3 Wait for SSL

Vercel automatically provisions SSL certificates:
- Wait 1-5 minutes for DNS propagation
- Vercel will provision Let's Encrypt SSL
- Status will change to "Valid Configuration"

### 3.4 Update Environment Variables (Optional)

If using custom domain, update CORS:

1. Go to Vercel → Settings → Environment Variables
2. Update `CORS_ORIGINS` to: `https://ely.karlharrenga.com`
3. Deploy again (Settings → Deployments → ... → Redeploy)

## Step 4: Verify Deployment

### 4.1 Test Health Endpoint

```bash
curl https://ely.karlharrenga.com/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

### 4.2 Test API Documentation

Visit: `https://ely.karlharrenga.com/docs`

You should see the Swagger UI with all endpoints.

### 4.3 Test Registration

```bash
curl -X POST https://ely.karlharrenga.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "username": "testuser"
  }'
```

## Step 5: Configure CLI to Use Production

Update your CLI configuration:

```bash
# In elysium/cli
export ELYSIUM_REGISTRY_URL=https://ely.karlharrenga.com
```

Or set in config file:

```yaml
# ~/.elysium/config.yaml
registry: https://ely.karlharrenga.com
```

## Troubleshooting

### Issue: "500 Internal Server Error"

**Cause:** Database tables not created

**Solution:**
1. Go to Supabase → SQL Editor
2. Re-run `supabase/migrations/20240303_initial_schema.sql`
3. Check Table Editor for tables

### Issue: "401 Unauthorized"

**Cause:** Supabase credentials incorrect

**Solution:**
1. Verify `SUPABASE_URL` format: `https://xxx.supabase.co`
2. Check `SUPABASE_ANON_KEY` starts with `eyJ`
3. Ensure `SUPABASE_SERVICE_KEY` is correct

### Issue: "CORS Error"

**Cause:** CORS misconfiguration

**Solution:**
1. Set `CORS_ORIGINS=*` for testing
2. For production, set specific domain: `https://ely.karlharrenga.com`

### Issue: "404 Not Found"

**Cause:** Vercel routing issue

**Solution:**
1. Verify `vercel.json` exists in `server/` directory
2. Ensure `api/index.py` exists
3. Check Vercel logs for build errors

### Issue: DNS Not Resolving

**Cause:** DNS propagation delay

**Solution:**
1. Wait up to 24 hours (usually 1-5 minutes)
2. Check propagation: `dig ely.karlharrenga.com`
3. Flush DNS cache: `sudo dscacheutil -flushcache` (macOS)

## Security Checklist

- [ ] `SUPABASE_SERVICE_KEY` is kept secret (not in git)
- [ ] RLS (Row Level Security) policies enabled
- [ ] CORS configured for your domain only
- [ ] Rate limiting enabled
- [ ] SSL certificate active (check green lock icon)
- [ ] Database password is strong
- [ ] API keys are hashed before storage

## Monitoring

### Vercel Logs
1. Go to Vercel → Your Project → Logs
2. Real-time logs for debugging

### Supabase Logs
1. Go to Supabase → Logs
2. Database queries and API calls

### Recommended Monitoring Tools
- [Sentry](https://sentry.io) for error tracking
- [Supabase Dashboard](https://supabase.com/dashboard) for database metrics

## Scaling

### Free Tier Limits

**Supabase:**
- 500MB database
- 1GB bandwidth
- 50MB file storage

**Vercel:**
- 100GB bandwidth
- 100 deployments/day
- Serverless function timeout: 10s

### If You Need More

**Supabase Pro ($25/mo):**
- 8GB database
- 250GB bandwidth
- Daily backups

**Vercel Pro ($20/mo):**
- 1TB bandwidth
- Unlimited team members
- Advanced analytics

## Support

- [Elysium Docs](https://github.com/Lo10Th/Elysium)
- [Vercel Documentation](https://vercel.com/docs)
- [Supabase Documentation](https://supabase.com/docs)
- [FastAPI Documentation](https://fastapi.tiangolo.com)

## License

MIT License - See LICENSE file for details.
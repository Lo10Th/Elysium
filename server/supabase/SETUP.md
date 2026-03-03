# Supabase Setup Guide (Updated for API v2+)

This guide covers setting up Supabase with the **new API key naming convention**.

## What Changed? 🆕

Supabase updated their API key naming. Here's the mapping:

| Old Name | New Name | Location in Dashboard |
|----------|----------|----------------------|
| `SUPABASE_ANON_KEY` | `SUPABASE_KEY` | Settings > API > anon public |
| `SUPABASE_SERVICE_KEY` | `SUPABASE_SERVICE_ROLE_KEY` | Settings > API > service_role |

## Getting Your Credentials

1. Go to [Supabase Dashboard](https://supabase.com/dashboard)
2. Select your project
3. Navigate to **Settings** (gear icon) → **API**
4. You'll see:

```
Configuration
├─ Project URL
│  └─ https://xxxxx.supabase.co  ← SUPABASE_URL
│
└─ Project API keys
   ├─ anon public     ← SUPABASE_KEY (safe for client)
   └─ service_role    ← SUPABASE_SERVICE_ROLE_KEY (secret!)
```

## Setting Environment Variables

### For Vercel Deployment

```bash
# In Vercel Dashboard > Settings > Environment Variables

SUPABASE_URL=https://xxxxx.supabase.co
SUPABASE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### For Local Development

```bash
# In server/.env

SUPABASE_URL=https://xxxxx.supabase.co
SUPABASE_KEY=your-anon-public-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
```

## Testing Your Setup

You can test if your keys work:

```bash
# Test anon key (public)
curl "https://xxxxx.supabase.co/rest/v1/" \
  -H "apikey: your-SUPABASE_KEY" \
  -H "Authorization: Bearer your-SUPABASE_KEY"

# Should return: {"message":"Not Found"} (404 is okay, means auth works)
```

## Security Notes

✅ **SUPABASE_KEY (anon)** - Safe to use in:
- Client-side code
- Browser apps
- Mobile apps
- Public repositories

⚠️ **SUPABASE_SERVICE_ROLE_KEY** - Must keep secret:
- ❌ Never in client-side code
- ❌ Never in public repositories
- ✅ Only in server-side code
- ✅ Only in secure environment variables
- ✅ Use Vercel Environment Variables (encrypted)

## Troubleshooting

### Error: "Invalid API key"

**Cause:** Using old key name or wrong key

**Solution:**
1. Check you're using the new names: `SUPABASE_KEY`, `SUPABASE_SERVICE_ROLE_KEY`
2. Verify keys are correct (start with `eyJ...`)
3. Ensure you copied the entire key (no truncation)

### Error: "JWT expired"

**Cause:** Old key format

**Solution:**
1. Go to Supabase Dashboard
2. Settings > API
3. Copy the **new** keys
4. Update your environment variables
5. Redeploy on Vercel

### Error: "permission denied for table"

**Cause:** RLS policies not set up

**Solution:**
1. Run the migration SQL in Supabase SQL Editor
2. Verify tables exist in Table Editor
3. Check RLS is enabled on each table

## Need Help?

- [Supabase Documentation](https://supabase.com/docs)
- [Supabase Discord](https://discord.supabase.com)
- [GitHub Issues](https://github.com/supabase/supabase)
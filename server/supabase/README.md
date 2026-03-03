# Supabase Configuration

This directory contains database migrations for setting up the Elysium Registry database schema.

## Files

- `20240303_initial_schema.sql` - Complete database schema initialization

## How to Run

### Option 1: Via Supabase Dashboard (Recommended)

1. Go to [Supabase Dashboard](https://supabase.com/dashboard)
2. Select your project
3. Click **SQL Editor** in the sidebar
4. Click "New query"
5. Copy/paste the entire contents of `20240303_initial_schema.sql`
6. Click "Run"

### Option 2: Via Supabase CLI

If you have the Supabase CLI installed:

```bash
# Install Supabase CLI
npm install -g supabase

# Login
supabase login

# Link to your project
supabase link --project-ref your-project-id

# Push migrations
supabase db push
```

## Schema Overview

The migration creates:

### Tables

1. **profiles** - User profiles
   - Stores username, email, avatar, bio
   - Links to Supabase auth.users

2. **emblems** - API definitions
   - Name, description, category, tags
   - License, repository, homepage URLs
   - Download count, timestamps

3. **emblem_versions** - Version history
   - YAML content, README, changelog
   - Semantic versioning support

4. **api_keys** - API key management
   - Key hashes (never stores plain text)
   - Expiration dates
   - Usage tracking

5. **emblem_pulls** - Analytics
   - Download tracking
   - User agent, IP address

### Security

- **Row Level Security (RLS)** enabled on all tables
- **Policies** for proper access control
- **Triggers** for auto-updates and analytics

### Indexes

Performance indexes on:
- `emblems(name)` - Fast lookups by name
- `emblems(category)` - Category filtering
- `emblems(tags)` - Tag-based search
- `emblem_versions(emblem_id)` - Version lookups
- `api_keys(user_id)` - User key lookups

## Verify Installation

After running the migration:

1. Go to **Table Editor** in Supabase Dashboard
2. You should see all 5 tables
3. Click on any table to view columns and data

## Rollback (if needed)

To completely reset:

```sql
-- Warning: This deletes all data!
drop table if exists emblem_pulls cascade;
drop table if exists api_keys cascade;
drop table if exists emblem_versions cascade;
drop table if exists emblems cascade;
drop table if exists profiles cascade;

-- Then re-run the migration
```

## Updates

When we add new tables or columns:

1. Create a new migration file: `YYYYMMDD_description.sql`
2. Run it through Supabase SQL Editor
3. Never modify existing migration files
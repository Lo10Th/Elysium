# Supabase Emblem

Supabase API — database tables, row-level queries, and auth user management via REST and Auth APIs.

## Setup

### 1. Get Your Supabase Credentials

1. Go to your [Supabase project dashboard](https://app.supabase.com)
2. Navigate to **Settings → API**
3. Copy your **Project URL** (e.g. `https://abcdefgh.supabase.co`)
4. Copy your **service_role** key (needed for auth admin endpoints)

### 2. Configure Environment

```bash
# Set your service role key
export SUPABASE_SERVICE_KEY=your-service-role-key-here
```

### 3. Update Base URL

Edit `emblem.yaml` and replace `https://your-project-ref.supabase.co` with your actual project URL:

```yaml
baseUrl: https://abcdefgh.supabase.co
```

### 4. Pull the Emblem

```bash
ely pull supabase
```

## Available Actions

| Action | Method | Path | Description |
|--------|--------|------|-------------|
| `list-tables` | GET | `/rest/v1/` | List all tables exposed via the REST API |
| `query-table` | GET | `/rest/v1/{table}` | Query rows with filters and pagination |
| `insert-row` | POST | `/rest/v1/{table}` | Insert one or more rows |
| `update-row` | PATCH | `/rest/v1/{table}` | Update rows matching filters |
| `delete-row` | DELETE | `/rest/v1/{table}` | Delete rows matching filters |
| `list-users` | GET | `/auth/v1/admin/users` | List all auth users |
| `create-user` | POST | `/auth/v1/admin/users` | Create a new auth user |

## Examples

### List All Tables

```bash
ely execute supabase list-tables --param apikey=$SUPABASE_SERVICE_KEY
```

### Query Rows from a Table

```bash
# Get all rows from the "users" table
ely execute supabase query-table \
  --param table=users \
  --param apikey=$SUPABASE_SERVICE_KEY

# Get specific columns with ordering and limit
ely execute supabase query-table \
  --param table=users \
  --param select=id,email,created_at \
  --param order=created_at.desc \
  --param limit=20 \
  --param apikey=$SUPABASE_SERVICE_KEY
```

### Insert a Row

```bash
ely execute supabase insert-row \
  --param table=profiles \
  --param apikey=$SUPABASE_SERVICE_KEY \
  --body '{"username": "jdoe", "full_name": "John Doe", "bio": "Hello world"}'
```

### Update Rows

Use [PostgREST filter operators](https://postgrest.org/en/stable/references/api/tables_views.html#horizontal-filtering) as query parameters to target rows (e.g. `?id=eq.1`):

```bash
ely execute supabase update-row \
  --param table=profiles \
  --param apikey=$SUPABASE_SERVICE_KEY \
  --body '{"bio": "Updated bio"}' \
  --query "id=eq.42"
```

### Delete Rows

```bash
ely execute supabase delete-row \
  --param table=profiles \
  --param apikey=$SUPABASE_SERVICE_KEY \
  --query "id=eq.42"
```

### List Auth Users

```bash
ely execute supabase list-users \
  --param apikey=$SUPABASE_SERVICE_KEY \
  --param page=1 \
  --param per_page=50
```

### Create an Auth User

```bash
ely execute supabase create-user \
  --param apikey=$SUPABASE_SERVICE_KEY \
  --body '{"email": "newuser@example.com", "password": "secure-password", "email_confirm": true}'
```

## Authentication Notes

- **Database operations** (`query-table`, `insert-row`, `update-row`, `delete-row`, `list-tables`): Can use either the `anon` key (subject to Row Level Security policies) or the `service_role` key (bypasses RLS).
- **Auth admin operations** (`list-users`, `create-user`): Require the `service_role` key. Never expose this key in client-side code.

The `apikey` parameter is Supabase's dual-authentication mechanism: the emblem auth configuration sets the `Authorization: Bearer <key>` header from `SUPABASE_SERVICE_KEY`, while the `apikey` header (passed as a parameter) is also required by Supabase's gateway to identify the project. Both must use the same service role key value.

## Row Filtering

For `query-table`, `update-row`, and `delete-row`, you can use PostgREST filter operators as additional query parameters:

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equals | `?id=eq.1` |
| `neq` | Not equals | `?status=neq.inactive` |
| `lt` / `lte` | Less than / or equal | `?age=lt.30` |
| `gt` / `gte` | Greater than / or equal | `?score=gte.90` |
| `like` | Pattern match | `?name=like.J*` |
| `ilike` | Case-insensitive pattern | `?email=ilike.*@gmail.com` |
| `is` | Is null / true / false | `?deleted_at=is.null` |
| `in` | In a list | `?status=in.(active,pending)` |

## Error Handling

| Code | Description |
|------|-------------|
| 400 | Bad request — invalid parameters or filter syntax |
| 401 | Unauthorized — missing or invalid API key |
| 403 | Forbidden — RLS policy blocked the operation |
| 404 | Not found — table or resource does not exist |
| 409 | Conflict — unique constraint violation |
| 500 | Internal server error |

## Resources

- [Supabase API Reference](https://supabase.com/docs/reference)
- [PostgREST Filtering](https://postgrest.org/en/stable/references/api/tables_views.html)
- [Supabase Auth Admin API](https://supabase.com/docs/reference/javascript/auth-admin-listusers)

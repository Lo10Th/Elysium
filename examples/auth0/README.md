# Auth0 Emblem

Auth0 Management API v2 for managing users in your tenant via the [Auth0 Management API](https://auth0.com/docs/api/management/v2).

## Setup

### 1. Get Your Auth0 Credentials

1. Go to your [Auth0 Dashboard](https://manage.auth0.com)
2. Navigate to your tenant
3. Go to **APIs → Auth0 Management API → API Explorer**
4. Generate or copy a Management API token

Or create a machine-to-machine application:

1. Go to **Applications → Applications**
2. Create a new **Machine to Machine** application
3. Select the **Auth0 Management API** and authorize the scopes you need
4. Use the client credentials to obtain a token

### 2. Configure Environment

```bash
export AUTH0_MANAGEMENT_TOKEN=your_management_api_token_here
```

### 3. Update Base URL

Edit `emblem.yaml` and replace `YOUR_DOMAIN` with your Auth0 tenant domain:

```yaml
baseUrl: https://my-tenant.auth0.com/api/v2
```

For regional tenants (e.g., US, EU, AU), use the appropriate domain:

```yaml
baseUrl: https://my-tenant.us.auth0.com/api/v2
baseUrl: https://my-tenant.eu.auth0.com/api/v2
baseUrl: https://my-tenant.au.auth0.com/api/v2
```

### 4. Pull the Emblem

```bash
ely pull auth0
```

## Actions

### `list-users` — List users

Retrieve a paginated list of users with optional filtering and sorting.

```bash
# List all users (default: 50 per page)
ely execute auth0 list-users

# Pagination
ely execute auth0 list-users --page 0 --per_page 100

# Include total count
ely execute auth0 list-users --include_totals true

# Search by email (Lucene syntax)
ely execute auth0 list-users --q 'email:"user@example.com"'

# Search by connection
ely execute auth0 list-users --q 'identities.connection:"google-oauth2"'

# Sort by email ascending
ely execute auth0 list-users --sort "email:1"

# Sort by created_at descending
ely execute auth0 list-users --sort "created_at:-1"

# Select specific fields
ely execute auth0 list-users --fields "user_id,email,name,created_at"
```

---

### `get-user` — Get user details

Retrieve details for a specific user by their Auth0 user ID.

```bash
ely execute auth0 get-user --id "auth0|abc123"

# Select specific fields
ely execute auth0 get-user --id "auth0|abc123" --fields "user_id,email,picture"
```

---

### `create-user` — Create a new user

Create a user in a database connection.

```bash
# Create a user with email and password
ely execute auth0 create-user \
  --body '{
    "email": "john.doe@example.com",
    "password": "SecurePassword123!",
    "connection": "Username-Password-Authentication",
    "email_verified": false
  }'

# Create a user with full profile
ely execute auth0 create-user \
  --body '{
    "email": "jane.doe@example.com",
    "password": "SecurePassword123!",
    "connection": "Username-Password-Authentication",
    "name": "Jane Doe",
    "given_name": "Jane",
    "family_name": "Doe",
    "nickname": "jane",
    "email_verified": true
  }'

# Create a user without password (for passwordless)
ely execute auth0 create-user \
  --body '{
    "email": "user@example.com",
    "connection": "email",
    "email_verified": true
  }'

# Create a user with phone number
ely execute auth0 create-user \
  --body '{
    "phone_number": "+14155552671",
    "connection": "sms",
    "phone_verified": false
  }'

# Create a blocked user
ely execute auth0 create-user \
  --body '{
    "email": "blocked@example.com",
    "password": "SecurePassword123!",
    "connection": "Username-Password-Authentication",
    "blocked": true
  }'
```

---

### `update-user` — Update a user

Update a user's profile or account settings.

```bash
# Update user email
ely execute auth0 update-user \
  --id "auth0|abc123" \
  --body '{
    "email": "newemail@example.com",
    "connection": "Username-Password-Authentication"
  }'

# Update profile fields
ely execute auth0 update-user \
  --id "auth0|abc123" \
  --body '{
    "name": "John Smith",
    "given_name": "John",
    "family_name": "Smith",
    "nickname": "johnsmith"
  }'

# Change password
ely execute auth0 update-user \
  --id "auth0|abc123" \
  --body '{
    "password": "NewSecurePassword456!",
    "connection": "Username-Password-Authentication"
  }'

# Verify email
ely execute auth0 update-user \
  --id "auth0|abc123" \
  --body '{"email_verified": true}'

# Block a user
ely execute auth0 update-user \
  --id "auth0|abc123" \
  --body '{"blocked": true}'

# Unblock a user
ely execute auth0 update-user \
  --id "auth0|abc123" \
  --body '{"blocked": false}'
```

---

### `delete-user` — Delete a user

Permanently remove a user from the tenant.

```bash
ely execute auth0 delete-user --id "auth0|abc123"
```

**Warning:** This action is irreversible. The user and all their associated data will be permanently deleted.

## Common Workflows

### Search for a user by email

```bash
ely execute auth0 list-users --q 'email:"user@example.com"' --search_engine v3
```

### Find blocked users

```bash
ely execute auth0 list-users --q 'blocked:true' --search_engine v3
```

### Find unverified users

```bash
ely execute auth0 list-users --q 'email_verified:false' --search_engine v3
```

### Find users by login count

```bash
ely execute auth0 list-users --q 'logins_count:{5 TO *}' --search_engine v3
```

### Update user then verify the change

```bash
# 1. Update the user
ely execute auth0 update-user --id "auth0|abc123" --body '{"name": "Updated Name"}'

# 2. Verify the change
ely execute auth0 get-user --id "auth0|abc123" --fields "user_id,name"
```

## Authentication

This emblem uses a **Bearer token** stored in `AUTH0_MANAGEMENT_TOKEN`. The token must have the necessary Management API scopes for the actions you want to perform:

| Action | Required Scopes |
|--------|-----------------|
| `list-users` | `read:users` |
| `get-user` | `read:users` |
| `create-user` | `create:users` |
| `update-user` | `update:users` |
| `delete-user` | `delete:users` |

Generate a token from the Auth0 Dashboard or use the client credentials flow with a Machine to Machine application.

## Lucene Query Syntax

The `q` parameter in `list-users` supports Lucene query syntax:

| Query | Description |
|-------|-------------|
| `email:"user@example.com"` | Exact match |
| `name:*john*` | Wildcard search |
| `logins_count:[10 TO 100]` | Range search |
| `blocked:true` | Boolean filter |
| `created_at:[2024-01-01 TO 2024-12-31]` | Date range |
| `connection:"google-oauth2"` | Filter by connection |
| `-email_verified:true` | Negation (not verified) |

For complex queries, use `search_engine: v3` for improved search capabilities.

## Error Handling

| Code | Description |
|------|-------------|
| 400 | Bad request — invalid parameters |
| 401 | Unauthorized — missing or invalid token |
| 403 | Forbidden — insufficient scope for this action |
| 404 | User not found |
| 409 | Conflict — user already exists (email/username conflict) |

## Resources

- [Auth0 Management API Reference](https://auth0.com/docs/api/management/v2)
- [User Search Query Syntax](https://auth0.com/docs/users/search/v3/query-syntax)
- [Management API Scopes](https://auth0.com/docs/api/management/v2#scopes)
- [Machine to Machine Applications](https://auth0.com/docs/applications/machine-to-machine)
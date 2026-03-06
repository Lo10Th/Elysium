# Sentry Emblem

Track errors, manage issues, and organize projects through the [Sentry REST API](https://docs.sentry.io/api/) using Elysium.

## Setup

Export your Sentry auth token (create one at **Settings → Auth Tokens** on sentry.io):

```bash
export SENTRY_AUTH_TOKEN=your_sentry_auth_token_here
```

## Actions

### `list-issues` — List error issues

List error issues for a project within an organization with filtering options.

```bash
# All issues in an organization
ely execute sentry list-issues --organization_slug my-org

# Filter by project ID
ely execute sentry list-issues --organization_slug my-org --project 12345

# Filter by environment
ely execute sentry list-issues --organization_slug my-org --environment production

# Only unresolved issues
ely execute sentry list-issues --organization_slug my-org --query "is:unresolved"

# Issues from last 24 hours
ely execute sentry list-issues --organization_slug my-org --statsPeriod 24h

# Limit results
ely execute sentry list-issues --organization_slug my-org --limit 50
```

---

### `get-issue` — Get issue details

Fetch detailed information about a specific error issue.

```bash
ely execute sentry get-issue --issue_id 1234567890
```

Returns the full `Issue` object including title, culprit, status, count, user count, first/last seen timestamps, and associated project information.

---

### `list-projects` — List all projects

Retrieve all projects for an organization.

```bash
# All projects
ely execute sentry list-projects --organization_slug my-org
```

---

### `create-project` — Create a new project

Create a project within a team in your organization.

```bash
# Create a Python project
ely execute sentry create-project \
  --organization_slug my-org \
  --team_slug my-team \
  --name "My New App" \
  --platform python

# Create a JavaScript project with a custom slug
ely execute sentry create-project \
  --organization_slug my-org \
  --team_slug my-team \
  --name "Frontend Dashboard" \
  --slug frontend-dashboard \
  --platform javascript
```

## Common Workflows

### Find unresolved issues in production

```bash
ely execute sentry list-issues \
  --organization_slug my-org \
  --environment production \
  --query "is:unresolved"
```

### Get details for the top error

```bash
# 1. List issues sorted by frequency
ely execute sentry list-issues --organization_slug my-org --limit 1

# 2. Copy the issue ID and get full details
ely execute sentry get-issue --issue_id <issue_id>
```

### Create a project for a new service

```bash
# First list projects to see existing ones
ely execute sentry list-projects --organization_slug my-org

# Create a new Node.js monitoring project
ely execute sentry create-project \
  --organization_slug my-org \
  --team_slug backend \
  --name "Payment Service" \
  --platform node
```

### Filter issues by time window

```bash
# Issues in the last hour
ely execute sentry list-issues --organization_slug my-org --statsPeriod 1h

# Issues in the last 7 days
ely execute sentry list-issues --organization_slug my-org --statsPeriod 7d

# Issues in the last 30 days
ely execute sentry list-issues --organization_slug my-org --statsPeriod 30d
```

## Authentication

This emblem uses a **Bearer token** stored in `SENTRY_AUTH_TOKEN`. Generate a token with appropriate scopes at:  
<https://sentry.io/settings/account/auth/>

Required scopes:
- `org:read` — to list projects and issues
- `project:write` — to create new projects

## Issue Status Values

| Status | Meaning |
|--------|---------|
| `resolved` | Issue has been marked as resolved |
| `unresolved` | Issue is active and needs attention |
| `ignored` | Issue has been silenced/ignored |

## Issue Severity Levels

| Level | Description |
|-------|-------------|
| `fatal` | Application crash or critical failure |
| `error` | Exception that was caught |
| `warning` | Unexpected but recoverable condition |
| `info` | General informational event |
| `debug` | Debug-level information |

## Resources

- [Sentry REST API Reference](https://docs.sentry.io/api/)
- [Issues API](https://docs.sentry.io/api/events/list-an-organizations-issues/)
- [Projects API](https://docs.sentry.io/api/projects/)
- [Search Syntax](https://docs.sentry.io/product/issues/search/)
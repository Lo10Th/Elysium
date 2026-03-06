# Linear Emblem

Linear issue tracking API — create and manage issues, projects, and teams via GraphQL using Elysium.

## Setup

### 1. Create a Linear API Key

1. Go to your [Linear settings](https://linear.app/settings/api)
2. Navigate to **Settings → API**
3. Click **Create API Key**
4. Copy your personal API key

### 2. Configure Environment

```bash
export LINEAR_API_KEY=your_linear_api_key_here
```

### 3. Pull the Emblem

```bash
ely pull linear
```

## Available Actions

| Action | Method | Path | Description |
|--------|--------|------|-------------|
| `list-issues` | POST | `/graphql` | List issues with optional filters for team, project, state, and priority |
| `get-issue` | POST | `/graphql` | Get a single issue by UUID or identifier |
| `create-issue` | POST | `/graphql` | Create a new issue in a Linear team |
| `update-issue` | POST | `/graphql` | Update an existing issue by UUID |
| `list-projects` | POST | `/graphql` | List all projects with optional pagination |

## Examples

### List Issues

Linear uses GraphQL, so actions require a `query` and optional `variables`:

```bash
# List all issues (default pagination)
ely execute linear list-issues \
  --param 'query={ issues { nodes { id identifier title priority url } pageInfo { hasNextPage endCursor } } }'

# List issues for a specific team
ely execute linear list-issues \
  --param 'query={ issues(filter: { team: { id: { eq: "team-uuid" } } }) { nodes { id identifier title } } }'

# List issues by priority (1=urgent, 2=high, 3=medium, 4=low)
ely execute linear list-issues \
  --param 'query={ issues(filter: { priority: { eq: 1 } }) { nodes { id identifier title } } }'

# Paginated query using variables
ely execute linear list-issues \
  --param 'query={ issues(first: 50, after: "cursor-value") { nodes { id identifier title } pageInfo { endCursor } } }'
```

### Get a Single Issue

Fetch an issue by its UUID or human-readable identifier (e.g., `ENG-123`):

```bash
# Get by identifier
ely execute linear get-issue \
  --param 'query={ issue(id: "ENG-123") { id identifier title description priority url createdAt } }'

# Get by UUID
ely execute linear get-issue \
  --param 'query={ issue(id: "uuid-here") { id identifier title state { id name } } }'
```

### Create a New Issue

```bash
# Create an issue with required fields
ely execute linear create-issue \
  --param 'query=mutation { issueCreate(input: { teamId: "team-uuid", title: "Fix login bug" }) { success issue { id identifier url } } }'

# Create with optional fields
ely execute linear create-issue \
  --param 'query=mutation { issueCreate(input: { teamId: "team-uuid", title: "Implement feature", description: "## Details\n\nMore info here", priority: 2, projectId: "project-uuid" }) { success issue { id identifier url } } }'

# Assign to a user
ely execute linear create-issue \
  --param 'query=mutation { issueCreate(input: { teamId: "team-uuid", title: "Task", assigneeId: "user-uuid" }) { success issue { id identifier } } }'
```

### Update an Issue

```bash
# Update title and priority
ely execute linear update-issue \
  --param 'query=mutation { issueUpdate(id: "issue-uuid", input: { title: "New title", priority: 1 }) { success issue { id identifier } } }'

# Update workflow state
ely execute linear update-issue \
  --param 'query=mutation { issueUpdate(id: "issue-uuid", input: { stateId: "state-uuid" }) { success issue { id state { id name } } } }'

# Reassign
ely execute linear update-issue \
  --param 'query=mutation { issueUpdate(id: "issue-uuid", input: { assigneeId: "new-user-uuid" }) { success } }'
```

### List Projects

```bash
# List all projects
ely execute linear list-projects \
  --param 'query={ projects { nodes { id name state url } pageInfo { hasNextPage endCursor } } }'

# Paginated query
ely execute linear list-projects \
  --param 'query={ projects(first: 50, after: "cursor") { nodes { id name description } } }'
```

## Priority Values

Issues use integer priority values:

| Priority | Value |
|----------|-------|
| None | 0 |
| Urgent | 1 |
| High | 2 |
| Medium | 3 |
| Low | 4 |

## Common Workflows

### Create a Bug Report

```bash
# Get team ID first
ely execute linear list-issues \
  --param 'query={ teams { nodes { id name key } } }'

# Create the bug issue with high priority
ely execute linear create-issue \
  --param 'query=mutation { issueCreate(input: { teamId: "team-uuid", title: "Bug: App crashes on login", description: "Steps to reproduce...", priority: 2 }) { success issue { id identifier url } } }'
```

### Find All Urgent Issues Across Projects

```bash
ely execute linear list-issues \
  --param 'query={ issues(filter: { priority: { eq: 1 } }) { nodes { id identifier title project { name } url } } }'
```

### Move Issue Through Workflow States

```bash
# First get available states for a team
ely execute linear get-issue \
  --param 'query={ team(id: "team-uuid") { states { nodes { id name } } } }'

# Then update the issue state
ely execute linear update-issue \
  --param 'query=mutation { issueUpdate(id: "issue-uuid", input: { stateId: "in-progress-state-uuid" }) { success } }'
```

## Authentication

This emblem uses a **Bearer token** stored in `LINEAR_API_KEY`. The token is passed in the `Authorization` header:

```
Authorization: Bearer <your-api-key>
```

Generate an API key from your Linear workspace settings. Personal API keys inherit your permissions; use a service account for automation workflows.

## GraphQL Notes

All Linear API requests are GraphQL queries sent via POST to `/graphql`:

- **Queries** (read operations): Use `query { ... }` syntax
- **Mutations** (write operations): Use `mutation { ... }` syntax
- **Variables**: Pass structured variables alongside your query for complex inputs

For full schema details, see the [Linear GraphQL API documentation](https://developers.linear.app/graphql).

## Resources

- [Linear API Documentation](https://developers.linear.app)
- [Linear GraphQL Reference](https://developers.linear.app/graphql)
- [Linear API Authentication](https://developers.linear.app/docs/authentication)
- [Linear GraphQL Playground](https://api.linear.app/graphql)
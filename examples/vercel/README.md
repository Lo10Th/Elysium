# Vercel Emblem

Deploy and manage your frontend infrastructure through the [Vercel REST API](https://vercel.com/docs/rest-api) using Elysium.

## Setup

Export your Vercel personal access token (create one at **Settings → Tokens** on vercel.com):

```bash
export VERCEL_TOKEN=your_vercel_token_here
```

## Actions

### `list-deployments` — List deployments

List recent deployments across all projects, or narrow results to a specific project.

```bash
# All recent deployments
ely execute vercel list-deployments

# Filter to a specific project
ely execute vercel list-deployments --projectId my-app

# Only READY deployments, last 10
ely execute vercel list-deployments --state READY --limit 10

# Team account — scope to a team
ely execute vercel list-deployments --teamId team_abc123 --projectId my-app
```

---

### `get-deployment` — Get deployment status

Fetch full details and the current state for a single deployment.

```bash
# Check deployment status
ely execute vercel get-deployment --id dpl_abc123xyz

# With team scope
ely execute vercel get-deployment --id dpl_abc123xyz --teamId team_abc123
```

Possible `state` values returned (mirrors the `state` filter enum in `list-deployments`):

| State | Meaning |
|-------|---------|
| `QUEUED` | Waiting to be picked up by a builder |
| `BUILDING` | Build in progress |
| `INITIALIZING` | Deployment initializing on edge network |
| `READY` | Live and serving traffic |
| `ERROR` | Build or deployment failed |
| `CANCELED` | Deployment was canceled |

---

### `create-deployment` — Create a new deployment

Trigger a deployment from a linked Git repository.

```bash
# Deploy default branch of a GitHub repo
ely execute vercel create-deployment \
  --name my-app \
  --gitSource.type github \
  --gitSource.repoId 123456789 \
  --target production

# Deploy a specific branch as a preview
ely execute vercel create-deployment \
  --name my-app \
  --gitSource.type github \
  --gitSource.repoId 123456789 \
  --gitSource.ref feature/new-landing \
  --target preview

# Deploy under a team account
ely execute vercel create-deployment \
  --teamId team_abc123 \
  --name my-app \
  --gitSource.type github \
  --gitSource.repoId 123456789
```

---

### `list-projects` — List all projects

Retrieve all projects for the authenticated account or team.

```bash
# All projects
ely execute vercel list-projects

# Search by name prefix
ely execute vercel list-projects --search my-app

# First 50 projects in a team
ely execute vercel list-projects --teamId team_abc123 --limit 50
```

---

### `create-project` — Create a new project

Create a project and link it to a Git repository.

```bash
# Minimal project (no Git link)
ely execute vercel create-project --name my-new-app

# Next.js project linked to GitHub
ely execute vercel create-project \
  --name my-nextjs-app \
  --framework nextjs \
  --gitRepository.type github \
  --gitRepository.repo my-org/my-repo

# Custom build settings
ely execute vercel create-project \
  --name my-vite-app \
  --framework vite \
  --gitRepository.type github \
  --gitRepository.repo my-org/my-repo \
  --buildCommand "npm run build" \
  --outputDirectory dist \
  --rootDirectory packages/web

# Under a team account
ely execute vercel create-project \
  --teamId team_abc123 \
  --name my-team-app \
  --framework nextjs \
  --gitRepository.type github \
  --gitRepository.repo my-org/my-repo
```

## Common Workflows

### Deploy and poll until ready

```bash
# 1. Trigger the deployment
ely execute vercel create-deployment \
  --name my-app \
  --gitSource.type github \
  --gitSource.repoId 123456789

# 2. Copy the returned deployment ID and poll for status
ely execute vercel get-deployment --id dpl_<returned_id>
```

### Find failing deployments across a project

```bash
ely execute vercel list-deployments --projectId my-app --state ERROR
```

### Audit all projects in a team

```bash
ely execute vercel list-projects --teamId team_abc123 --limit 100
```

## Authentication

This emblem uses a **Bearer token** stored in `VERCEL_TOKEN`. Generate a token at:  
<https://vercel.com/account/tokens>

Tokens can be scoped to your personal account or a specific team. Use the `teamId` query parameter on any action to scope requests to a team.

## Resources

- [Vercel REST API Reference](https://vercel.com/docs/rest-api)
- [Deployments API](https://vercel.com/docs/rest-api/endpoints/deployments)
- [Projects API](https://vercel.com/docs/rest-api/endpoints/projects)
- [Authentication](https://vercel.com/docs/rest-api#authentication)

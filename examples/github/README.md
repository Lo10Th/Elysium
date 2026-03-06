# GitHub Emblem

Interact with GitHub's REST API — repositories, issues, pull requests, and users — using Elysium.

## Setup

Export your GitHub personal access token (create one at **Settings → Developer settings → Personal access tokens** on github.com):

```bash
export GITHUB_TOKEN=your_github_token_here
```

For most actions, a token with `repo` scope is sufficient. For creating repositories or accessing private repos, ensure your token has the appropriate permissions.

## Actions

### `get-repo` — Get repository details

Fetch details for a specific repository.

```bash
ely execute github get-repo --owner octocat --repo Hello-World
```

---

### `list-repos` — List repositories

List public repositories for a user or organization.

```bash
# All repos for a user
ely execute github list-repos --username octocat

# Only repos they own
ely execute github list-repos --username octocat --type owner

# Sort by most recently pushed
ely execute github list-repos --username octocat --sort pushed --per_page 50
```

---

### `create-repo` — Create a new repository

Create a repository for the authenticated user. Requires `repo` scope.

```bash
# Public repo with README
ely execute github create-repo \
  --body '{"name": "my-new-repo", "description": "A test repository", "auto_init": true}'

# Private repo
ely execute github create-repo \
  --body '{"name": "my-private-repo", "private": true}'
```

---

### `list-issues` — List issues

List issues in a repository.

```bash
# All open issues
ely execute github list-issues --owner octocat --repo Hello-World

# All issues (open and closed)
ely execute github list-issues --owner octocat --repo Hello-World --state all

# Filter by label
ely execute github list-issues --owner octocat --repo Hello-World --labels bug,enhancement
```

---

### `get-issue` — Get issue details

Fetch a single issue by its number.

```bash
ely execute github get-issue --owner octocat --repo Hello-World --issue_number 42
```

---

### `create-issue` — Create a new issue

Create an issue in a repository. Requires `repo` scope.

```bash
ely execute github create-issue \
  --owner octocat \
  --repo Hello-World \
  --body '{"title": "Bug: Something is broken", "body": "Here are the details..."}'
```

---

### `list-pull-requests` — List pull requests

List pull requests in a repository.

```bash
# All open PRs
ely execute github list-pull-requests --owner octocat --repo Hello-World

# All PRs (open and closed)
ely execute github list-pull-requests --owner octocat --repo Hello-World --state all
```

---

### `get-pull-request` — Get pull request details

Fetch a single pull request by its number.

```bash
ely execute github get-pull-request --owner octocat --repo Hello-World --pull_number 123
```

---

### `create-pr` — Create a new pull request

Create a pull request in a repository. Requires `repo` scope.

```bash
ely execute github create-pr \
  --owner octocat \
  --repo Hello-World \
  --body '{"title": "Add new feature", "head": "feature-branch", "base": "main", "body": "This PR adds..."}'
```

---

### `get-user` — Get user profile

Fetch a user's public profile.

```bash
ely execute github get-user --username octocat
```

---

### `get-authenticated-user` — Get authenticated user

Get the profile of the authenticated user (requires `GITHUB_TOKEN`).

```bash
ely execute github get-authenticated-user
```

## Common Workflows

### Check repository info before creating an issue

```bash
# 1. Verify repo exists and get details
ely execute github get-repo --owner my-org --repo my-repo

# 2. List existing issues to avoid duplicates
ely execute github list-issues --owner my-org --repo my-repo --state open

# 3. Create the issue
ely execute github create-issue \
  --owner my-org \
  --repo my-repo \
  --body '{"title": "New bug report", "body": "Description here"}'
```

### Create a PR from a feature branch

```bash
# 1. Get the default branch
ely execute github get-repo --owner my-org --repo my-repo

# 2. Create the PR (use returned default_branch as base)
ely execute github create-pr \
  --owner my-org \
  --repo my-repo \
  --body '{"title": "Feature: New functionality", "head": "feature/my-feature", "base": "main"}'
```

## Authentication

This emblem uses a **Bearer token** stored in `GITHUB_TOKEN`. Generate a token at:  
<https://github.com/settings/tokens>

### Token Scopes

| Scope | Required For |
|-------|--------------|
| `repo` | Creating repos, issues, PRs; accessing private repos |
| `read:org` | Listing organization repositories |
| `public_repo` | Public repo operations (alternative to `repo`) |

For public read-only operations, no token is required (but you'll hit lower rate limits).

## Rate Limits

GitHub's API has rate limits:

| Auth Type | Requests/hour |
|-----------|---------------|
| Unauthenticated | 60 |
| Authenticated | 5,000 |

Rate limit headers are returned in responses. If you hit the limit, wait for the `X-RateLimit-Reset` timestamp.

## Error Handling

| Code | Error | Description |
|------|-------|-------------|
| 401 | `unauthorized` | Authentication required — check your `GITHUB_TOKEN` |
| 403 | `forbidden` | Permission denied — check token scopes |
| 404 | `not_found` | Resource doesn't exist — verify owner/repo names |
| 422 | `validation_failed` | Invalid request body — check required fields |
| 429 | `rate_limit_exceeded` | Too many requests — wait and retry |

## Resources

- [GitHub REST API Reference](https://docs.github.com/en/rest)
- [Repositories API](https://docs.github.com/en/rest/repos)
- [Issues API](https://docs.github.com/en/rest/issues)
- [Pull Requests API](https://docs.github.com/en/rest/pulls)
- [Personal Access Tokens](https://github.com/settings/tokens)
# Notion Emblem

Create and manage Notion pages, databases, and workspace content through the [Notion API](https://developers.notion.com) using Elysium.

## Setup

Export your Notion integration token (create one at **Settings → Integrations** on notion.so):

```bash
export NOTION_API_KEY=your_notion_integration_token_here
```

Share any pages or databases you want to access with your integration from **Share → Invite** in Notion.

## Actions

### `create-page` — Create a new page

Create a page inside a Notion database or as a child of an existing page.

```bash
# Create a page in a database
ely execute notion create-page \
  --parent.database_id abc123-def456 \
  --properties '{"Name": {"title": [{"text": {"content": "My New Page"}}]}}'

# Create a child page under an existing page
ely execute notion create-page \
  --parent.page_id parent-page-uuid \
  --properties '{"Title": {"title": [{"text": {"content": "Child Page"}}]}}'

# Create a page with initial content blocks
ely execute notion create-page \
  --parent.database_id abc123-def456 \
  --properties '{"Name": {"title": [{"text": {"content": "Task"}}]}}' \
  --children '[{"object": "block", "type": "paragraph", "paragraph": {"rich_text": [{"type": "text", "text": {"content": "Task description"}}]}}]'
```

---

### `list-pages` — Query database pages

Query a Notion database to list its pages with optional filters, sorts, and pagination.

```bash
# List all pages in a database
ely execute notion list-pages --database_id abc123-def456

# Filter by property value
ely execute notion list-pages \
  --database_id abc123-def456 \
  --filter '{"property": "Status", "select": {"equals": "In Progress"}}'

# Sort by property
ely execute notion list-pages \
  --database_id abc123-def456 \
  --sorts '[{"property": "Due Date", "direction": "ascending"}]'

# Paginated query
ely execute notion list-pages \
  --database_id abc123-def456 \
  --page_size 10 \
  --start_cursor next-cursor-from-previous-response
```

---

### `get-page` — Retrieve a page

Fetch full details and property values for a single page.

```bash
# Get a page by ID
ely execute notion get-page --page_id abc123-def456-789
```

---

### `update-page` — Update a page

Update properties or archive and restore a page.

```bash
# Update a property
ely execute notion update-page \
  --page_id abc123-def456-789 \
  --properties '{"Status": {"select": {"name": "Completed"}}}'

# Archive a page
ely execute notion update-page \
  --page_id abc123-def456-789 \
  --archived true

# Restore an archived page
ely execute notion update-page \
  --page_id abc123-def456-789 \
  --archived false
```

---

### `search` — Search pages and databases

Search all Notion pages and databases the integration can access by title keyword.

```bash
# Search by title keyword
ely execute notion search --query "meeting"

# Filter to pages only
ely execute notion search \
  --query "project" \
  --filter '{"property": "object", "value": "page"}'

# Filter to databases only
ely execute notion search \
  --filter '{"property": "object", "value": "database"}'

# Sort by last edited time
ely execute notion search \
  --query "notes" \
  --sort '{"timestamp": "last_edited_time", "direction": "descending"}'

# Paginated search
ely execute notion search \
  --query "report" \
  --page_size 20 \
  --start_cursor previous-cursor
```

## Common Workflows

### Create and update a task

```bash
# 1. Create a task in a database
ely execute notion create-page \
  --parent.database_id your-database-id \
  --properties '{"Name": {"title": [{"text": {"content": "New Task"}}]}, "Status": {"select": {"name": "To Do"}}}'

# 2. Copy the returned page ID and update status
ely execute notion update-page \
  --page_id returned-page-id \
  --properties '{"Status": {"select": {"name": "In Progress"}}}'
```

### Find and list all databases

```bash
ely execute notion search --filter '{"property": "object", "value": "database"}'
```

### Query all pages in a database sorted by creation date

```bash
ely execute notion list-pages \
  --database_id your-database-id \
  --sorts '[{"timestamp": "created_time", "direction": "descending"}]'
```

## Authentication

This emblem uses a **Bearer token** stored in `NOTION_API_KEY`. Create an integration at:  
<https://www.notion.so/my-integrations>

After creating the integration:
1. Copy the **Internal Integration Token**
2. Share any pages or databases with the integration from **Share → Invite** in Notion

The Notion API requires a `Notion-Version` header on all requests (e.g., `2022-06-28`). This emblem handles that automatically.

## API Versioning

Notion uses header-based API versioning. The emblem includes the required `Notion-Version` header on all requests. Check the [Notion API changelog](https://developers.notion.com/reference/changelog) for version updates.

## Error Handling

| Code | Error | Description |
|------|-------|-------------|
| 400 | `validation_error` | Invalid request body or missing required fields |
| 401 | `unauthorized` | Missing or invalid Bearer token |
| 403 | `restricted_resource` | Integration lacks permission for this resource |
| 404 | `object_not_found` | Resource not found or not shared with integration |
| 429 | `rate_limited` | API rate limit exceeded; check `Retry-After` header |
| 500 | `internal_server_error` | Unexpected server error on Notion's side |

## Pagination

`list-pages` and `search` return paginated results:

- `has_more: true` indicates additional pages
- Use `next_cursor` value as `start_cursor` in the next request
- `page_size` defaults to 100 (maximum)

## Resources

- [Notion API Reference](https://developers.notion.com/reference)
- [Pages API](https://developers.notion.com/reference/pages)
- [Databases API](https://developers.notion.com/reference/databases)
- [Search API](https://developers.notion.com/reference/search)
- [Create an Integration](https://www.notion.so/my-integrations)
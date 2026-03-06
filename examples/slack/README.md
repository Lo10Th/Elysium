# Slack Emblem

Slack Web API — send messages, manage channels, and interact with users through the [Slack API](https://api.slack.com/web).

## Setup

### 1. Create a Slack App

1. Go to [Slack Apps](https://api.slack.com/apps)
2. Click **Create New App** → **From scratch**
3. Name your app and select your workspace

### 2. Configure Bot Permissions

Navigate to **OAuth & Permissions** and add these Bot Token Scopes:

**Messaging:**
- `chat:write` — Send messages
- `chat:write.public` — Send messages to public channels without being a member
- `chat:edit` — Edit messages
- `chat:delete` — Delete messages

**Channels:**
- `channels:read` — View public channels
- `groups:read` — View private channels
- `channels:history` — View messages in public channels
- `groups:history` — View messages in private channels
- `channels:manage` — Create and manage channels

**Users:**
- `users:read` — View users
- `users:read.email` — View user email addresses

### 3. Install and Get Token

1. Click **Install to Workspace**
2. Copy the **Bot User OAuth Token** (starts with `xoxb-`)

### 4. Configure Environment

```bash
export SLACK_BOT_TOKEN=xoxb-your-token-here
```

### 5. Pull the Emblem

```bash
ely pull slack
```

## Actions

### `send-message` — Send a message

Post a message to a channel or user (DM).

```bash
# Send to a channel
ely execute slack send-message \
  --channel C012AB3CD \
  --text "Hello from Elysium!"

# Send to a user (DM)
ely execute slack send-message \
  --channel U012AB3CD \
  --text "Direct message"

# Reply in a thread
ely execute slack send-message \
  --channel C012AB3CD \
  --text "Thread reply" \
  --thread_ts 1234567890.123456

# Disable markdown parsing
ely execute slack send-message \
  --channel C012AB3CD \
  --text "*not bold*" \
  --mrkdwn false
```

---

### `update-message` — Update an existing message

Modify the content of a previously sent message.

```bash
ely execute slack update-message \
  --channel C012AB3CD \
  --ts 1234567890.123456 \
  --text "Updated message content"
```

---

### `delete-message` — Delete a message

Remove a message from a channel.

```bash
ely execute slack delete-message \
  --channel C012AB3CD \
  --ts 1234567890.123456
```

---

### `get-permalink` — Get a message permalink

Generate a permanent URL to a specific message.

```bash
ely execute slack get-permalink \
  --channel C012AB3CD \
  --message_ts 1234567890.123456
```

---

### `list-channels` — List all channels

Retrieve all channels in the workspace.

```bash
# List all public channels
ely execute slack list-channels

# Include private channels and DMs
ely execute slack list-channels \
  --types "public_channel,private_channel,im,mpim"

# Limit results
ely execute slack list-channels --limit 50

# Paginate through results
ely execute slack list-channels --cursor dGhpcyBpcyBhIGN1cnNvcg==
```

---

### `get-channel-info` — Get channel details

Fetch detailed information about a specific channel.

```bash
ely execute slack get-channel-info --channel C012AB3CD
```

Returns: channel name, privacy status, member count, topic, purpose, etc.

---

### `get-channel-history` — View channel messages

Retrieve recent messages from a channel.

```bash
# Last 100 messages (default)
ely execute slack get-channel-history --channel C012AB3CD

# Limit to 20 messages
ely execute slack get-channel-history \
  --channel C012AB3CD \
  --limit 20

# Messages within a time range (Unix timestamps)
ely execute slack get-channel-history \
  --channel C012AB3CD \
  --oldest 1704067200.000000 \
  --latest 1704153600.000000
```

---

### `create-channel` — Create a new channel

Create a public or private channel.

```bash
# Public channel
ely execute slack create-channel --name new-project-updates

# Private channel
ely execute slack create-channel \
  --name confidential-planning \
  --is_private true
```

Channel naming rules: lowercase, no spaces (use hyphens), max 80 characters.

---

### `list-users` — List all users

List all members in the workspace.

```bash
# All users
ely execute slack list-users

# Limit results
ely execute slack list-users --limit 100

# Paginate
ely execute slack list-users --cursor dXNlcl9jdXJzb3I=
```

---

### `get-user-info` — Get user details

Fetch information about a specific user.

```bash
ely execute slack get-user-info --user U012AB3CD
```

Returns: name, display name, email, timezone, status, etc.

---

### `lookup-user-by-email` — Find user by email

Look up a user by their email address.

```bash
ely execute slack lookup-user-by-email --email user@example.com
```

Returns user info if found, or `ok=false` if no match.

## Common Workflows

### Send a message and get its permalink

```bash
# 1. Send the message
ely execute slack send-message \
  --channel C012AB3CD \
  --text "Important announcement"

# 2. Copy the returned ts and create a permalink
ely execute slack get-permalink \
  --channel C012AB3CD \
  --message_ts 1234567890.123456
```

### Find and message a user by email

```bash
# 1. Look up the user
ely execute slack lookup-user-by-email --email teammate@company.com

# 2. Get their user ID from the response and send a DM
ely execute slack send-message \
  --channel U012AB3CD \
  --text "Hi! I found you via the API."
```

### Broadcast to multiple channels

```bash
for channel in C012AB3CD C0987ZYXW C1234QRST; do
  ely execute slack send-message \
    --channel $channel \
    --text "Scheduled maintenance tonight at 10 PM"
done
```

### Audit channel activity

```bash
# List all channels
ely execute slack list-channels \
  --types "public_channel,private_channel"

# Check recent activity in each
ely execute slack get-channel-history \
  --channel C012AB3CD \
  --limit 10
```

## Authentication

This emblem uses **Bearer token** authentication with a Slack Bot Token (`xoxb-` prefix).

| Token Type | Prefix | Access |
|------------|--------|--------|
| Bot Token | `xoxb-` | Actions the bot has been granted permission for |
| User Token | `xoxp-` | Actions on behalf of a specific user |

The bot token is recommended for most use cases. User tokens require additional OAuth flows.

All requests include `Authorization: Bearer <token>` header. The token is set via the `SLACK_BOT_TOKEN` environment variable.

## Response Format

Slack API responses follow this structure:

```json
{
  "ok": true,
  "channel": "C012AB3CD",
  "ts": "1234567890.123456",
  "message": {
    "text": "Hello!",
    "user": "U012AB3CD",
    "ts": "1234567890.123456"
  }
}
```

Always check the `ok` field — `true` indicates success, `false` indicates an error with details in `error`.

## Error Handling

| Error | Description |
|-------|-------------|
| `invalid_auth` | Invalid or expired token |
| `not_authed` | No authentication token provided |
| `channel_not_found` | Channel ID doesn't exist or bot lacks access |
| `not_in_channel` | Bot is not a member of the channel |
| `no_permission` | Bot lacks required scope |
| `ratelimited` | Rate limit exceeded — retry after pause |
| `user_not_found` | User ID doesn't exist |

## Rate Limits

Slack applies tiered rate limits per method:

| Tier | Requests per minute |
|------|---------------------|
| Tier 1 | 1 |
| Tier 2 | 20 |
| Tier 3 | 50 |
| Tier 4 | 100+ |

Most messaging methods are Tier 3. Channel/user listing is Tier 2. See [Slack Rate Limits](https://api.slack.com/docs/rate-limits) for details.

## Resources

- [Slack Web API Reference](https://api.slack.com/web)
- [chat.postMessage](https://api.slack.com/methods/chat.postMessage)
- [conversations.list](https://api.slack.com/methods/conversations.list)
- [conversations.history](https://api.slack.com/methods/conversations.history)
- [users.list](https://api.slack.com/methods/users.list)
- [Bot Permissions Guide](https://api.slack.com/start/quickstart)
# SendGrid Emblem

Send transactional emails and manage dynamic templates through the [SendGrid Email API](https://docs.sendgrid.com/api-reference) — all from the Elysium CLI.

## Overview

The SendGrid emblem provides access to the core SendGrid v3 API, covering:

| Action | Method | Endpoint | Description |
|---|---|---|---|
| `send-email` | POST | `/mail/send` | Send a transactional email |
| `list-emails` | GET | `/messages` | List sent email activity |
| `get-email` | GET | `/messages/{msg_id}` | Get details of a specific email |
| `create-template` | POST | `/templates` | Create a dynamic email template |
| `list-templates` | GET | `/templates` | List all email templates |

**Base URL:** `https://api.sendgrid.com/v3`  
**Auth:** Bearer token (`Authorization: Bearer <key>`)  
**Category:** communication

---

## Setup

### 1. Get a SendGrid API Key

1. Log in to [app.sendgrid.com](https://app.sendgrid.com)
2. Navigate to **Settings → API Keys**
3. Click **Create API Key**
4. Choose **Full Access** (or restrict to *Mail Send* and *Template Engine* as needed)
5. Copy the generated key — it is only shown once

### 2. Set the Environment Variable

```bash
export SENDGRID_API_KEY="SG.xxxxxxxxxxxxxxxxxxxx"
```

Add the line to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to persist it across sessions.

### 3. Verify the Setup

```bash
ely execute sendgrid list-templates
```

A successful response returns an array of templates (empty `[]` if none exist yet).

---

## Usage Examples

### Send an Email

Send a basic transactional email with plain-text and HTML content:

```bash
ely execute sendgrid send-email \
  --body '{
    "personalizations": [
      {
        "to": [{ "email": "recipient@example.com", "name": "Jane Doe" }]
      }
    ],
    "from": { "email": "sender@yourdomain.com", "name": "Your App" },
    "subject": "Welcome to Our Platform!",
    "content": [
      { "type": "text/plain", "value": "Hi Jane, welcome aboard!" },
      { "type": "text/html",  "value": "<p>Hi Jane, <strong>welcome aboard!</strong></p>" }
    ]
  }'
```

**Send using a dynamic template** (handlebars substitution):

```bash
ely execute sendgrid send-email \
  --body '{
    "personalizations": [
      {
        "to": [{ "email": "recipient@example.com" }],
        "dynamic_template_data": {
          "first_name": "Jane",
          "confirm_url": "https://example.com/confirm/abc123"
        }
      }
    ],
    "from": { "email": "no-reply@yourdomain.com", "name": "Your App" },
    "subject": "Confirm your email",
    "template_id": "d-0123456789abcdef0123456789abcdef"
  }'
```

A `202 Accepted` response indicates the email was queued for delivery.

---

### List Sent Emails

Retrieve email activity using an SGQL filter expression:

```bash
# Filter by recipient address
ely execute sendgrid list-emails \
  --query "to_email=recipient@example.com"

# Filter by delivery status
ely execute sendgrid list-emails \
  --query "status=delivered" \
  --limit 25

# Filter by subject (contains)
ely execute sendgrid list-emails \
  --query "subject=Welcome" \
  --limit 50 \
  --offset 0
```

> **Note:** Email activity is available for 7 days on free plans. Upgrade to the *Email Activity Feed add-on* for up to 30 days of history.

---

### Get Email Details

Retrieve full details for a single sent message by its ID:

```bash
ely execute sendgrid get-email \
  --msg_id "aBCDefghIJKLmnOP"
```

The response includes delivery status, open/click counts, and a full event timeline.

---

### Create a Dynamic Template

Create a new dynamic template using handlebars syntax:

```bash
ely execute sendgrid create-template \
  --body '{
    "name": "Welcome Email",
    "generation": "dynamic"
  }'
```

After creation, add a version to the template in the [SendGrid dashboard](https://mc.sendgrid.com/dynamic-templates) or via the Template Versions API. Use the returned `id` as the `template_id` in `send-email`.

---

### List Templates

Retrieve all templates in your account:

```bash
# List dynamic templates (default)
ely execute sendgrid list-templates

# List both dynamic and legacy templates
ely execute sendgrid list-templates \
  --generations "dynamic,legacy"

# Paginate results
ely execute sendgrid list-templates \
  --page_size 5 \
  --page_token "YOUR_PAGE_TOKEN"
```

---

## Data Types

### `Email`
A sent email activity record returned by `list-emails` and `get-email`.

| Field | Type | Description |
|---|---|---|
| `id` | string | Unique message identifier |
| `from_email` | string | Sender email address |
| `to_email` | string | Primary recipient email address |
| `subject` | string | Email subject line |
| `status` | string | Delivery status (`delivered`, `not_delivered`, `processing`, `bounce`) |
| `opens_count` | integer | Number of times the email was opened |
| `clicks_count` | integer | Number of times links were clicked |
| `last_event_time` | string | ISO 8601 timestamp of the most recent event |

### `EmailAddress`
An email address with an optional display name.

| Field | Type | Required | Description |
|---|---|---|---|
| `email` | string | ✅ | The email address |
| `name` | string | | Display name |

### `Template`
A SendGrid email template returned by `create-template` and `list-templates`.

| Field | Type | Description |
|---|---|---|
| `id` | string | Unique template identifier |
| `name` | string | Human-readable template name |
| `generation` | string | Template type: `dynamic` or `legacy` |
| `updated_at` | string | ISO 8601 timestamp of last update |

### `TemplateVersion`
A specific version of a template.

| Field | Type | Description |
|---|---|---|
| `id` | string | Version identifier |
| `template_id` | string | Parent template ID |
| `active` | integer | `1` = active, `0` = inactive |
| `name` | string | Version name |
| `subject` | string | Subject line (supports `{{handlebars}}`) |
| `html_content` | string | HTML email body |
| `plain_content` | string | Plain-text email body |

---

## Error Reference

| Code | Error | Description |
|---|---|---|
| `400` | `bad_request` | Invalid request body or missing required fields |
| `401` | `unauthorized` | Missing or invalid API key |
| `403` | `forbidden` | Insufficient permissions for this action |
| `404` | `not_found` | The requested resource does not exist |
| `429` | `rate_limit_exceeded` | Rate limit exceeded — retry after the indicated delay |
| `500` | `internal_error` | Unexpected error on the SendGrid platform |

---

## Resources

- [SendGrid API Reference](https://docs.sendgrid.com/api-reference)
- [Dynamic Templates Guide](https://docs.sendgrid.com/ui/sending-email/how-to-send-an-email-with-dynamic-templates)
- [Email Activity Feed](https://docs.sendgrid.com/ui/analytics-and-reporting/email-activity-feed)
- [SGQL Query Language](https://docs.sendgrid.com/for-developers/sending-email/getting-started-email-activity-api#filter-parameters)

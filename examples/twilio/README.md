# Twilio Emblem

Send SMS messages and verify phone numbers using the [Twilio Messaging](https://www.twilio.com/docs/sms) and [Twilio Verify](https://www.twilio.com/docs/verify) APIs with Elysium.

## Setup

Export your Twilio credentials (Account SID and Auth Token from the [Twilio Console](https://www.twilio.com/console)):

```bash
export TWILIO_CREDENTIALS="your_account_sid:your_auth_token"
```

Format: `<AccountSID>:<AuthToken>` (both values from your Twilio Console Dashboard).

## Actions

### `send-sms` — Send an SMS message

Send a text message to a phone number.

```bash
ely execute twilio send-sms \
  --AccountSid ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --To +15551234567 \
  --From +15559876543 \
  --Body "Hello from Elysium!"
```

| Parameter | Description |
|-----------|-------------|
| `AccountSid` | Your Twilio Account SID (required, begins with `AC`) |
| `To` | Destination phone number in E.164 format (e.g., `+15551234567`) |
| `From` | Your Twilio phone number or Messaging Service SID |
| `Body` | Message text (max 1600 characters) |

---

### `list-messages` — List messages

Retrieve messages with optional filters.

```bash
# All recent messages
ely execute twilio list-messages --AccountSid ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Messages sent to a specific number
ely execute twilio list-messages \
  --AccountSid ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --To +15551234567

# Messages from a specific sender on a date
ely execute twilio list-messages \
  --AccountSid ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --From +15559876543 \
  --DateSent 2024-01-15

# Pagination
ely execute twilio list-messages \
  --AccountSid ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --PageSize 100 \
  --Page 1
```

| Parameter | Description |
|-----------|-------------|
| `AccountSid` | Your Twilio Account SID (required) |
| `To` | Filter by recipient phone number |
| `From` | Filter by sender phone number |
| `DateSent` | Filter by date in `YYYY-MM-DD` format |
| `PageSize` | Results per page (1–1000, default 50) |
| `Page` | Zero-indexed page number |

---

### `get-message` — Get message details

Retrieve details of a specific message by its SID.

```bash
ely execute twilio get-message \
  --AccountSid ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --MessageSid SMxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

| Parameter | Description |
|-----------|-------------|
| `AccountSid` | Your Twilio Account SID (required) |
| `MessageSid` | Message SID to retrieve (begins with `SM` or `MM`) |

---

### `verify-start` — Start phone verification

Send a verification code to a phone number. **Note:** Uses base URL `https://verify.twilio.com/v2`.

```bash
# Send SMS verification code
ely execute twilio verify-start \
  --ServiceSid VAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --To +15551234567 \
  --Channel sms

# Send verification via voice call
ely execute twilio verify-start \
  --ServiceSid VAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --To +15551234567 \
  --Channel call
```

| Parameter | Description |
|-----------|-------------|
| `ServiceSid` | Your Verify Service SID (required, begins with `VA`) |
| `To` | Phone number to verify in E.164 format |
| `Channel` | Verification channel: `sms`, `call`, or `email` |

---

### `verify-check` — Check verification code

Validate an OTP code entered by the user. **Note:** Uses base URL `https://verify.twilio.com/v2`.

```bash
ely execute twilio verify-check \
  --ServiceSid VAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --To +15551234567 \
  --Code 123456
```

| Parameter | Description |
|-----------|-------------|
| `ServiceSid` | Your Verify Service SID (required, begins with `VA`) |
| `To` | Phone number being verified in E.164 format |
| `Code` | OTP code entered by the user |

Returns `valid: true` if the code is correct.

## Common Workflows

### Send SMS and check delivery status

```bash
# 1. Send the message
ely execute twilio send-sms \
  --AccountSid ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --To +15551234567 \
  --From +15559876543 \
  --Body "Your verification code is 123456"

# 2. Check delivery status using the returned MessageSid
ely execute twilio get-message \
  --AccountSid ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --MessageSid SMxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### Phone verification flow

```bash
# 1. Start verification - sends OTP to user
ely execute twilio verify-start \
  --ServiceSid VAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --To +15551234567 \
  --Channel sms

# 2. User receives code via SMS, then verify it
ely execute twilio verify-check \
  --ServiceSid VAxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  --To +15551234567 \
  --Code 123456
```

## Message Status Values

| Status | Meaning |
|--------|---------|
| `queued` | Message queued for sending |
| `sending` | Currently being sent |
| `sent` | Successfully sent to carrier |
| `delivered` | Delivered to handset |
| `undelivered` | Failed to deliver |
| `failed` | Delivery failed |

## Authentication

This emblem uses **HTTP Basic Authentication**. The `TWILIO_CREDENTIALS` environment variable should contain:

```
<AccountSID>:<AuthToken>
```

Find these values in your [Twilio Console Dashboard](https://www.twilio.com/console).

## Resources

- [Twilio Messaging API](https://www.twilio.com/docs/sms/api)
- [Twilio Verify API](https://www.twilio.com/docs/verify/api)
- [Phone Number Format (E.164)](https://www.twilio.com/docs/glossary/what-e164)
- [Twilio Console](https://www.twilio.com/console)
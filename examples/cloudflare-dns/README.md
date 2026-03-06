# Cloudflare DNS Emblem

Manage Cloudflare DNS records and CDN cache across your zones through the [Cloudflare API](https://developers.cloudflare.com/api/) using Elysium.

## Setup

Export your Cloudflare API token (create one at **My Profile → API Tokens** on dash.cloudflare.com):

```bash
export CLOUDFLARE_API_TOKEN=your_api_token_here
```

The token needs permissions for:
- **Zone → Zone → Read** (to list zones)
- **Zone → DNS → Edit** (to manage DNS records)
- **Zone → Cache Purge → Purge** (to purge CDN cache)

## Actions

### `list-zones` — List all zones

List all zones in your Cloudflare account with optional filters.

```bash
# All zones
ely execute cloudflare-dns list-zones

# Filter by domain name
ely execute cloudflare-dns list-zones --name example.com

# Filter by status
ely execute cloudflare-dns list-zones --status active

# Pagination
ely execute cloudflare-dns list-zones --page 1 --per_page 50
```

---

### `list-dns-records` — List DNS records

List DNS records for a zone with optional filters.

```bash
# All records for a zone
ely execute cloudflare-dns list-dns-records --zone_id abc123xyz

# Filter by record type
ely execute cloudflare-dns list-dns-records --zone_id abc123xyz --type A

# Filter by name
ely execute cloudflare-dns list-dns-records --zone_id abc123xyz --name www.example.com

# Filter by content
ely execute cloudflare-dns list-dns-records --zone_id abc123xyz --content 192.0.2.1

# Pagination
ely execute cloudflare-dns list-dns-records --zone_id abc123xyz --page 1 --per_page 100
```

---

### `create-dns-record` — Create a DNS record

Create a new DNS record in a zone.

```bash
# Create an A record
ely execute cloudflare-dns create-dns-record \
  --zone_id abc123xyz \
  --body '{"type": "A", "name": "www", "content": "192.0.2.1", "ttl": 3600}'

# Create a proxied A record (orange cloud)
ely execute cloudflare-dns create-dns-record \
  --zone_id abc123xyz \
  --body '{"type": "A", "name": "app", "content": "192.0.2.1", "proxied": true}'

# Create a CNAME record
ely execute cloudflare-dns create-dns-record \
  --zone_id abc123xyz \
  --body '{"type": "CNAME", "name": "blog", "content": "example.github.io", "ttl": 1}'

# Create a TXT record
ely execute cloudflare-dns create-dns-record \
  --zone_id abc123xyz \
  --body '{"type": "TXT", "name": "@", "content": "v=spf1 include:_spf.google.com ~all"}'

# Create an MX record
ely execute cloudns-dns create-dns-record \
  --zone_id abc123xyz \
  --body '{"type": "MX", "name": "@", "content": "mail.example.com", "ttl": 3600}'
```

Use `@` for the zone apex (root domain). Set `ttl: 1` for automatic TTL.

---

### `update-dns-record` — Update a DNS record

Update an existing DNS record in a zone.

```bash
# Update the IP address of an A record
ely execute cloudflare-dns update-dns-record \
  --zone_id abc123xyz \
  --dns_record_id def456uvw \
  --body '{"type": "A", "name": "www", "content": "192.0.2.50", "ttl": 3600}'

# Enable proxying on an existing record
ely execute cloudflare-dns update-dns-record \
  --zone_id abc123xyz \
  --dns_record_id def456uvw \
  --body '{"type": "A", "name": "www", "content": "192.0.2.50", "proxied": true}'

# Update with a comment
ely execute cloudflare-dns update-dns-record \
  --zone_id abc123xyz \
  --dns_record_id def456uvw \
  --body '{"type": "A", "name": "www", "content": "192.0.2.50", "comment": "Updated IP for new server"}'
```

---

### `purge-cache` — Purge CDN cache

Purge cached content from Cloudflare's CDN for a zone.

```bash
# Purge everything (use with caution)
ely execute cloudflare-dns purge-cache \
  --zone_id abc123xyz \
  --body '{"purge_everything": true}'

# Purge specific files
ely execute cloudflare-dns purge-cache \
  --zone_id abc123xyz \
  --body '{"files": ["https://example.com/style.css", "https://example.com/app.js"]}'

# Purge by cache tags
ely execute cloudflare-dns purge-cache \
  --zone_id abc123xyz \
  --body '{"tags": ["product-images", "homepage"]}'

# Purge by hostname
ely execute cloudflare-dns purge-cache \
  --zone_id abc123xyz \
  --body '{"hosts": ["cdn.example.com", "assets.example.com"]}'
```

## Common Workflows

### Find your zone ID

```bash
ely execute cloudflare-dns list-zones --name example.com
```

The response returns the zone `id` needed for other actions.

---

### Update DNS after server migration

```bash
# 1. List current DNS records
ely execute cloudflare-dns list-dns-records --zone_id abc123xyz --type A

# 2. Update the A record with new IP
ely execute cloudflare-dns update-dns-record \
  --zone_id abc123xyz \
  --dns_record_id def456uvw \
  --body '{"type": "A", "name": "@", "content": "192.0.2.100", "proxied": true}'

# 3. Purge cache so changes take effect immediately
ely execute cloudflare-dns purge-cache \
  --zone_id abc123xyz \
  --body '{"purge_everything": true}'
```

---

### Audit all TXT records for SPF/DKIM

```bash
ely execute cloudflare-dns list-dns-records --zone_id abc123xyz --type TXT
```

---

### Create multiple DNS records for a new service

```bash
# A record for the server
ely execute cloudflare-dns create-dns-record \
  --zone_id abc123xyz \
  --body '{"type": "A", "name": "api", "content": "192.0.2.10", "proxied": true}'

# CNAME for www subdomain
ely execute cloudflare-dns create-dns-record \
  --zone_id abc123xyz \
  --body '{"type": "CNAME", "name": "www.api", "content": "api.example.com", "proxied": true}'
```

## Authentication

This emblem uses a **Bearer token** stored in `CLOUDFLARE_API_TOKEN`. Create a token at:  
<https://dash.cloudflare.com/profile/api-tokens>

Recommended token permissions:
- **Zone → Zone → Read**
- **Zone → DNS → Edit**
- **Zone → Cache Purge → Purge**

For zone-specific tokens, limit the token scope to specific zones rather than all zones.

## DNS Record Types

| Type | Description |
|------|-------------|
| `A` | Maps a hostname to an IPv4 address |
| `AAAA` | Maps a hostname to an IPv6 address |
| `CNAME` | Maps a hostname to another hostname |
| `TXT` | Text record (SPF, DKIM, verification) |
| `MX` | Mail exchange record |
| `NS` | Nameserver record |
| `SRV` | Service record |
| `CAA` | Certificate Authority Authorization |
| `PTR` | Reverse DNS record |

## Error Handling

| Code | Description |
|------|-------------|
| 400 | Invalid request parameters |
| 401 | Missing or invalid authentication token |
| 404 | Zone or DNS record not found |
| 429 | Rate limit exceeded |

## Resources

- [Cloudflare API Reference](https://developers.cloudflare.com/api/)
- [DNS Records API](https://developers.cloudflare.com/api/operations/dns-records-for-a-zone-list-dns-records)
- [Zone API](https://developers.cloudflare.com/api/operations/zones-get-zones)
- [Cache Purge API](https://developers.cloudflare.com/api/operations/zone-purge)
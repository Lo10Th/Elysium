# Mapbox Emblem

Mapbox APIs for forward and reverse geocoding, turn-by-turn directions, and place search powered by precise location data.

## Setup

Export your Mapbox public access token:

```bash
export MAPBOX_ACCESS_TOKEN=your_mapbox_token_here
```

Get a token at: <https://account.mapbox.com/access-tokens/>

## Actions

### `geocode` — Forward geocode an address

Convert a text address or place name to geographic coordinates.

```bash
# Geocode an address
ely execute mapbox geocode \
  --search_text "1600 Pennsylvania Ave NW, Washington DC" \
  --access_token $MAPBOX_ACCESS_TOKEN

# Bias results toward a location
ely execute mapbox geocode \
  --search_text "coffee shop" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --proximity "-77.0369,38.9072"

# Restrict to specific countries
ely execute mapbox geocode \
  --search_text "Paris" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --country "us,ca"

# Filter by feature type
ely execute mapbox geocode \
  --search_text "San Francisco" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --types "place,region,country"

# Limit results and set language
ely execute mapbox geocode \
  --search_text "Berlin" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --limit 5 \
  --language de
```

---

### `reverse-geocode` — Reverse geocode coordinates

Convert a longitude and latitude coordinate pair to a human-readable address.

```bash
# Basic reverse geocode
ely execute mapbox reverse-geocode \
  --longitude -77.0369 \
  --latitude 38.9072 \
  --access_token $MAPBOX_ACCESS_TOKEN

# Get more results
ely execute mapbox reverse-geocode \
  --longitude -73.9857 \
  --latitude 40.7484 \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --limit 5

# Filter by feature type
ely execute mapbox reverse-geocode \
  --longitude -118.2437 \
  --latitude 34.0522 \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --types "address,place"

# Set response language
ely execute mapbox reverse-geocode \
  --longitude 2.3522 \
  --latitude 48.8566 \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --language fr
```

---

### `directions` — Get turn-by-turn directions

Get routing between two or more waypoints using a specified travel profile.

```bash
# Driving directions (default)
ely execute mapbox directions \
  --profile driving \
  --coordinates "-77.0369,38.9072;-73.9857,40.7484" \
  --access_token $MAPBOX_ACCESS_TOKEN

# Walking directions with steps
ely execute mapbox directions \
  --profile walking \
  --coordinates "-122.4194,37.7749;-122.4066,37.7879" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --steps true \
  --language en

# Cycling route with alternatives
ely execute mapbox directions \
  --profile cycling \
  --coordinates "-0.1278,51.5074;-0.0754,51.5187" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --alternatives true \
  --geometries geojson

# Driving with traffic, excluding tolls
ely execute mapbox directions \
  --profile driving-traffic \
  --coordinates "-118.2437,34.0522;-117.1611,32.7157" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --exclude "toll,motorway"

# Multi-waypoint route
ely execute mapbox directions \
  --profile driving \
  --coordinates "-77.0369,38.9072;-76.6123,39.2904;-75.1652,39.9526" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --steps true \
  --overview full
```

Available profiles:

| Profile | Description |
|---------|-------------|
| `driving` | Car routing via highways and streets |
| `driving-traffic` | Car routing with real-time traffic |
| `walking` | Pedestrian routing via paths and sidewalks |
| `cycling` | Bicycle routing via bike lanes and streets |

---

### `search` — Search for places with autocomplete

Find places and points of interest with autocomplete suggestions.

```bash
# Basic search
ely execute mapbox search \
  --q "pizza" \
  --access_token $MAPBOX_ACCESS_TOKEN

# Bias toward a location
ely execute mapbox search \
  --q "coffee" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --proximity "-122.4194,37.7749"

# Restrict to country with limit
ely execute mapbox search \
  --q "museum" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --country "fr,de,it" \
  --limit 10

# Filter by type
ely execute mapbox search \
  --q "restaurant" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --types "poi,address"

# Set origin for distance calculations
ely execute mapbox search \
  --q "hotel" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --origin "-0.1278,51.5074" \
  --proximity "-0.1278,51.5074"

# With session token for billing
ely execute mapbox search \
  --q "cafe" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --session_token "$(uuidgen)"
```

## Common Workflows

### Geocode an address and get directions

```bash
# 1. Geocode the destination
ely execute mapbox geocode \
  --search_text "Times Square, New York" \
  --access_token $MAPBOX_ACCESS_TOKEN

# 2. Use the returned coordinates to get directions
ely execute mapbox directions \
  --profile walking \
  --coordinates "-74.0060,40.7128;-73.9855,40.7580" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --steps true
```

### Find nearby POIs by type

```bash
# Search for restaurants near a location
ely execute mapbox search \
  --q "restaurant" \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --proximity "-73.9857,40.7484" \
  --types "poi" \
  --limit 10
```

### Reverse geocode coordinates from GPS

```bash
# Convert GPS coordinates to an address
ely execute mapbox reverse-geocode \
  --longitude -122.4194 \
  --latitude 37.7749 \
  --access_token $MAPBOX_ACCESS_TOKEN \
  --language en
```

## Authentication

This emblem uses a **public access token** passed via the `access_token` query parameter.

To obtain a token:
1. Sign in at <https://account.mapbox.com/>
2. Navigate to **Tokens** in your dashboard
3. Use a **public token** or create a new one

For production apps, create tokens with appropriate scopes:
- `geocode` requires `geocoding` scope
- `directions` requires `directions` scope
- `search` requires `search` scope

Tokens can be restricted by domain, URL, or authorization type for security.

## Rate Limits

Mapbox has rate limits based on your plan. Common responses:

| Code | Description |
|------|-------------|
| 429 | Rate limit exceeded — check `Retry-After` header for backoff time |

Free tier: 100,000 requests/month for geocoding. Check your plan at <https://account.mapbox.com/>.

## Resources

- [Mapbox API Documentation](https://docs.mapbox.com/api/)
- [Geocoding API](https://docs.mapbox.com/api/search/geocoding/)
- [Directions API](https://docs.mapbox.com/api/navigation/directions/)
- [Search API](https://docs.mapbox.com/api/search/search-box/)
- [Access Tokens Guide](https://docs.mapbox.com/accounts/guides/tokens/)
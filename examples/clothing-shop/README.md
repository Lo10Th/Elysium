# Clothing Shop Emblem

REST API for an online clothing store with product management and order processing.

## Setup

### 1. Start the Backend Server

```bash
cd ../clothing_shop
python -m venv env
source env/bin/activate
pip install -r requirements.txt
python app.py
```

The API runs at `http://localhost:5000`.

### 2. Generate an API Key

```bash
curl -X POST http://localhost:5000/api/auth/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-key"}'
```

Copy the returned API key.

### 3. Configure Environment

```bash
export CLOTHING_SHOP_API_KEY=your-api-key-here
```

### 4. Pull the Emblem

```bash
ely pull clothing-shop
```

## Available Actions

| Action | Method | Path | Description |
|--------|--------|------|-------------|
| `list-products` | GET | `/products` | List all products with optional filtering |
| `get-product` | GET | `/products/{id}` | Retrieve a single product by ID |
| `create-product` | POST | `/products` | Create a new product |
| `update-product` | PUT | `/products/{id}` | Update an existing product |
| `delete-product` | DELETE | `/products/{id}` | Delete a product |
| `list-orders` | GET | `/orders` | List all orders with optional filtering |
| `get-order` | GET | `/orders/{id}` | Retrieve a single order by ID |
| `create-order` | POST | `/orders` | Place a new order for products |
| `update-order-status` | PUT | `/orders/{id}/status` | Update the status of an order |

## Examples

### List Products

```bash
# All products
ely execute clothing-shop list-products

# Filter by category
ely execute clothing-shop list-products --category "shirts"
```

### Get a Product

```bash
ely execute clothing-shop get-product --id 1
```

### Create a Product

```bash
ely execute clothing-shop create-product --body '{
  "name": "Cotton T-Shirt",
  "description": "Comfortable 100% cotton t-shirt",
  "price": 29.99,
  "size": "M",
  "color": "blue",
  "category": "shirts",
  "stock": 100,
  "image_url": "https://example.com/tshirt.jpg"
}'
```

### Update a Product

```bash
ely execute clothing-shop update-product --id 1 --body '{
  "price": 24.99,
  "stock": 85
}'
```

### Delete a Product

```bash
ely execute clothing-shop delete-product --id 1
```

### List Orders

```bash
# All orders
ely execute clothing-shop list-orders

# Filter by status
ely execute clothing-shop list-orders --status pending
```

### Get an Order

```bash
ely execute clothing-shop get-order --id 1
```

### Create an Order

```bash
ely execute clothing-shop create-order --body '{
  "customer_name": "John Doe",
  "customer_email": "john@example.com",
  "customer_address": "123 Main St, Anytown, USA",
  "items": [
    {"product_id": 1, "quantity": 2, "size": "M", "color": "blue"},
    {"product_id": 5, "quantity": 1, "size": "L"}
  ]
}'
```

### Update Order Status

```bash
ely execute clothing-shop update-order-status --id 1 --body '{
  "status": "shipped"
}'
```

## Product Sizes

| Size | Code |
|------|------|
| Small | `S` |
| Medium | `M` |
| Large | `L` |
| Extra Large | `XL` |

## Order Statuses

| Status | Description |
|--------|-------------|
| `pending` | Order placed, awaiting processing |
| `processing` | Order is being prepared |
| `shipped` | Order has been shipped |
| `delivered` | Order has been delivered |
| `cancelled` | Order was cancelled |

## Authentication

This emblem uses an **API key** stored in `CLOTHING_SHOP_API_KEY`. The key is sent in the `X-API-Key` header.

Generate a key using the auth endpoint:

```bash
curl -X POST http://localhost:5000/api/auth/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "my-key-name"}'
```

## Error Handling

| Code | Error | Description |
|------|-------|-------------|
| 400 | `bad_request` | Invalid request parameters or missing required fields |
| 401 | `unauthorized` | Valid API key required |
| 404 | `not_found` | The requested resource does not exist |
| 500 | `internal_error` | An unexpected error occurred on the server |

## Resources

- Backend API: `../clothing_shop/app.py`
- Emblem Specification: `../../docs/EMBLEM_SPEC.md`
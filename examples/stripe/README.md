# Stripe Emblem

Accept payments and manage customers with the [Stripe REST API](https://stripe.com/docs/api) using Elysium.

## Setup

Export your Stripe secret API key (create one at **Developers → API Keys** on dashboard.stripe.com):

```bash
export STRIPE_API_KEY=sk_test_your_stripe_secret_key_here
```

## Actions

### `list-customers` — List all customers

Retrieve a paginated list of customers, most recent first.

```bash
# List recent customers
ely execute stripe list-customers

# Limit to 25 customers
ely execute stripe list-customers --limit 25

# Filter by email
ely execute stripe list-customers --email customer@example.com
```

---

### `get-customer` — Retrieve a customer

Fetch full details for a single customer by ID.

```bash
ely execute stripe get-customer --id cus_abc123xyz
```

---

### `create-customer` — Create a new customer

Create a new Stripe customer object.

```bash
# Minimal customer (email only)
ely execute stripe create-customer --email customer@example.com

# With name and description
ely execute stripe create-customer \
  --email customer@example.com \
  --name "Jane Doe" \
  --description "VIP customer"
```

---

### `delete-customer` — Delete a customer

Permanently delete a customer and cancel any active subscriptions.

```bash
ely execute stripe delete-customer --id cus_abc123xyz
```

---

### `list-payment-intents` — List all PaymentIntents

Retrieve a paginated list of PaymentIntents, most recent first.

```bash
# List recent PaymentIntents
ely execute stripe list-payment-intents

# Limit to 20 results
ely execute stripe list-payment-intents --limit 20

# Filter by customer
ely execute stripe list-payment-intents --customer cus_abc123xyz
```

---

### `create-payment-intent` — Create a PaymentIntent

Create a PaymentIntent to collect a payment from your customer.

```bash
# Create a $25.00 USD payment
ely execute stripe create-payment-intent \
  --amount 2500 \
  --currency usd

# Payment associated with a customer
ely execute stripe create-payment-intent \
  --amount 2500 \
  --currency usd \
  --customer cus_abc123xyz \
  --description "Order #12345"
```

**Important**: Amount is in the smallest currency unit (e.g. cents for USD). 2500 = $25.00.

---

### `get-payment-intent` — Retrieve a PaymentIntent

Fetch details for a single PaymentIntent by ID.

```bash
ely execute stripe get-payment-intent --id pi_abc123xyz
```

---

### `confirm-payment-intent` — Confirm a PaymentIntent

Confirm a PaymentIntent to attempt payment collection.

```bash
ely execute stripe confirm-payment-intent \
  --id pi_abc123xyz \
  --payment_method pm_card_visa
```

---

### `cancel-payment-intent` — Cancel a PaymentIntent

Cancel a PaymentIntent before it reaches a terminal state (succeeded or canceled).

```bash
ely execute stripe cancel-payment-intent --id pi_abc123xyz
```

---

### `list-charges` — List all charges

Retrieve a paginated list of charges, most recent first.

```bash
# List recent charges
ely execute stripe list-charges

# Limit to 15 results
ely execute stripe list-charges --limit 15

# Filter by customer
ely execute stripe list-charges --customer cus_abc123xyz
```

---

### `get-charge` — Retrieve a charge

Fetch details for a single charge by ID.

```bash
ely execute stripe get-charge --id ch_abc123xyz
```

## Common Workflows

### Create customer and initiate payment

```bash
# 1. Create a customer
ely execute stripe create-customer \
  --email alice@example.com \
  --name "Alice Smith"

# 2. Create a PaymentIntent for that customer (use returned customer ID)
ely execute stripe create-payment-intent \
  --amount 9999 \
  --currency usd \
  --customer cus_<returned_id> \
  --description "Pro subscription"

# 3. Confirm with a payment method (use returned PaymentIntent ID)
ely execute stripe confirm-payment-intent \
  --id pi_<returned_id> \
  --payment_method pm_card_visa
```

### Verify payment status

```bash
# After creating a PaymentIntent, poll for status
ely execute stripe get-payment-intent --id pi_abc123xyz

# Status will be one of:
# - requires_payment_method
# - requires_confirmation
# - processing
# - succeeded
# - canceled
```

### View all charges for a customer

```bash
ely execute stripe list-charges --customer cus_abc123xyz
```

## Payment Intent Lifecycle

| Status | Meaning |
|--------|---------|
| `requires_payment_method` | No payment method attached |
| `requires_confirmation` | Ready to be confirmed |
| `processing` | Payment is being processed |
| `succeeded` | Payment completed successfully |
| `canceled` | Payment was canceled |

## Charge Status

| Status | Meaning |
|--------|---------|
| `succeeded` | Charge completed |
| `pending` | Charge is pending |
| `failed` | Charge failed |

## Authentication

This emblem uses a **Bearer token** stored in `STRIPE_API_KEY`. Create a secret key at:  
<https://dashboard.stripe.com/apikeys>

Use test mode keys (`sk_test_*`) during development. Never expose your secret key in client-side code.

## Error Handling

| Code | Error | Description |
|------|-------|-------------|
| 404 | `resource_missing` | Customer, PaymentIntent, or Charge not found |
| 402 | `card_declined` | Card was declined by the issuer |
| 400 | `invalid_request` | Invalid parameters provided |
| 401 | `authentication_error` | Invalid API key |

## Resources

- [Stripe API Reference](https://stripe.com/docs/api)
- [PaymentIntents Overview](https://stripe.com/docs/payments/payment-intents)
- [Customers API](https://stripe.com/docs/api/customers)
- [Charges API](https://stripe.com/docs/api/charges)
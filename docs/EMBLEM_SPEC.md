# Emblem Specification v1

This document describes the complete specification for defining an API emblem in Elysium.

## Overview

An **Emblem** is a YAML file that describes an API's endpoints, parameters, authentication, and types. It enables developers (and AI agents) to interact with APIs programmatically without reading extensive documentation.

## File Structure

Every emblem must be named `emblem.yaml` and placed in the emblem's root directory.

```yaml
apiVersion: v1
name: my-api
version: 1.0.0
description: Brief description of what this API does
author: Your Name
license: MIT

baseUrl: https://api.example.com/v1

auth:
  type: bearer
  keyEnv: MY_API_TOKEN

tags:
  - api
  - example

category: other

types:
  User:
    description: A user object
    properties:
      id:
        type: integer
        required: true
      name:
        type: string
        required: true
      email:
        type: string
        required: true

actions:
  list-users:
    description: List all users
    method: GET
    path: /users
    responses:
      200:
        description: Successfully retrieved users
        schema:
          type: array
          items:
            type: User
```

## Required Fields

### `apiVersion`

The version of the emblem specification. Currently only `v1` is supported.

```yaml
apiVersion: v1
```

### `name`

Unique identifier for the emblem. Must be lowercase, alphanumeric with hyphens allowed (but not at start/end).

```yaml
name: clothing-shop        # Valid
name: my_api              # Invalid: underscore not allowed
name: My-Shop             # Invalid: uppercase not allowed
name: -shop               # Invalid: cannot start with hyphen
```

Constraints:
- Length: 1-64 characters
- Pattern: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`

### `version`

Semantic version string following [SemVer 2.0.0](https://semver.org/).

```yaml
version: 1.0.0
version: 2.1.3
version: 0.1.0-beta
```

### `description`

Brief description of the API and its purpose. Should be concise but informative.

```yaml
description: REST API for an online clothing store with product management and order processing
```

Constraints:
- Length: 10-500 characters

### `baseUrl`

The base URL for all API requests. All paths in actions are relative to this URL.

```yaml
baseUrl: https://api.example.com/v1
baseUrl: http://localhost:5000/api
```

### `actions`

At least one action must be defined. See [Actions](#actions) section for details.

```yaml
actions:
  list-products:
    description: List all products
    method: GET
    path: /products
```

## Optional Fields

### `author`

Name of the author or organization maintaining the emblem.

```yaml
author: John Doe
author: Acme Corporation
```

### `license`

SPDX license identifier. Defaults to `MIT` if not specified.

```yaml
license: MIT
license: Apache-2.0
license: GPL-3.0
```

### `repository`

URL to the source code repository.

```yaml
repository: https://github.com/user/repo
```

### `homepage`

URL to documentation or homepage.

```yaml
homepage: https://docs.example.com
```

### `tags`

Keywords for search and categorization. Maximum 10 tags, each up to 50 characters.

```yaml
tags:
  - api
  - ecommerce
  - shop
  - rest
```

### `category`

Primary category for the API. Must be one of:

- `payments`
- `storage`
- `ai`
- `communication`
- `database`
- `infrastructure`
- `analytics`
- `ecommerce`
- `social`
- `productivity`
- `security`
- `media`
- `integration`
- `other`

```yaml
category: ecommerce
```

## Authentication

### `auth`

Defines how the API authenticates requests.

#### No Authentication

```yaml
auth:
  type: none
```

#### API Key

```yaml
auth:
  type: api_key
  keyEnv: MY_API_KEY           # Environment variable name
  header: X-API-Key            # Header name (optional, defaults to X-API-Key)
```

#### Bearer Token

```yaml
auth:
  type: bearer
  keyEnv: MY_API_TOKEN         # Environment variable name
  header: Authorization        # Optional, defaults to Authorization
  prefix: "Bearer "            # Optional, defaults to "Bearer "
```

#### Basic Authentication

```yaml
auth:
  type: basic
  keyEnv: MY_BASIC_CREDS       # Should contain "username:password"
```

#### OAuth 2.0 (Future)

```yaml
auth:
  type: oauth2
  flows:
    authorizationCode:
      authorizationUrl: https://example.com/oauth/authorize
      tokenUrl: https://example.com/oauth/token
      scopes:
        read: Read access
        write: Write access
```

**Note**: Environment variables are never stored in the emblem. Users must set them locally before using the emblem.

## Type Definitions

### `types`

Define reusable type schemas for documentation and type checking.

```yaml
types:
  Product:
    description: A product in the store
    properties:
      id:
        type: integer
        description: Unique product identifier
        required: true
      name:
        type: string
        description: Product name
        required: true
      price:
        type: number
        description: Price in USD
        required: true
      stock:
        type: integer
        description: Available inventory
        required: false
        default: 0

  Order:
    description: A customer order
    properties:
      id:
        type: integer
        required: true
      customer_name:
        type: string
        required: true
      items:
        type: array
        required: true
        items:
          type: OrderItem
        
  OrderItem:
    description: An item in an order
    properties:
      product_id:
        type: integer
        required: true
      quantity:
        type: integer
        required: true
      size:
        type: string
        required: false
      color:
        type: string
        required: false
```

### Property Types

Types can be:

- `string` - Text value
- `integer` - Whole number
- `number` - Floating point number
- `boolean` - True/false
- `array` - List of items (use `items` to define item type)
- `object` - Nested object (use `properties` to define fields)

```yaml
properties:
  count:
    type: integer
  price:
    type: number
  active:
    type: boolean
    default: true
  tags:
    type: array
    items:
      type: string
  metadata:
    type: object
    properties:
      key:
        type: string
      value:
        type: string
```

## Actions

Actions are the core of an emblem—they define what operations can be performed on the API.

### Action Structure

```yaml
actions:
  action-name:
    description: Brief description of what this action does
    method: GET
    path: /resource/{id}
    parameters:
      - name: id
        type: integer
        in: path
        required: true
        description: Resource ID
    responses:
      200:
        description: Success
        schema:
          type: Resource
```

### `description` (Required)

What this action does. Be clear and concise.

```yaml
description: Retrieve a single product by its ID
```

### `method` (Required)

HTTP method. Must be one of:

- `GET` - Retrieve data
- `POST` - Create data
- `PUT` - Replace data
- `DELETE` - Remove data
- `PATCH` - Partially update data

```yaml
method: GET
```

### `path` (Required)

URL path relative to `baseUrl`. Can contain path parameters in curly braces.

```yaml
path: /products
path: /products/{id}
path: /orders/{order_id}/items/{item_id}
```

### `parameters`

Input parameters for the action.

#### Parameter Location (`in`)

- `query` - URL query parameter (`?key=value`)
- `path` - URL path segment (`/users/{id}`)
- `header` - HTTP header
- `body` - Request body

#### Query Parameter

```yaml
parameters:
  - name: category
    type: string
    in: query
    required: false
    description: Filter products by category
    enum:
      - shirts
      - pants
      - shoes
```

#### Path Parameter

```yaml
parameters:
  - name: id
    type: integer
    in: path
    required: true
    description: Product ID
```

#### Header Parameter

```yaml
parameters:
  - name: X-Custom-Header
    type: string
    in: header
    required: false
    description: Custom header for special requests
```

#### Body Parameter

For simple bodies:

```yaml
parameters:
  - name: name
    type: string
    in: body
    required: true
  - name: price
    type: number
    in: body
    required: true
```

For complex bodies, use `requestBody`:

```yaml
requestBody:
  description: Order data
  contentType: application/json
  schema:
    type: object
    properties:
      customer_name:
        type: string
        required: true
      items:
        type: array
        required: true
        items:
          type: OrderItem
```

#### Parameter Properties

```yaml
- name: quantity
  type: integer
  in: body
  required: false
  default: 1
  description: Number of items
  enum:
    - 1
    - 5
    - 10
```

### `responses`

Define expected responses by status code.

```yaml
responses:
  200:
    description: Successfully retrieved product
    schema:
      type: Product
  404:
    description: Product not found
```

The `schema` can reference a type definition:

```yaml
schema:
  type: Product
```

Or define inline:

```yaml
schema:
  type: object
  properties:
    id:
      type: integer
    message:
      type: string
```

### `errors` (Optional)

Document known error conditions.

```yaml
errors:
  - code: 400
    error: bad_request
    message: Invalid parameters
    description: The request parameters are invalid or missing
  - code: 401
    error: unauthorized
    message: Authentication required
  - code: 404
    error: not_found
    message: Resource not found
  - code: 500
    error: internal_error
    message: Server error
```

## Action Naming Conventions

Use kebab-case (lowercase with hyphens) for action names. Follow these patterns:

| HTTP Method | Naming Pattern | Example |
|-------------|---------------|---------|
| GET (list) | `list-{resource}` | `list-products` |
| GET (single) | `get-{resource}` | `get-product` |
| POST | `create-{resource}` | `create-order` |
| PUT | `update-{resource}` | `update-product` |
| PATCH | `patch-{resource}` | `patch-product` |
| DELETE | `delete-{resource}` | `delete-product` |

For custom actions, use descriptive names:

```yaml
actions:
  send-invitation:
    description: Send an invitation email
    method: POST
    path: /invitations/send
    
  search-products:
    description: Search products by criteria
    method: GET
    path: /products/search
```

## Complete Example

```yaml
apiVersion: v1
name: clothing-shop
version: 1.0.0
description: REST API for an online clothing store with product management and order processing
author: Elysium Team
license: MIT
repository: https://github.com/elysium/clothing-shop
homepage: https://docs.elysium.dev/clothing-shop

baseUrl: https://api.clothing-shop.example.com/v1

auth:
  type: api_key
  keyEnv: CLOTHING_SHOP_API_KEY
  header: X-API-Key

tags:
  - api
  - ecommerce
  - shop
  - rest

category: ecommerce

types:
  Product:
    description: A product in the store
    properties:
      id:
        type: integer
        required: true
        description: Unique product identifier
      name:
        type: string
        required: true
        description: Product name
      description:
        type: string
        required: false
        description: Product description
      price:
        type: number
        required: true
        description: Price in USD
      size:
        type: string
        required: false
        description: Product size (S, M, L, XL)
        enum:
          - S
          - M
          - L
          - XL
      color:
        type: string
        required: false
        description: Product color
      category:
        type: string
        required: false
        description: Product category
      stock:
        type: integer
        required: false
        default: 0
        description: Available inventory
      created_at:
        type: string
        required: false
        description: Creation timestamp (ISO 8601)
      updated_at:
        type: string
        required: false
        description: Last update timestamp (ISO 8601)

  Order:
    description: A customer order
    properties:
      id:
        type: integer
        required: true
        description: Unique order identifier
      customer_name:
        type: string
        required: true
        description: Customer's full name
      customer_email:
        type: string
        required: true
        description: Customer's email address
      customer_address:
        type: string
        required: true
        description: Shipping address
      total_price:
        type: number
        required: true
        description: Total order price in USD
      status:
        type: string
        required: true
        description: Order status
        enum:
          - pending
          - processing
          - shipped
          - delivered
          - cancelled
      items:
        type: array
        required: true
        description: List of ordered items
        items:
          type: OrderItem
      created_at:
        type: string
        required: false
        description: Creation timestamp (ISO 8601)

  OrderItem:
    description: An item in an order
    properties:
      id:
        type: integer
        required: true
        description: Unique order item identifier
      product_id:
        type: integer
        required: true
        description: Product ID
      quantity:
        type: integer
        required: true
        description: Quantity ordered
      price:
        type: number
        required: true
        description: Price per unit
      size:
        type: string
        required: false
        description: Selected size
      color:
        type: string
        required: false
        description: Selected color

actions:
  list-products:
    description: List all products with optional filtering
    method: GET
    path: /products
    parameters:
      - name: category
        type: string
        in: query
        required: false
        description: Filter by category
      - name: size
        type: string
        in: query
        required: false
        description: Filter by size
        enum:
          - S
          - M
          - L
          - XL
      - name: color
        type: string
        in: query
        required: false
        description: Filter by color
    responses:
      200:
        description: List of products
        schema:
          type: array
          items:
            type: Product

  get-product:
    description: Retrieve a single product by ID
    method: GET
    path: /products/{id}
    parameters:
      - name: id
        type: integer
        in: path
        required: true
        description: Product ID
    responses:
      200:
        description: Product details
        schema:
          type: Product
      404:
        description: Product not found

  create-product:
    description: Create a new product
    method: POST
    path: /products
    requestBody:
      description: Product data
      schema:
        type: object
        properties:
          name:
            type: string
            required: true
          price:
            type: number
            required: true
          description:
            type: string
            required: false
          size:
            type: string
            required: false
          color:
            type: string
            required: false
          category:
            type: string
            required: false
          stock:
            type: integer
            required: false
            default: 0
    responses:
      201:
        description: Product created
        schema:
          type: Product
      400:
        description: Invalid product data

  update-product:
    description: Update an existing product
    method: PUT
    path: /products/{id}
    parameters:
      - name: id
        type: integer
        in: path
        required: true
        description: Product ID
    requestBody:
      description: Updated product data
      schema:
        type: Product
    responses:
      200:
        description: Product updated
        schema:
          type: Product
      404:
        description: Product not found

  delete-product:
    description: Delete a product
    method: DELETE
    path: /products/{id}
    parameters:
      - name: id
        type: integer
        in: path
        required: true
        description: Product ID
    responses:
      204:
        description: Product deleted
      404:
        description: Product not found

  list-orders:
    description: List all orders
    method: GET
    path: /orders
    parameters:
      - name: status
        type: string
        in: query
        required: false
        description: Filter by status
        enum:
          - pending
          - processing
          - shipped
          - delivered
          - cancelled
    responses:
      200:
        description: List of orders
        schema:
          type: array
          items:
            type: Order

  get-order:
    description: Retrieve a single order by ID
    method: GET
    path: /orders/{id}
    parameters:
      - name: id
        type: integer
        in: path
        required: true
        description: Order ID
    responses:
      200:
        description: Order details
        schema:
          type: Order
      404:
        description: Order not found

  create-order:
    description: Place a new order
    method: POST
    path: /orders
    requestBody:
      description: Order data
      schema:
        type: object
        properties:
          customer_name:
            type: string
            required: true
            description: Customer's full name
          customer_email:
            type: string
            required: true
            description: Customer's email
          customer_address:
            type: string
            required: true
            description: Shipping address
          items:
            type: array
            required: true
            description: Order items
            items:
              type: object
              properties:
                product_id:
                  type: integer
                  required: true
                quantity:
                  type: integer
                  required: true
                size:
                  type: string
                  required: false
                color:
                  type: string
                  required: false
    responses:
      201:
        description: Order created
        schema:
          type: Order
      400:
        description: Invalid order data

  update-order-status:
    description: Update order status
    method: PUT
    path: /orders/{id}/status
    parameters:
      - name: id
        type: integer
        in: path
        required: true
        description: Order ID
    requestBody:
      description: Status update
      schema:
        type: object
        properties:
          status:
            type: string
            required: true
            enum:
              - pending
              - processing
              - shipped
              - delivered
              - cancelled
    responses:
      200:
        description: Order status updated
        schema:
          type: Order
      404:
        description: Order not found

errors:
  - code: 400
    error: bad_request
    message: Invalid request parameters
    description: The request parameters are invalid or missing required fields
  - code: 401
    error: unauthorized
    message: Authentication required
    description: Valid API key required
  - code: 404
    error: not_found
    message: Resource not found
    description: The requested resource does not exist
  - code: 500
    error: internal_error
    message: Internal server error
    description: An unexpected error occurred
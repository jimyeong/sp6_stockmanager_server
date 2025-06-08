# Owlverload API

A Go-based API server for inventory management with authentication middleware.

## Setup

### Prerequisites

- Go 1.19 or later
- MySQL database

### Environment Variables

Create `.env.development` and `.env.production` files with the following variables:

```
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_HOST=your_db_host
DB_PORT=your_db_port
DB_NAME=your_db_name
ENV=development
```

### Database Setup

Run the schema.sql file to set up your database:

```sql
mysql -u your_user -p your_database < db/schema.sql
```

## API Endpoints

### Authentication

All endpoints except `/api/v1/auth/signin`, `/health`, and `/api/v1/health` require authentication via Bearer token in the Authorization header.

Example:
```
Authorization: Bearer your_token_here
```

### Public Endpoints

- `POST /api/v1/auth/signin` - User login
- `GET /health` - Server health check
- `GET /api/v1/health` - Server health check

### Protected Endpoints (Require Authentication)

#### Item Management

- `GET /api/v1/getItem?barcode={barcode}` - Get an item by barcode
- `GET /api/v1/getItems` - Get all items
- `POST /api/v1/createItem` - Create a new item
- `POST /api/v1/registerItem` - Register a new item (alias for createItem)

#### Stock Management

- `POST /api/v1/stockIn` - Add stock to an item
- `POST /api/v1/stockOut` - Remove stock from an item

## Request/Response Examples

### Get Item by Barcode

Request:
```
GET /api/v1/getItem?barcode=123456789
```

Response:
```json
{
  "message": "Item found",
  "payload": {
    "id": "item_1714289345678",
    "barcode": "123456789",
    "name": "Sample Item",
    "description": "This is a sample item",
    "category": "Test",
    "quantityInStock": 10,
    "unitPrice": 9.99,
    "lastUpdated": "2023-05-01T12:34:56Z",
    "creatorId": "user123",
    "createdAt": "2023-05-01T12:34:56Z"
  },
  "success": true
}
```

### Stock In

Request:
```json
POST /api/v1/stockIn
{
  "itemId": "item_1714289345678",
  "quantity": 5,
  "notes": "Received from supplier"
}
```

Response:
```json
{
  "message": "Stock in completed successfully",
  "payload": null,
  "success": true
}
```

### Stock Out

Request:
```json
POST /api/v1/stockOut
{
  "itemId": "item_1714289345678",
  "quantity": 2,
  "notes": "Customer order #1234"
}
```

Response:
```json
{
  "message": "Stock out completed successfully",
  "payload": null,
  "success": true
}
```

### Create Item

Request:
```json
POST /api/v1/createItem
{
  "barcode": "987654321",
  "name": "New Product",
  "description": "This is a new product",
  "category": "Electronics",
  "quantityInStock": 0,
  "unitPrice": 29.99
}
```

Response:
```json
{
  "message": "Item created successfully",
  "payload": {
    "id": "item_1714289399999",
    "barcode": "987654321",
    "name": "New Product",
    "description": "This is a new product",
    "category": "Electronics",
    "quantityInStock": 0,
    "unitPrice": 29.99,
    "lastUpdated": "2023-05-01T12:45:56Z",
    "creatorId": "user123",
    "createdAt": "2023-05-01T12:45:56Z"
  },
  "success": true
}
```

### Lookup Items

Endpoint: POST /api/v1/lookupItems

Request Format:
  {
    "search_type": "code" | "barcode" | "name",
    "value": "search_string"
  }

  Response Format:
  {
    "success": true,
    "message": "Items found successfully",
    "payload": {
      "items": [
        {
          "id": "item_123",
          "code": "PROD-001",
          "barcode": "1234567890123",
          "barcodeForBox": "1234567890124",
          "name": "Apple Juice 1L",
          "name_kor": "사과주스 1L",
          "name_eng": "Apple Juice 1L",
          "name_chi": "苹果汁 1升",
          "name_jap": "アップルジュース1L",
          "type": "SINGLE",
          "availableForOrder": true,
          "imagePath": "https://example.com/image.jpg",
          "ingredients": "Apple juice, natural flavoring",
          "isBeefContained": false,
          "isPorkContained": false,
          "isHalal": true,
          "reasoning": "No animal products, certified halal",
          "createdAt": "2024-01-01T00:00:00Z",
          "stock": []
        }
      ]
    }
  }
```

## Running the Server

```bash
go run main.go
```

The server will start on port 8080.





		payload := map[string]interface{}{
			"model": "gpt-4",
			"messages": []map[string]string{
				{"role": "user", "content": prompt},
			},
			"max_tokens": 300,
		}

		body, _ := json.Marshal(payload)
		reqGPT, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
		reqGPT.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
		reqGPT.Header.Set("Content-Type", "application/json")

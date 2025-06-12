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







```

8801043029360
8801043029353
8801043029353
8801043029353
18801045575534

645175570288
8801128503068
18801128505823
8801128503068
8801043031028

8801043029353
8801043029360
087703050853
087703294127
8801043055611

8801043055604
14580352290128
8801043035538
28809054401513
8801043053174

8801043952996
8801073221499
8801073223561
8801073127821
8801073127838

8801043952972
8801043031028
18801037054634
8801037054637
18801619800857

18809971320020
12000001368357
087703039469
18809090186903
18801007872497

8801024946907
38801024946908
8801043021883
8801043022002
8801043062848

8801043043441
8801073140622
8801043035477
8801045521077
8801128503051

8801043060370
645175521440
645175521556
8801073144446
8801073144453

8801073116863
8801073116849
8801073116832
8801045522562
8801045826981

8801045521329
8801073143777
8801073141858
8801073141896
8801073142961

8801073140578
8801073144491
8801073144484
8801073116467
8801073116474

8801073110502
8801073114531
8801073113428
8801073113381
8801073115514

8801043157766
8801043032049
8801045522555
645175522119
8801043157735

8801073115606
8801045522555
8801043032049
8801043157766
8801073115514

8801073113381
8801073113428
8801073114531
8801073110502
8801073116474

8801073116467
8801043150620
8801043070362
8801043069588
8801043018470

8801043060226
8801043053167
8801043022705
8801043157759
074603006356

074603005892
8801043157711
8801043157773
8801043157728

8801073115606
8801043157735
645175522119
8809054401519

8809054401533
8809054403506
8809054402097
8809054409034
8801005198462

8801005198455
8801392037078
8801392037092
8809054400987

8809054409652
8809054401670
8809054401687


```
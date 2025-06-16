# Create Item with Tags API Documentation

## Endpoint
**POST** `/api/v1/createNewItem`

## Description
Creates a new item in the database and associates it with the provided tags. The function will automatically handle tag associations during the item creation process.

## Authentication
Requires Firebase authentication token in the Authorization header:
```
Authorization: Bearer your_firebase_token
```

## Request Format
The request must be sent as JSON with the following fields:

### Required Fields
- `code`: Unique item code (string)
- `name`: Item name (string)

### Optional Fields
- `id`: Item ID (auto-generated if not provided)
- `barcode`: Product barcode (string)
- `box_barcode`: Box barcode (string)
- `name_jpn`: Japanese name (string)
- `name_chn`: Chinese name (string)
- `name_kor`: Korean name (string)
- `name_eng`: English name (string)
- `type`: Item type/category (string)
- `availableForOrder`: Availability flag (integer: 0 or 1)
- `imagePath`: Path to item image (string)
- `tag`: Array of tag objects with `id` and `tagName` fields

## Example Request

```json
{
  "code": "ABC123",
  "barcode": "1234567890123",
  "box_barcode": "BOX1234567890123",
  "name": "Sample Product",
  "name_jpn": "サンプル商品",
  "name_chn": "样品产品",
  "name_kor": "샘플 제품",
  "name_eng": "Sample Product",
  "type": "Electronics",
  "availableForOrder": 1,
  "imagePath": "/550e8400-e29b-41d4-a716-446655440000.jpg",
  "tag": [
    {
      "id": "tag_1",
      "tagName": "electronics"
    },
    {
      "id": "tag_2", 
      "tagName": "gadget"
    },
    {
      "id": "sp6",
      "tagName": "special-promotion"
    }
  ]
}
```

## Example cURL Request

```bash
curl -X POST \
  -H "Authorization: Bearer YOUR_FIREBASE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "ABC123",
    "name": "Sample Product",
    "type": "Electronics",
    "availableForOrder": 1,
    "imagePath": "/550e8400-e29b-41d4-a716-446655440000.jpg",
    "tag": [
      {
        "id": "tag_1",
        "tagName": "electronics"
      },
      {
        "id": "sp6",
        "tagName": "special-promotion"
      }
    ]
  }' \
  http://localhost:8080/api/v1/createNewItem
```

## Example Response

```json
{
  "message": "Item created successfully",
  "payload": {
    "item": {
      "id": "item_1642867200000000000",
      "code": "ABC123",
      "barcode": "1234567890123",
      "box_barcode": "BOX1234567890123",
      "name": "Sample Product",
      "name_jpn": "サンプル商品",
      "name_chn": "样品产品",
      "name_kor": "샘플 제품",
      "name_eng": "Sample Product",
      "type": "Electronics",
      "availableForOrder": 1,
      "imagePath": "/550e8400-e29b-41d4-a716-446655440000.jpg",
      "createdAt": "2024-01-22T10:00:00Z",
      "stock": [],
      "tag": [
        {
          "id": "tag_1",
          "tagName": "electronics"
        },
        {
          "id": "sp6",
          "tagName": "special-promotion"
        }
      ]
    },
    "tagNames": ["electronics", "special-promotion"],
    "message": "Item created successfully"
  },
  "success": true,
  "userExists": true
}
```

## Key Features

### Tag Processing
- ✅ **Automatic Tag Association**: Tags provided in the request are automatically associated with the created item
- ✅ **Transaction Safety**: Item creation and tag associations are handled within a database transaction
- ✅ **Duplicate Prevention**: Uses `INSERT IGNORE` to prevent duplicate tag associations
- ✅ **Tag Validation**: Only attempts to associate tags that have valid `id` values

### Image Path Processing
- ✅ **Filename Extraction**: Automatically extracts and formats image filenames from paths
- ✅ **Path Normalization**: Ensures image paths are properly formatted with leading slash

### Response Data
- ✅ **Complete Item Data**: Returns the complete item with all associated tags and stock information
- ✅ **Tag Names Array**: Provides a convenient array of tag names for easy frontend processing
- ✅ **Error Handling**: Graceful fallback if tag or stock data cannot be retrieved

## Error Responses

### Missing Required Fields
```json
{
  "message": "Name is required",
  "payload": null,
  "success": false,
  "userExists": true
}
```

### Code Already Exists
```json
{
  "message": "An item with this code already exists",
  "payload": null,
  "success": false,
  "userExists": true
}
```

### Invalid Image Path
```json
{
  "message": "Invalid image path format",
  "payload": null,
  "success": false,
  "userExists": true
}
```

## Database Operations

When creating an item with tags, the following database operations occur in a single transaction:

1. **Insert Item**: Creates the main item record in the `items` table
2. **Associate Tags**: For each tag provided, creates an association in the `item_tags` table
3. **Commit Transaction**: Ensures all operations succeed or fail together

## Notes

- Items are created with empty stock arrays initially
- Tag associations use the existing tag IDs - tags must already exist in the database
- The function automatically generates item IDs if not provided
- Image paths are processed to ensure proper formatting
- All tag associations include creation timestamps
- The response includes both the complete item object and a simplified tag names array
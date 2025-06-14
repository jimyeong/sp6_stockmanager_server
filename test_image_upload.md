# Image Upload API Test Guide

## Environment Variables Required

Before running the server, make sure to set these environment variables in your `.env.development` or `.env.production` file:

```bash
# R2 Cloudflare Configuration
R2_ACCESS_KEY_ID=your_r2_access_key
R2_SECRET_ACCESS_KEY=your_r2_secret_key
R2_ENDPOINT=https://your_account_id.r2.cloudflarestorage.com
R2_BUCKET_NAME=your_bucket_name
R2_PUBLIC_DOMAIN=your_custom_domain.com  # Optional: if you have a custom domain configured
```

## API Endpoints

**POST** `/api/v1/upload/image` - Upload and process images
**DELETE** `/api/v1/upload/image` - Delete images from storage

## Authentication

Requires Firebase authentication token in the Authorization header:
```
Authorization: Bearer your_firebase_token
```

## Request Format

### Upload (POST)
The request must be sent as `multipart/form-data` with the following field:
- `image`: The image file (JPEG, PNG, or WebP)

### Delete (DELETE)
The request must be sent as JSON with the following field:
- `imagePath`: The full URL or filename of the image to delete

## Image Processing

The API will automatically:
1. ✅ Resize image to maximum 600px width (maintaining aspect ratio)
2. ✅ Convert to JPEG format
3. ✅ Set quality to 70% for optimal compression
4. ✅ Strip EXIF metadata for privacy
5. ✅ Generate UUID for unique filename
6. ✅ Upload to R2 Cloudflare storage

## Example cURL Requests

### Upload Image
```bash
curl -X POST \
  -H "Authorization: Bearer YOUR_FIREBASE_TOKEN" \
  -F "image=@/path/to/your/image.jpg" \
  http://localhost:8080/api/v1/upload/image
```

### Delete Image
```bash
curl -X DELETE \
  -H "Authorization: Bearer YOUR_FIREBASE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"imagePath": "https://your-domain.com/images/550e8400-e29b-41d4-a716-446655440000.jpg"}' \
  http://localhost:8080/api/v1/upload/image
```

Or using just the filename:
```bash
curl -X DELETE \
  -H "Authorization: Bearer YOUR_FIREBASE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"imagePath": "550e8400-e29b-41d4-a716-446655440000.jpg"}' \
  http://localhost:8080/api/v1/upload/image
```

## Example Responses

### Upload Response
```json
{
  "message": "Image uploaded successfully",
  "payload": {
    "imageUrl": "https://your-domain.com/images/550e8400-e29b-41d4-a716-446655440000.jpg",
    "imageId": "550e8400-e29b-41d4-a716-446655440000",
    "fileSize": 145678,
    "message": "Image uploaded successfully",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  },
  "success": true,
  "userExists": true
}
```

### Delete Response
```json
{
  "message": "Image deleted successfully",
  "payload": {
    "message": "Image deleted successfully",
    "success": true,
    "timestamp": "2024-01-15T10:35:00Z",
    "imagePath": "https://your-domain.com/images/550e8400-e29b-41d4-a716-446655440000.jpg"
  },
  "success": true,
  "userExists": true
}
```

## Error Responses

### Authentication Error (401)
```json
{
  "message": "User authentication required",
  "payload": null,
  "success": false,
  "userExists": false
}
```

### Invalid File Type (400)
```json
{
  "message": "Invalid image type. Only JPEG, PNG, and WebP are supported",
  "payload": null,
  "success": false,
  "userExists": true
}
```

### Missing Configuration (500)
```json
{
  "message": "Failed to upload image to storage",
  "payload": null,
  "success": false,
  "userExists": true
}
```

### Missing imagePath (DELETE - 400)
```json
{
  "message": "imagePath is required",
  "payload": null,
  "success": false,
  "userExists": true
}
```

### Invalid Image Path (DELETE - 400)
```json
{
  "message": "Invalid image path format",
  "payload": null,
  "success": false,
  "userExists": true
}
```

## Testing with JavaScript/Fetch

### Upload Image
```javascript
const formData = new FormData();
formData.append('image', fileInput.files[0]);

fetch('/api/v1/upload/image', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${firebaseToken}`
  },
  body: formData
})
.then(response => response.json())
.then(data => {
  console.log('Upload successful:', data.payload.imageUrl);
})
.catch(error => {
  console.error('Upload failed:', error);
});
```

### Delete Image
```javascript
const deleteRequest = {
  imagePath: 'https://your-domain.com/images/550e8400-e29b-41d4-a716-446655440000.jpg'
};

fetch('/api/v1/upload/image', {
  method: 'DELETE',
  headers: {
    'Authorization': `Bearer ${firebaseToken}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(deleteRequest)
})
.then(response => response.json())
.then(data => {
  console.log('Delete successful:', data.payload.message);
})
.catch(error => {
  console.error('Delete failed:', error);
});
```

## Notes

**Upload:**
- Maximum file size is limited by the multipart form parser (default: 32MB)
- All images are converted to JPEG format regardless of input format
- EXIF metadata is automatically stripped for privacy
- Images are resized if width exceeds 600px (height is adjusted proportionally)
- Each image gets a unique UUID filename to prevent conflicts
- The API returns the full public URL for immediate use

**Delete:**
- Accepts either full URLs or just filenames
- Automatically extracts filename from full URLs
- Files are stored in the "images/" folder within the R2 bucket
- Deletion is permanent and cannot be undone
- Returns success even if the file doesn't exist (idempotent operation)
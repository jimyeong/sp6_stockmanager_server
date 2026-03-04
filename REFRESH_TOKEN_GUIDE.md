# Refresh Token Implementation Guide

## Overview

Your Google OAuth2 login now supports **7-day sessions** using a secure refresh token system. This extends the session duration from Firebase's default 1-hour tokens to 7 days, reducing the frequency of re-authentication.

## What Changed

### Backend Changes

1. **Database**: New `refresh_tokens` table to store long-lived refresh tokens
2. **Sign-in Flow**: Now returns a `refresh_token` along with user data
3. **New Endpoints**:
   - `/public/api/v1/auth/refresh` - Refresh expired Firebase tokens
   - `/public/api/v1/auth/revoke` - Revoke a single refresh token (logout)
   - `/api/v1/auth/revoke-all` - Revoke all user tokens (logout all devices)

### Security Features

- **Token Rotation**: Each refresh generates a new token and revokes the old one
- **7-Day Expiration**: Refresh tokens automatically expire after 7 days
- **Revocation**: Tokens can be manually revoked for security
- **Audit Trail**: Tracks IP address, user agent, and last used timestamp

## Database Setup

### 1. Run the Migration

Apply the database migration to create the `refresh_tokens` table:

```bash
mysql -h YOUR_HOST -u YOUR_USER -pYOUR_PASSWORD YOUR_DATABASE < db/migration_refresh_tokens.sql
```

Or manually execute the SQL in `db/schema.sql` (lines 33-49).

### 2. Verify Table Creation

```sql
SHOW TABLES LIKE 'refresh_tokens';
DESC refresh_tokens;
```

## Client-Side Implementation

### 1. Update Sign-In Flow

When users sign in with Google OAuth2, the backend now returns a refresh token:

**Request:**
```javascript
POST /public/api/v1/auth/signin
Content-Type: application/json

{
  "user": {
    "uid": "firebase_user_id",
    "email": "user@example.com",
    "displayName": "John Doe",
    "photoURL": "https://...",
    "emailVerified": true
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Success",
  "data": {
    "user": {
      "uid": "firebase_user_id",
      "email": "user@example.com",
      "displayName": "John Doe",
      "photoURL": "https://...",
      "emailVerified": true,
      "createdAt": "2026-02-12T10:00:00Z",
      "loginAt": "2026-02-12T10:00:00Z"
    },
    "refresh_token": "base64_encoded_token_here",
    "expires_in": 604800
  }
}
```

**Store the refresh token securely:**

```javascript
// Store in secure storage (e.g., encrypted localStorage or secure cookie)
const { user, refresh_token, expires_in } = response.data.data;

// Store refresh token
localStorage.setItem('refresh_token', refresh_token);
localStorage.setItem('refresh_token_expires_at', Date.now() + expires_in * 1000);

// Continue with Firebase authentication as usual
```

### 2. Implement Token Refresh Logic

When Firebase ID tokens expire (after 1 hour), use the refresh token to get a new Firebase custom token:

```javascript
async function refreshFirebaseToken() {
  const refreshToken = localStorage.getItem('refresh_token');

  if (!refreshToken) {
    // No refresh token, user needs to sign in again
    redirectToLogin();
    return null;
  }

  try {
    const response = await fetch('/public/api/v1/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        refresh_token: refreshToken
      })
    });

    if (!response.ok) {
      // Refresh token expired or invalid
      localStorage.removeItem('refresh_token');
      redirectToLogin();
      return null;
    }

    const data = await response.json();
    const { id_token, refresh_token: newRefreshToken } = data.data;

    // Store the new refresh token
    localStorage.setItem('refresh_token', newRefreshToken);
    localStorage.setItem('refresh_token_expires_at', Date.now() + 7 * 24 * 60 * 60 * 1000);

    // Sign in to Firebase with the custom token
    await firebase.auth().signInWithCustomToken(id_token);

    // Get the new Firebase ID token
    const user = firebase.auth().currentUser;
    const firebaseIdToken = await user.getIdToken();

    return firebaseIdToken;
  } catch (error) {
    console.error('Token refresh failed:', error);
    localStorage.removeItem('refresh_token');
    redirectToLogin();
    return null;
  }
}
```

### 3. Implement Automatic Token Refresh

Set up an interceptor to automatically refresh tokens when API calls fail with 401:

```javascript
// Axios interceptor example
axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // If 401 and not already retrying
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // Refresh the token
        const newToken = await refreshFirebaseToken();

        if (newToken) {
          // Retry the original request with new token
          originalRequest.headers['Authorization'] = `Bearer ${newToken}`;
          return axios(originalRequest);
        }
      } catch (refreshError) {
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);
```

### 4. Implement Logout

Revoke the refresh token when users log out:

```javascript
async function logout() {
  const refreshToken = localStorage.getItem('refresh_token');

  if (refreshToken) {
    try {
      await fetch('/public/api/v1/auth/revoke', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          refresh_token: refreshToken
        })
      });
    } catch (error) {
      console.error('Failed to revoke token:', error);
    }
  }

  // Clean up local storage
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('refresh_token_expires_at');

  // Sign out from Firebase
  await firebase.auth().signOut();

  // Redirect to login
  redirectToLogin();
}
```

### 5. Logout from All Devices (Optional)

Revoke all refresh tokens for the current user (requires authentication):

```javascript
async function logoutAllDevices() {
  const firebaseIdToken = await firebase.auth().currentUser.getIdToken();

  try {
    await fetch('/api/v1/auth/revoke-all', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${firebaseIdToken}`,
        'Content-Type': 'application/json',
      }
    });

    // Clean up local storage and sign out
    localStorage.removeItem('refresh_token');
    await firebase.auth().signOut();
    redirectToLogin();
  } catch (error) {
    console.error('Failed to revoke all tokens:', error);
  }
}
```

## API Reference

### POST /public/api/v1/auth/refresh

Refresh an expired Firebase ID token using a refresh token.

**Request:**
```json
{
  "refresh_token": "base64_encoded_token"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "id_token": "firebase_custom_token",
    "refresh_token": "new_base64_encoded_token",
    "expires_in": 3600
  }
}
```

**Errors:**
- `400 Bad Request`: Missing or invalid refresh token
- `401 Unauthorized`: Refresh token expired or revoked

### POST /public/api/v1/auth/revoke

Revoke a single refresh token (logout from current device).

**Request:**
```json
{
  "refresh_token": "base64_encoded_token"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Token revoked successfully",
  "data": null
}
```

### POST /api/v1/auth/revoke-all

Revoke all refresh tokens for the authenticated user (logout from all devices).

**Headers:**
```
Authorization: Bearer <firebase_id_token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "All tokens revoked successfully",
  "data": null
}
```

**Errors:**
- `401 Unauthorized`: Invalid or missing Firebase ID token

## Complete React/Next.js Example

```javascript
// hooks/useAuth.js
import { useState, useEffect } from 'react';
import { auth } from '../firebase';
import axios from 'axios';

export function useAuth() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Listen to Firebase auth state
    const unsubscribe = auth.onAuthStateChanged(async (firebaseUser) => {
      if (firebaseUser) {
        setUser(firebaseUser);
      } else {
        // Try to refresh token if we have a refresh token
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          await refreshFirebaseToken();
        }
      }
      setLoading(false);
    });

    return unsubscribe;
  }, []);

  const signInWithGoogle = async () => {
    const provider = new firebase.auth.GoogleAuthProvider();
    const result = await auth.signInWithPopup(provider);
    const user = result.user;

    // Send user data to backend to get refresh token
    const response = await axios.post('/public/api/v1/auth/signin', {
      user: {
        uid: user.uid,
        email: user.email,
        displayName: user.displayName,
        photoURL: user.photoURL,
        emailVerified: user.emailVerified,
      }
    });

    const { refresh_token, expires_in } = response.data.data;
    localStorage.setItem('refresh_token', refresh_token);
    localStorage.setItem('refresh_token_expires_at', Date.now() + expires_in * 1000);

    setUser(user);
  };

  const refreshFirebaseToken = async () => {
    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) return null;

    try {
      const response = await axios.post('/public/api/v1/auth/refresh', {
        refresh_token: refreshToken
      });

      const { id_token, refresh_token: newRefreshToken } = response.data.data;

      localStorage.setItem('refresh_token', newRefreshToken);
      await auth.signInWithCustomToken(id_token);

      const firebaseUser = auth.currentUser;
      setUser(firebaseUser);

      return await firebaseUser.getIdToken();
    } catch (error) {
      console.error('Token refresh failed:', error);
      localStorage.removeItem('refresh_token');
      return null;
    }
  };

  const logout = async () => {
    const refreshToken = localStorage.getItem('refresh_token');

    if (refreshToken) {
      try {
        await axios.post('/public/api/v1/auth/revoke', {
          refresh_token: refreshToken
        });
      } catch (error) {
        console.error('Failed to revoke token:', error);
      }
    }

    localStorage.removeItem('refresh_token');
    await auth.signOut();
    setUser(null);
  };

  return {
    user,
    loading,
    signInWithGoogle,
    logout,
    refreshFirebaseToken
  };
}
```

## Testing

### Test the Flow

1. **Sign in** and verify you receive a refresh token
2. **Wait 1 hour** for Firebase token to expire
3. **Make an API call** and verify it automatically refreshes
4. **Check refresh_tokens table** to see the token rotation
5. **Logout** and verify the token is revoked

### Manual Testing with curl

```bash
# 1. Sign in (you'll need to get a Firebase user first)
curl -X POST http://localhost:8080/public/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{"user": {"uid": "test_uid", "email": "test@example.com", "displayName": "Test User", "emailVerified": true, "photoURL": "", "isAnonymous": false, "phoneNumber": "", "providerId": "google.com"}}'

# 2. Refresh token (use the refresh_token from step 1)
curl -X POST http://localhost:8080/public/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN_HERE"}'

# 3. Revoke token
curl -X POST http://localhost:8080/public/api/v1/auth/revoke \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN_HERE"}'
```

## Security Best Practices

1. **Store refresh tokens securely**: Use httpOnly cookies or encrypted storage
2. **Use HTTPS**: Always transmit tokens over secure connections
3. **Implement CSRF protection**: If using cookies for storage
4. **Monitor suspicious activity**: Track IP addresses and user agents
5. **Periodic cleanup**: Run `models.CleanupExpiredTokens()` regularly to remove old tokens
6. **Rate limiting**: Limit refresh token endpoint to prevent abuse

## Maintenance

### Cleanup Expired Tokens

Add a cron job or periodic task to clean up expired tokens:

```go
// In your main.go or a separate cleanup service
func startTokenCleanup() {
    ticker := time.NewTicker(24 * time.Hour) // Run daily
    go func() {
        for range ticker.C {
            err := models.CleanupExpiredTokens()
            if err != nil {
                log.Printf("Failed to cleanup expired tokens: %v", err)
            } else {
                log.Println("Successfully cleaned up expired tokens")
            }
        }
    }()
}
```

## Troubleshooting

### "Refresh token expired" errors

- Refresh tokens expire after 7 days
- Users need to sign in again after 7 days of inactivity
- Check `expires_at` column in `refresh_tokens` table

### "Invalid refresh token" errors

- Token may have been revoked manually
- Token may not exist in database
- Check `is_revoked` column in database

### Database connection errors

- Verify database credentials in `.env` files
- Check if `refresh_tokens` table exists
- Ensure foreign key constraint with `users` table is satisfied

## Summary

Your application now supports **7-day sessions** with the following benefits:

- ✅ Users stay logged in for 7 days instead of 1 hour
- ✅ Secure token rotation on each refresh
- ✅ Ability to revoke tokens (logout) and invalidate all devices
- ✅ Audit trail with IP addresses and timestamps
- ✅ Automatic cleanup of expired tokens

The Firebase ID tokens still expire after 1 hour for security, but they're automatically refreshed using the long-lived refresh token in the background.

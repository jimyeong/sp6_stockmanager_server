# Quick Start: 7-Day Sessions with Refresh Tokens

## What's New?

Your Google OAuth2 login now supports **7-day sessions** instead of the default 1-hour Firebase token expiration. Users stay logged in for 7 days without needing to re-authenticate.

## Setup (3 Steps)

### 1. Run Database Migration

```bash
# Connect to your database and run:
mysql -h YOUR_HOST -u YOUR_USER -pYOUR_PASSWORD YOUR_DATABASE < db/migration_refresh_tokens.sql

# Or manually execute the SQL from db/schema.sql (refresh_tokens table)
```

### 2. Update Client Sign-In Code

After Google OAuth2 sign-in, the backend now returns a `refresh_token`:

```javascript
// Sign in with Google (existing code)
const result = await signInWithGoogle();

// NEW: Send user data to backend
const response = await fetch('/public/api/v1/auth/signin', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    user: {
      uid: result.user.uid,
      email: result.user.email,
      displayName: result.user.displayName,
      photoURL: result.user.photoURL,
      emailVerified: result.user.emailVerified,
    }
  })
});

const data = await response.json();

// NEW: Store the refresh token
localStorage.setItem('refresh_token', data.data.refresh_token);
```

### 3. Add Token Refresh Logic

When Firebase tokens expire (after 1 hour), automatically refresh them:

```javascript
// Add this function to refresh expired tokens
async function refreshFirebaseToken() {
  const refreshToken = localStorage.getItem('refresh_token');

  const response = await fetch('/public/api/v1/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken })
  });

  const data = await response.json();

  // Store new refresh token
  localStorage.setItem('refresh_token', data.data.refresh_token);

  // Sign in with new Firebase custom token
  await firebase.auth().signInWithCustomToken(data.data.id_token);

  return await firebase.auth().currentUser.getIdToken();
}

// Add axios interceptor to auto-refresh on 401 errors
axios.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 401 && !error.config._retry) {
      error.config._retry = true;
      const newToken = await refreshFirebaseToken();
      error.config.headers['Authorization'] = `Bearer ${newToken}`;
      return axios(error.config);
    }
    return Promise.reject(error);
  }
);
```

## New API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/public/api/v1/auth/signin` | POST | Sign in (now returns refresh_token) |
| `/public/api/v1/auth/refresh` | POST | Refresh expired Firebase token |
| `/public/api/v1/auth/revoke` | POST | Logout (revoke refresh token) |
| `/api/v1/auth/revoke-all` | POST | Logout all devices (requires auth) |

## Test It

```bash
# Run the test script
node examples/test_refresh_token.js

# You should see:
# ✅ Sign in successful!
# ✅ Token refresh successful!
# ✅ All tests completed successfully!
```

## How It Works

```
User Signs In
    ↓
Backend creates refresh token (valid 7 days)
    ↓
Client stores refresh token locally
    ↓
Firebase token expires (1 hour later)
    ↓
Client sends refresh token to /auth/refresh
    ↓
Backend creates new Firebase custom token
Backend creates new refresh token (rotation for security)
    ↓
Client signs in with custom token
Client stores new refresh token
    ↓
Process repeats for 7 days
```

## Benefits

- ✅ **7-day sessions** - Users stay logged in for a week
- ✅ **Secure** - Token rotation on each refresh
- ✅ **Flexible** - Can revoke tokens anytime
- ✅ **Audit trail** - Tracks IP, user agent, timestamps
- ✅ **Automatic cleanup** - Expired tokens removed automatically

## Need More Help?

See the full guide: `REFRESH_TOKEN_GUIDE.md`

## Summary

Your users now enjoy 7-day sessions while maintaining security through token rotation. The Firebase ID tokens still expire after 1 hour (for security), but they're automatically refreshed in the background using the long-lived refresh token.

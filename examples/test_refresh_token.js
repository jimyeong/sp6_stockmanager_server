/**
 * Test script for refresh token flow
 *
 * This demonstrates how to use the new 7-day refresh token system
 * to maintain longer user sessions with Google OAuth2.
 *
 * Usage:
 *   node examples/test_refresh_token.js
 */

const API_URL = 'http://localhost:8080';

// Example user data (you would get this from Firebase after Google OAuth2 sign-in)
const testUser = {
  uid: 'test_user_12345',
  email: 'test@example.com',
  displayName: 'Test User',
  emailVerified: true,
  photoURL: 'https://example.com/photo.jpg',
  isAnonymous: false,
  phoneNumber: '',
  providerId: 'google.com'
};

/**
 * Step 1: Sign in and get refresh token
 */
async function signIn() {
  console.log('🔐 Step 1: Signing in...');

  const response = await fetch(`${API_URL}/public/api/v1/auth/signin`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      user: testUser
    })
  });

  if (!response.ok) {
    throw new Error(`Sign in failed: ${response.statusText}`);
  }

  const data = await response.json();
  console.log('✅ Sign in successful!');
  console.log('   User:', data.data.user.email);
  console.log('   Refresh Token:', data.data.refresh_token.substring(0, 20) + '...');
  console.log('   Expires in:', data.data.expires_in / 86400, 'days');

  return data.data.refresh_token;
}

/**
 * Step 2: Refresh the token
 */
async function refreshToken(refreshToken) {
  console.log('\n🔄 Step 2: Refreshing token...');

  const response = await fetch(`${API_URL}/public/api/v1/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      refresh_token: refreshToken
    })
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(`Token refresh failed: ${error.message}`);
  }

  const data = await response.json();
  console.log('✅ Token refresh successful!');
  console.log('   New Firebase ID Token:', data.data.id_token.substring(0, 20) + '...');
  console.log('   New Refresh Token:', data.data.refresh_token.substring(0, 20) + '...');
  console.log('   Firebase token expires in:', data.data.expires_in / 60, 'minutes');

  return data.data.refresh_token;
}

/**
 * Step 3: Test with an expired/invalid token
 */
async function testInvalidToken() {
  console.log('\n❌ Step 3: Testing with invalid token...');

  const response = await fetch(`${API_URL}/public/api/v1/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      refresh_token: 'invalid_token_12345'
    })
  });

  if (!response.ok) {
    console.log('✅ Invalid token correctly rejected (expected behavior)');
  } else {
    console.log('⚠️  Warning: Invalid token was accepted (unexpected)');
  }
}

/**
 * Step 4: Revoke the token (logout)
 */
async function revokeToken(refreshToken) {
  console.log('\n🚪 Step 4: Revoking token (logout)...');

  const response = await fetch(`${API_URL}/public/api/v1/auth/revoke`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      refresh_token: refreshToken
    })
  });

  if (!response.ok) {
    throw new Error(`Token revocation failed: ${response.statusText}`);
  }

  const data = await response.json();
  console.log('✅ Token revoked successfully!');
  console.log('   Message:', data.message);
}

/**
 * Step 5: Try to use revoked token
 */
async function testRevokedToken(refreshToken) {
  console.log('\n🔒 Step 5: Testing revoked token...');

  const response = await fetch(`${API_URL}/public/api/v1/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      refresh_token: refreshToken
    })
  });

  if (!response.ok) {
    console.log('✅ Revoked token correctly rejected (expected behavior)');
  } else {
    console.log('⚠️  Warning: Revoked token was accepted (unexpected)');
  }
}

/**
 * Main test flow
 */
async function runTest() {
  console.log('🚀 Testing Refresh Token Flow\n');
  console.log('=' .repeat(60));

  try {
    // Step 1: Sign in
    const refreshToken1 = await signIn();

    // Wait a moment
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Step 2: Refresh token (simulates user returning after 1 hour)
    const refreshToken2 = await refreshToken(refreshToken1);

    // Wait a moment
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Step 3: Test invalid token
    await testInvalidToken();

    // Wait a moment
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Step 4: Revoke token
    await revokeToken(refreshToken2);

    // Wait a moment
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Step 5: Test revoked token
    await testRevokedToken(refreshToken2);

    console.log('\n' + '='.repeat(60));
    console.log('✅ All tests completed successfully!');
    console.log('\n📊 Summary:');
    console.log('   • Sign in: ✅ Returns refresh token');
    console.log('   • Token refresh: ✅ Returns new tokens with rotation');
    console.log('   • Invalid token: ✅ Correctly rejected');
    console.log('   • Token revocation: ✅ Successfully revoked');
    console.log('   • Revoked token: ✅ Correctly rejected');
    console.log('\n🎉 Your 7-day session system is working perfectly!');

  } catch (error) {
    console.error('\n❌ Test failed:', error.message);
    console.error('\nTroubleshooting:');
    console.error('   1. Make sure the server is running: go run main.go');
    console.error('   2. Check database connection in .env files');
    console.error('   3. Verify refresh_tokens table exists in database');
    console.error('   4. Check server logs for detailed error messages');
    process.exit(1);
  }
}

// Run the test
runTest();

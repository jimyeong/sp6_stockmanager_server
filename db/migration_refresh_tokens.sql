-- Migration: Add refresh tokens table for extended session management
-- Created: 2026-02-12
-- Purpose: Enable 7-day session duration with secure refresh token rotation

-- Create refresh_tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(128) PRIMARY KEY,
    user_id VARCHAR(128) NOT NULL,
    token VARCHAR(512) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NULL,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    user_agent VARCHAR(512),
    ip_address VARCHAR(45),
    FOREIGN KEY (user_id) REFERENCES users(firebase_uid) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Verify table creation
SELECT 'refresh_tokens table created successfully!' AS Status;

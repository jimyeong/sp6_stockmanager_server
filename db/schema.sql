-- Items table
CREATE TABLE IF NOT EXISTS items (
    id VARCHAR(128) PRIMARY KEY,
    barcode VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    quantity_in_stock INT NOT NULL DEFAULT 0,
    unit_price DECIMAL(10, 2) NOT NULL,
    last_updated TIMESTAMP NOT NULL,
    creator_id VARCHAR(128),
    created_at TIMESTAMP NOT NULL
);

-- Stock Transactions table
CREATE TABLE IF NOT EXISTS stock_transactions (
    id VARCHAR(128) PRIMARY KEY,
    item_id VARCHAR(128) NOT NULL,
    quantity INT NOT NULL,
    type ENUM('in', 'out') NOT NULL,
    user_id VARCHAR(128) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);

-- Index for faster queries
CREATE INDEX idx_stock_transactions_item_id ON stock_transactions(item_id);
CREATE INDEX idx_stock_transactions_user_id ON stock_transactions(user_id);
CREATE INDEX idx_items_barcode ON items(barcode);

-- Refresh Tokens table
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

-- Index for faster queries on refresh tokens
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
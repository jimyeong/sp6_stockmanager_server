

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

-- v1
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

CREATE TABLE IF NOT EXISTS inventory_logs (
    inventory_log_id int AUTO_INCREMENT PRIMARY KEY,
    event_type enum('stock_in', 'stock_out', 'expired', 'damaged', 'sold', 'discounted') NOT NULL,
    product_id INT NOT NULL,
    product_code VARCHAR(50) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    product_image_path VARCHAR(255) NULL,
    stock_id INT NOT NULL,
    stock_type ENUM('BOX', 'PCS') NOT NULL,
    stock_quantity INT NOT NULL,
    expiry_date DATE,
    original_price DECIMAL(10, 2),
    discounted_price DECIMAL(10, 2),
    discount_rate DECIMAL(5, 2),
    performer_id INT NOT NULL,
    performer_name VARCHAR(50) NOT NULL,
    performer_email VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)

CREATE INDEX idx_inventory_logs_product_id ON inventory_logs(product_id);
CREATE INDEX idx_inventory_logs_product_event_date 
ON inventory_logs(product_id, event_type, created_at)
CREATE INDEX idx_inventory_logs_event_date ON inventory_logs(event_type, created_at)
CREATE INDEX idx_inventory_logs_created_at ON inventory_logs(created_at)

-- CREATE INDEX idx_inventory_logs_product_event_date 
-- ON inventory_logs(product_id, event_type, created_at);

-- CREATE INDEX idx_inventory_logs_event_date 
-- ON inventory_logs(event_type, created_at);

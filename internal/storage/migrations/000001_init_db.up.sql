BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS users (
    login VARCHAR(500) PRIMARY KEY,
    pass_hash VARCHAR(500) NOT NULL,
    balance DECIMAL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS orders (
    number VARCHAR(500) PRIMARY KEY,
    user_login VARCHAR(500) REFERENCES users (login),
    status VARCHAR(20),
    accrual DECIMAL DEFAULT 0,
    uploaded_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS withdrawals (
    user_login VARCHAR(500),
    order_id VARCHAR(500),
    withdrawal_sum DECIMAL,
    processed_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_user_login ON orders(user_login);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_login ON withdrawals(user_login);
COMMIT TRANSACTION;
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS users (
    login TEXT PRIMARY KEY,
    pass_hash TEXT NOT NULL,
    balance DECIMAL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS orders (
    number TEXT PRIMARY KEY,
    user_login TEXT REFERENCES users (login),
    status VARCHAR(20),
    accrual DECIMAL DEFAULT 0,
    uploaded_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS withdrawals (
    user_login TEXT,
    order_id TEXT,
    withdrawal_sum DECIMAL,
    processed_at TIMESTAMP DEFAULT NOW()
);
COMMIT TRANSACTION;
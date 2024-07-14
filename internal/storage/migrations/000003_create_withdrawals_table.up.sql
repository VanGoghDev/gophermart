CREATE TABLE IF NOT EXISTS withdrawals (
    user_login TEXT,
    order_id TEXT,
    withdrawal_sum DECIMAL,
    processed_at TIMESTAMP
)
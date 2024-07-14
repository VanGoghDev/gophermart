CREATE TABLE IF NOT EXISTS withdrawals (
    user_login TEXT,
    order_id TEXT,
    withdrawal_sum INTEGER,
    processed_at TIMESTAMP
)
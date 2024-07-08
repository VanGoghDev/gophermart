CREATE TABLE IF NOT EXISTS withdrawals (
    user_login TEXT,
    order_id INTEGER,
    withdrawal_sum INTEGER,
    processed_at TIMESTAMP
)
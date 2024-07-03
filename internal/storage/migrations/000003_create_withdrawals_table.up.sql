CREATE TABLE IF NOT EXISTS withdrawals (
    user_login   TEXT,
    order_id     INTEGER,
    sum          INTEGER,
    processed_at TIME
)
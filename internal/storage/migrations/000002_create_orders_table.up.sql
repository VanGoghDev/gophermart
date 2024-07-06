CREATE TABLE IF NOT EXISTS orders (
    number         INTEGER PRIMARY KEY,
    user_login     TEXT REFERENCES users (login),
    status         VARCHAR(20),
    accrual        INTEGER DEFAULT 0,
    uploaded_at    TIMESTAMP DEFAULT NOW()
)
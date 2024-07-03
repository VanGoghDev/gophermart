CREATE TABLE IF NOT EXISTS orders (
    id         INTEGER PRIMARY KEY,
    user_login TEXT REFERENCES users (login),
    status     VARCHAR(20),
    accural    INTEGER
)
BEGIN;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS withdrawals CASCADE;
DROP INDEX IF EXISTS idx_orders_user_login;
DROP INDEX IF EXISTS idx_withdrawals_user_login;
DROP INDEX IF EXISTS idx_withdrawals_order_id;
DROP INDEX IF EXISTS idx_withdrawals;
COMMIT;
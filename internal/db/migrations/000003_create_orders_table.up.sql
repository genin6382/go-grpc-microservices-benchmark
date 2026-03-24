-- 000003_create_orders_table.sql
CREATE TABLE IF NOT EXISTS orders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    product_id  UUID NOT NULL,
    quantity    INT NOT NULL,
    total_cost  NUMERIC(10, 2) NOT NULL,
    status      VARCHAR(50) DEFAULT 'pending',  -- pending, confirmed, failed
    created_at  TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_orders_user_id ON orders (user_id);
-- 000001_create_users_table.sql
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    password    VARCHAR(255) NOT NULL,        
    created_at  TIMESTAMP DEFAULT NOW()
);


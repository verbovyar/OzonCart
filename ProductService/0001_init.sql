-- +goose Up
CREATE TABLE IF NOT EXISTS products (
    sku   BIGINT PRIMARY KEY,
    name  TEXT    NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0)
);

-- +goose Down
DROP TABLE IF EXISTS products;

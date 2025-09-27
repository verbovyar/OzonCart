-- +goose Up
CREATE TABLE IF NOT EXISTS cart (
    user_id BIGINT NOT NULL,
    sku_id  BIGINT NOT NULL,
    count   BIGINT NOT NULL CHECK (count > 0),
    PRIMARY KEY (user_id, sku_id)
);

-- +goose Down
DROP TABLE IF EXISTS cart;

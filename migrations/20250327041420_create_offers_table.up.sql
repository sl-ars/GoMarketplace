CREATE TABLE offers (
                        id BIGSERIAL PRIMARY KEY,
                        product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                        seller_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        price DECIMAL(10,2) NOT NULL,
                        stock INT NOT NULL DEFAULT 0,
                        is_available BOOLEAN NOT NULL DEFAULT TRUE,
                        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                        UNIQUE (product_id, seller_id)
);
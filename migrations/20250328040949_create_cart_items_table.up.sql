CREATE TABLE cart_items (
                            id BIGSERIAL PRIMARY KEY,
                            user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                            offer_id BIGINT NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
                            quantity INT NOT NULL CHECK (quantity > 0),
                            created_at TIMESTAMP DEFAULT now(),

                            UNIQUE (user_id, offer_id)
);
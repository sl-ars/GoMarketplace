
CREATE TABLE refunds (
                         id              BIGSERIAL PRIMARY KEY,
                         order_item_id   BIGINT      NOT NULL REFERENCES order_items(id)        ON DELETE CASCADE,
                         requester_id    BIGINT      NOT NULL REFERENCES users(id)              ON DELETE CASCADE,
                         seller_id       BIGINT      NOT NULL REFERENCES users(id)              ON DELETE CASCADE,
                         amount          NUMERIC(10,2) NOT NULL,
                         reason          TEXT,
                         status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
                         created_at      TIMESTAMP    NOT NULL DEFAULT now(),
                         updated_at      TIMESTAMP    NOT NULL DEFAULT now(),
                         CONSTRAINT refund_one_per_item UNIQUE(order_item_id)
);
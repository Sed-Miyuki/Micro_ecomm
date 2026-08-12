ALTER TABLE orders 
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending',
    ADD CONSTRAINT chk_orders_status CHECK (status IN ('pending', 'shipped', 'delivered'));
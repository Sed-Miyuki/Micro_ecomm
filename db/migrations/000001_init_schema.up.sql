CREATE TABLE products (
  id BIGSERIAL PRIMARY KEY,
  name varchar(255) NOT NULL,
  image varchar(255) NOT NULL,
  category varchar(255) NOT NULL,
  description text,
  rating int NOT NULL,
  num_reviews int NOT NULL DEFAULT 0,
  price decimal(10,2) NOT NULL,
  count_in_stock int NOT NULL,
  created_at timestamp DEFAULT now(),
  updated_at timestamp
);

CREATE TABLE orders (
  id BIGSERIAL PRIMARY KEY,
  payment_method varchar(255) NOT NULL,
  tax_price decimal(10,2) NOT NULL,
  shipping_price decimal(10,2) NOT NULL,
  total_price decimal(10,2) NOT NULL,
  created_at timestamp DEFAULT now(),
  updated_at timestamp
);

CREATE TABLE order_items (
  id BIGSERIAL PRIMARY KEY,
  order_id bigint NOT NULL,
  product_id bigint NOT NULL,
  name varchar(255) NOT NULL,
  quantity int NOT NULL,
  image varchar(255) NOT NULL,
  price int NOT NULL
);

ALTER TABLE order_items ADD FOREIGN KEY (order_id) REFERENCES orders (id);

ALTER TABLE order_items ADD FOREIGN KEY (product_id) REFERENCES products (id);

CREATE TABLE orders (
    id int NOT NULL,
    item_id int NOT NULL,
    quantity int NOT NULL,
    order_date TIMESTAMP WITH TIME ZONE,
    charged boolean NOT NULL DEFAULT false,
    fulfilled boolean NOT NULL DEFAULT false
);

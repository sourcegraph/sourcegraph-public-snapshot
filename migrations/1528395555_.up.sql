ALTER TABLE users ADD COLUMN billing_customer_id text;
CREATE UNIQUE INDEX users_billing_customer_id ON users(billing_customer_id) WHERE deleted_at IS NULL;

CREATE TABLE product_subscriptions (
  id uuid NOT NULL PRIMARY KEY,
  user_id integer NOT NULL REFERENCES users(id),
  billing_subscription_id text,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  archived_at timestamp with time zone
);

CREATE TABLE product_licenses (
  id uuid NOT NULL PRIMARY KEY,
  product_subscription_id uuid NOT NULL REFERENCES product_subscriptions(id),
  license_key text NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now()
);

DROP TABLE product_licenses;
DROP TABLE product_subscriptions;

DROP INDEX users_billing_customer_id;
ALTER TABLE users DROP COLUMN billing_customer_id;

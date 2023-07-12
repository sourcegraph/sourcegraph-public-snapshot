ALTER TABLE IF EXISTS event_logs
    DROP COLUMN IF EXISTS billing_product_category,
    DROP COLUMN IF EXISTS billing_event_id,
    DROP COLUMN IF EXISTS client;

ALTER TABLE IF EXISTS event_logs
    ADD COLUMN IF NOT EXISTS billing_product_category text,
    ADD COLUMN IF NOT EXISTS billing_event_id text,
    ADD COLUMN IF NOT EXISTS client text;

ALTER TABLE order_items
ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

UPDATE order_items
SET updated_at = created_at
WHERE updated_at IS NULL;
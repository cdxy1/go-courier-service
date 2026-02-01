-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'delivery_courier_fk'
    ) THEN
        ALTER TABLE delivery
            ADD CONSTRAINT delivery_courier_fk
            FOREIGN KEY (courier_id)
            REFERENCES couriers(id);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE delivery
    DROP CONSTRAINT IF EXISTS delivery_courier_fk;
-- +goose StatementEnd

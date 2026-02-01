-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_couriers_status_assignments_id
    ON couriers (status, assignments_count, id);

CREATE INDEX IF NOT EXISTS idx_delivery_order_id
    ON delivery (order_id);

CREATE INDEX IF NOT EXISTS idx_delivery_courier_deadline
    ON delivery (courier_id, deadline);

CREATE INDEX IF NOT EXISTS idx_delivery_deadline
    ON delivery (deadline);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_delivery_deadline;
DROP INDEX IF EXISTS idx_delivery_courier_deadline;
DROP INDEX IF EXISTS idx_delivery_order_id;
DROP INDEX IF EXISTS idx_couriers_status_assignments_id;
-- +goose StatementEnd

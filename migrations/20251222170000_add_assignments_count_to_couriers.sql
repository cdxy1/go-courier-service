-- +goose Up
-- +goose StatementBegin
ALTER TABLE couriers
    ADD COLUMN IF NOT EXISTS assignments_count BIGINT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE couriers
    DROP COLUMN IF EXISTS assignments_count;
-- +goose StatementEnd

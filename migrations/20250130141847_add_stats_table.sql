-- +goose Up
-- +goose StatementBegin
CREATE TABLE analytics (
    id SERIAL PRIMARY KEY,
    total_url_count INT,
    day_peak INT,
    leaders JSONB,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS analytics;
-- +goose StatementEnd

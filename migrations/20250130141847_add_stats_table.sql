-- +goose Up
-- +goose StatementBegin
CREATE TABLE analytics (
    id SERIAL PRIMARY KEY,
    total_url_count INT,
    url_per_min INT,
    day_peak INT,
    timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS url;
-- +goose StatementEnd

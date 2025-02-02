-- +goose Up
-- +goose StatementBegin
WITH cte AS (
    SELECT MIN(id) AS id, url
    FROM url
    GROUP BY url
)
DELETE FROM url
WHERE id NOT IN (SELECT id FROM cte);

ALTER TABLE url
    ADD CONSTRAINT url_unique UNIQUE (url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE url
    DROP CONSTRAINT IF EXISTS alias_unique;
-- +goose StatementEnd

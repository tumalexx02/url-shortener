-- +goose Up
-- +goose StatementBegin
WITH cte AS (
    SELECT MIN(id) AS id, alias
    FROM url
    GROUP BY alias
)
DELETE FROM url
WHERE id NOT IN (SELECT id FROM cte);

ALTER TABLE url
    ADD CONSTRAINT alias_unique UNIQUE (alias);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE url
    DROP CONSTRAINT IF EXISTS alias_unique;
-- +goose StatementEnd

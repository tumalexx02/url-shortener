-- +goose Up
CREATE INDEX idx_alias ON url (alias);

-- +goose Down
DROP INDEX IF EXISTS idx_alias;
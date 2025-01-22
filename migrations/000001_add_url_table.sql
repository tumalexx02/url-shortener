-- +goose Up
CREATE TABLE url (
   id SERIAL PRIMARY KEY,
   alias TEXT NOT NULL UNIQUE,
   url TEXT NOT NULL);

-- +goose Down
DROP TABLE IF EXISTS url;
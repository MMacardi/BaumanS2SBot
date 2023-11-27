-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
     id SERIAL PRIMARY KEY,
     user_id INTEGER UNIQUE NOT NULL,
     username VARCHAR(255) NOT NULL,
     ranking INT DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd

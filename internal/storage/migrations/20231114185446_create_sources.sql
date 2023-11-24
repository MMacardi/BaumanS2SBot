-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
     Id SERIAL PRIMARY KEY,
     UserID INTEGER NOT NULL,
     Name VARCHAR(255) NOT NULL,
     Category VARCHAR(255) NOT NULL,
     Ranking INT DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_categories (
     user_id INT REFERENCES users(user_id),
     category_id INT REFERENCES categories(id),
     PRIMARY KEY (user_id, category_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_categories;
-- +goose StatementEnd

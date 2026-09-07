-- +goose Up
CREATE TABLE users (
    user_id UUID,
    username TEXT NOT NULL UNIQUE ,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY(user_id)
);

-- +goose Down
DROP TABLE users;

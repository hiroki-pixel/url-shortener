-- +goose Up
CREATE TABLE urls (
    url_id UUID,
    user_id UUID NOT NULL,
    original_url TEXT NOT NULL CHECK (char_length(original_url) <= 2048),
    memo TEXT CHECK (char_length(memo) <= 256),
    short_code TEXT NOT NULL UNIQUE  CHECK (char_length(short_code) = 6),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY(url_id),
    FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    UNIQUE(user_id,original_url)
);

-- +goose Down
DROP TABLE urls;

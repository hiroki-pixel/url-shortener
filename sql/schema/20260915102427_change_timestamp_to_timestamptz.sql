-- +goose Up

ALTER TABLE users
ALTER COLUMN created_at TYPE TIMESTAMPTZ
USING created_at AT TIME ZONE 'Asia/Tokyo';

ALTER TABLE users
ALTER COLUMN updated_at TYPE TIMESTAMPTZ
USING updated_at AT TIME ZONE 'Asia/Tokyo';


ALTER TABLE sessions
ALTER COLUMN created_at TYPE TIMESTAMPTZ
USING created_at AT TIME ZONE 'Asia/Tokyo';

ALTER TABLE sessions
ALTER COLUMN expires_at TYPE TIMESTAMPTZ
USING expires_at AT TIME ZONE 'Asia/Tokyo';


ALTER TABLE urls
ALTER COLUMN created_at TYPE TIMESTAMPTZ
USING created_at AT TIME ZONE 'Asia/Tokyo';

ALTER TABLE urls
ALTER COLUMN updated_at TYPE TIMESTAMPTZ
USING updated_at AT TIME ZONE 'Asia/Tokyo';

-- +goose Down

ALTER TABLE users
ALTER COLUMN created_at TYPE TIMESTAMP
USING created_at AT TIME ZONE 'Asia/Tokyo';

ALTER TABLE users
ALTER COLUMN updated_at TYPE TIMESTAMP
USING updated_at AT TIME ZONE 'Asia/Tokyo';


ALTER TABLE sessions
ALTER COLUMN created_at TYPE TIMESTAMP
USING created_at AT TIME ZONE 'Asia/Tokyo';

ALTER TABLE sessions
ALTER COLUMN expires_at TYPE TIMESTAMP
USING expires_at AT TIME ZONE 'Asia/Tokyo';


ALTER TABLE urls
ALTER COLUMN created_at TYPE TIMESTAMP
USING created_at AT TIME ZONE 'Asia/Tokyo';

ALTER TABLE urls
ALTER COLUMN updated_at TYPE TIMESTAMP
USING updated_at AT TIME ZONE 'Asia/Tokyo';

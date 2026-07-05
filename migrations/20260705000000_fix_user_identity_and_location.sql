-- +goose Up
ALTER TABLE users
    MODIFY login VARCHAR(50) NULL,
    MODIFY chatID BIGINT NOT NULL;

CREATE UNIQUE INDEX ux_users_chat_id ON users (chatID);

DELETE older
FROM user_location AS older
INNER JOIN user_location AS newer
    ON older.user_id = newer.user_id
    AND (
        older.created_at < newer.created_at
        OR (older.created_at = newer.created_at AND older.id < newer.id)
    );

CREATE UNIQUE INDEX ux_user_location_user_id ON user_location (user_id);

-- +goose Down
DROP INDEX ux_user_location_user_id ON user_location;
DROP INDEX ux_users_chat_id ON users;

UPDATE users SET login = CONCAT('user_', id) WHERE login IS NULL;

ALTER TABLE users
    MODIFY login VARCHAR(50) NOT NULL,
    MODIFY chatID INT;

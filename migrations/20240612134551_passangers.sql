-- +goose Up
CREATE TABLE IF NOT EXISTS passengers(
                                         name VARCHAR(50) not null,
                                         event_id bigint unsigned NOT NULL,
                                         user_chatId BIGINT NOT NULL,
                                         FOREIGN KEY (event_id) REFERENCES events (id_events)
);

-- +goose Down
DROP TABLE IF EXISTS passengers;

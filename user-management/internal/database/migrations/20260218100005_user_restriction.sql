-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_restrictions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_uuid VARCHAR(36) NOT NULL DEFAULT '',
    restriction_code VARCHAR(255) NOT NULL,
    status VARCHAR(36) NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_user_uuid (user_uuid),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_restrictions;
-- +goose StatementEnd

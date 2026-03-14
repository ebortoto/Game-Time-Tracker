CREATE TABLE IF NOT EXISTS daily_history (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    game_name VARCHAR(255) NOT NULL,
    play_date DATE NOT NULL,
    total_time_secs BIGINT NOT NULL DEFAULT 0,
    last_played_date DATETIME NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_game_date (game_name, play_date),
    KEY idx_last_played_date (last_played_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

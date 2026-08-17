CREATE TABLE tasks (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    team_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    status VARCHAR(50) NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    assignee_id BIGINT UNSIGNED NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP NULL,
    version BIGINT UNSIGNED NOT NULL DEFAULT 1,

    PRIMARY KEY (id),

    CONSTRAINT fk_tasks_team
        FOREIGN KEY (team_id)
        REFERENCES teams(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_tasks_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id),

    CONSTRAINT fk_tasks_assignee
        FOREIGN KEY (assignee_id)
        REFERENCES users(id)
        ON DELETE SET NULL,

    INDEX idx_tasks_team_status (team_id, status),
    INDEX idx_tasks_team_assignee (team_id, assignee_id),
    INDEX idx_tasks_team_created_at (team_id, created_at)
) ENGINE=InnoDB;
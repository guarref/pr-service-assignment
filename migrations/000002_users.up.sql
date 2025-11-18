CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(120) PRIMARY KEY,
    username VARCHAR(120) NOT NULL,
    team_name VARCHAR(120) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Foreign Key на команду
    CONSTRAINT fk_users_team FOREIGN KEY (team_name) 
        REFERENCES teams(team_name) 
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

-- Индексы для быстрого поиска
CREATE INDEX idx_users_team_name ON users(team_name);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_team_active ON users(team_name, is_active); -- составной индекс

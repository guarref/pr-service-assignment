CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(120) PRIMARY KEY,
    username VARCHAR(120) NOT NULL,
    team_name VARCHAR(120) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_users_team FOREIGN KEY (team_name) 
        REFERENCES teams(team_name) 
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE INDEX idx_users_team_active ON users(team_name, is_active);
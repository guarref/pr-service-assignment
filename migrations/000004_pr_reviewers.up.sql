CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(100) NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Составной первичный ключ (один юзер не может быть дважды на одном PR)
    PRIMARY KEY (pull_request_id, user_id),
    
    -- Foreign Key на PR
    CONSTRAINT fk_pr_reviewers_pr FOREIGN KEY (pull_request_id) 
        REFERENCES pull_requests(pull_request_id) 
        ON DELETE CASCADE,
    
    -- Foreign Key на пользователя
    CONSTRAINT fk_pr_reviewers_user FOREIGN KEY (user_id) 
        REFERENCES users(user_id) 
        ON DELETE CASCADE
);

-- Индекс для быстрого поиска PR'ов по ревьюверу (для endpoint /users/getReview)
CREATE INDEX idx_pr_reviewers_user_id ON pr_reviewers(user_id);

-- Индекс для поиска ревьюверов по PR(для получения списка ревьюверов конкретного PR)
CREATE INDEX idx_pr_reviewers_pr_id ON pr_reviewers(pull_request_id);

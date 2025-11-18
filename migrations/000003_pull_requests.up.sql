CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(120) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(120) NOT NULL,
    status VARCHAR(6) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMPTZ NULL,
    
    -- Ограничение на статус (только OPEN или MERGED)
    CONSTRAINT chk_pr_status CHECK (status IN ('OPEN', 'MERGED')),
    
    -- Foreign Key на автора (пользователь)
    CONSTRAINT fk_pr_author FOREIGN KEY (author_id) 
        REFERENCES users(user_id) 
        ON DELETE CASCADE
);

-- Индексы
CREATE INDEX idx_pr_author_id ON pull_requests(author_id);
CREATE INDEX idx_pr_status ON pull_requests(status);
CREATE INDEX idx_pr_created_at ON pull_requests(created_at DESC); -- для сортировки
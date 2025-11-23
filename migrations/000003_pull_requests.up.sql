CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(120) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(120) NOT NULL,
    status VARCHAR(6) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMPTZ NULL,

    CONSTRAINT chk_pr_status CHECK (status IN ('OPEN', 'MERGED')),

    CONSTRAINT fk_pr_author FOREIGN KEY (author_id) 
        REFERENCES users(user_id) 
        ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(100) NOT NULL,
    user_id VARCHAR(120) NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (pull_request_id, user_id),

    CONSTRAINT fk_pr_reviewers_pr FOREIGN KEY (pull_request_id) 
        REFERENCES pull_requests(pull_request_id) 
        ON DELETE CASCADE,

    CONSTRAINT fk_pr_reviewers_user FOREIGN KEY (user_id) 
        REFERENCES users(user_id) 
        ON DELETE CASCADE
);

CREATE INDEX idx_pr_reviewers_user_id ON pr_reviewers(user_id);
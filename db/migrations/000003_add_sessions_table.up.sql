CREATE TABLE sessions(
    id              VARCHAR(255)    NOT NULL PRIMARY KEY,
    user_email      VARCHAR(255)    NOT NULL,
    refresh_token   VARCHAR(512)    NOT NULL,
    is_revoked      BOOLEAN         NOT NULL DEFAULT FALSE,
    created_at      timestamptz     DEFAULT now(),
    expires_at      timestamptz
);
-- +goose Up
-- 会員（EDS-MEM-1）: 論理削除・チャレンジレスポンス・セッション

CREATE TABLE members (
    id              UUID PRIMARY KEY,
    login_id        TEXT NOT NULL,
    nickname        TEXT NOT NULL,
    salt            BYTEA NOT NULL,
    secret          BYTEA NOT NULL,
    status          TEXT NOT NULL,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT members_status_check CHECK (status IN ('active', 'deleted')),
    CONSTRAINT members_deleted_at_check CHECK (
        (status = 'active' AND deleted_at IS NULL)
        OR (status = 'deleted' AND deleted_at IS NOT NULL)
    )
);

-- login_id / nickname は全行一意。
-- 論理削除時にマスク置換するため、旧値での再登録が可能（EDS-MEM-1）。
CREATE UNIQUE INDEX members_login_id_uidx ON members (login_id);
CREATE UNIQUE INDEX members_nickname_uidx ON members (nickname);

CREATE TABLE member_auth_challenges (
    id              TEXT PRIMARY KEY,
    login_id        TEXT NOT NULL,
    member_id       UUID REFERENCES members (id),
    salt            BYTEA NOT NULL,
    challenge       BYTEA NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    consumed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX member_auth_challenges_expires_idx
    ON member_auth_challenges (expires_at);

CREATE TABLE member_sessions (
    token_hash      BYTEA PRIMARY KEY,
    member_id       UUID NOT NULL REFERENCES members (id),
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX member_sessions_member_id_idx
    ON member_sessions (member_id);

CREATE INDEX member_sessions_expires_idx
    ON member_sessions (expires_at);

-- +goose Down
DROP TABLE IF EXISTS member_sessions;
DROP TABLE IF EXISTS member_auth_challenges;
DROP TABLE IF EXISTS members;

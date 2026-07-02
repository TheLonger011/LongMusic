CREATE TABLE favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    track_id BIGINT not null,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, track_id)
);
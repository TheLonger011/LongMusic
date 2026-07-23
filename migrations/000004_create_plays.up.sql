CREATE TABLE plays (
    id BIGSERIAL PRIMARY KEY NOT NULL ,
    user_id BIGINT,
    track_id BIGINT,
    played_at TIMESTAMP DEFAULT NOW()
)
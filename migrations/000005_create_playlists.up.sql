CREATE TABLE playlists (
    id BIGSERIAL PRIMARY KEY NOT NULL ,
    user_id BIGINT NOT NULL ,
    name VARCHAR(125) NOT NULL ,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE playlist_tracks (
    playlist_id BIGINT NOT NULL ,
    track_id BIGINT NOT NULL,
    PRIMARY KEY (playlist_id, track_id)
)
CREATE TABLE tracks (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    album VARCHAR(255),
    duration INT,
    file_path VARCHAR(255),
    created_at TIMESTAMP
);

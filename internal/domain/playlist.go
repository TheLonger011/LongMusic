package domain

import "time"

type Playlist struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IsPublic  bool      `db:"is_public" json:"is_public"`

	TrackCount int64 `json:"track_count" db:"track_count"`
}

type PlaylistTrack struct {
	PlaylistID int64 `json:"playlist_id" db:"playlist_id"`
	TrackID    int64 `json:"track_id" db:"track_id"`
}

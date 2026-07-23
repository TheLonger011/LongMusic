package domain

import "time"

type Play struct {
	ID       int64     `db:"id" json:"id"`
	UserID   int64     `db:"user_id" json:"user_id"`
	TrackID  int64     `db:"track_id" json:"track_id"`
	PlayedAt time.Time `db:"played_at" json:"played_at"`
}

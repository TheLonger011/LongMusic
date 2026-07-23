package domain

type Artist struct {
	ID           int64  `json:"id" db:"id"`
	Name         string `db:"name" json:"name"`
	TrackCount   int    `db:"track_count" json:"track_count"`
	CoverTrackID int64  `db:"cover_track_id" json:"cover_track_id"`
}

package repository

import (
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/jmoiron/sqlx"
)

type PlayRepository struct {
	db *sqlx.DB
}

func NewPlayRepository(db *sqlx.DB) *PlayRepository {
	return &PlayRepository{db: db}
}

func (t *PlayRepository) Create(play domain.Play) (int64, error) {
	var id int64
	err := t.db.QueryRow(
		"INSERT INTO plays (user_id, track_id) VALUES ($1, $2) RETURNING id",
		play.UserID,
		play.TrackID,
	).Scan(&id)
	return id, err
}

func (t *PlayRepository) GetByUserID(userID int64) ([]domain.Play, error) {
	var plays []domain.Play
	err := t.db.Select(&plays, "SELECT * FROM plays WHERE user_id = $1", userID)
	return plays, err
}

func (t *PlayRepository) GetHistoryWithTracks(userID int64, limit int) ([]domain.Track, error) {
	var plays []domain.Track
	err := t.db.Select(&plays, `
SELECT tracks.* FROM tracks
        JOIN plays ON tracks.id = plays.track_id
        WHERE plays.user_id = $1
        ORDER BY plays.played_at DESC
        LIMIT $2`, userID, limit)
	return plays, err
}

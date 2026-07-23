package repository

import (
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/jmoiron/sqlx"
)

type FavoriteRepository struct {
	db *sqlx.DB
}

func NewFavoriteRepository(db *sqlx.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

func (t *FavoriteRepository) Add(userID, trackID int64) error {
	_, err := t.db.Exec(`
	INSERT INTO favorites (user_id, track_id) VALUES ($1, $2)`, userID, trackID)
	return err
}

func (t *FavoriteRepository) Remove(userID, trackID int64) error {
	_, err := t.db.Exec(
		`
		DELETE FROM favorites WHERE user_id = $1 AND track_id = $2`, userID, trackID)
	return err
}

func (t *FavoriteRepository) GetByUserID(userID int64) ([]domain.Track, error) {
	var tracks []domain.Track
	err := t.db.Select(&tracks, `
        SELECT tracks.* FROM tracks
        JOIN favorites ON tracks.id = favorites.track_id
        WHERE favorites.user_id = $1
        ORDER BY favorites.created_at DESC
    `, userID)
	return tracks, err
}

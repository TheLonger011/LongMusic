package repository

import (
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/jmoiron/sqlx"
)

type ChartRepository struct {
	db *sqlx.DB
}

func NewChartRepository(db *sqlx.DB) *ChartRepository {
	return &ChartRepository{db: db}
}

func (t *ChartRepository) GetTopTracks(limit int) ([]domain.Track, error) {
	var tracks []domain.Track
	err := t.db.Select(&tracks, `SELECT tracks.* FROM tracks
JOIN plays ON tracks.id = plays.track_id
GROUP BY tracks.id
ORDER BY COUNT(*) DESC
LIMIT $1`, limit)
	return tracks, err
}

func (t *ChartRepository) GetTopTracksToday(limit int) ([]domain.Track, error) {
	var tracks []domain.Track
	err := t.db.Select(&tracks, `
SELECT tracks.* FROM tracks
JOIN plays ON tracks.id = plays.track_id
WHERE plays.created_at > NOW() - INTERVAL '24 hours'
GROUP BY tracks.id
ORDER BY COUNT(*) DESC
LIMIT $1`, limit)
	return tracks, err
}

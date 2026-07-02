package repository

import (
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/jmoiron/sqlx"
)

type TrackRepository struct {
	db *sqlx.DB
}

func NewTrackRepository(db *sqlx.DB) *TrackRepository {
	return &TrackRepository{db: db}
}

func (t *TrackRepository) Create(track domain.Track) (int64, error) {
	var id int64
	err := t.db.QueryRow(
		`INSERT INTO tracks (name,artist,album,duration,file_path) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		track.Name, track.Artist, track.Album, track.Duration, track.FilePath,
	).Scan(&id)
	return id, err
}

func (t *TrackRepository) GetByID(id int64) (domain.Track, error) {
	var track domain.Track
	err := t.db.Get(&track, "SELECT * FROM tracks WHERE id = $1", id)
	return track, err
}

func (t *TrackRepository) GetAll() ([]domain.Track, error) {
	var tracks []domain.Track
	err := t.db.Select(&tracks, "SELECT * FROM tracks")
	return tracks, err
}

func (t *TrackRepository) Search(query string) ([]domain.Track, error) {
	var tracks []domain.Track
	err := t.db.Select(&tracks, "SELECT * FROM tracks WHERE name ILIKE $1 OR artist ILIKE $1", "%"+query+"%")
	return tracks, err

}

func (t *TrackRepository) GetArtists() ([]domain.Artist, error) {
	var artists []domain.Artist
	err := t.db.Select(&artists, "SELECT artist as name, COUNT(*) as track_count, MIN(id) as cover_track_id FROM tracks GROUP BY artist")
	return artists, err
}

func (t *TrackRepository) GetTracksByArtistID(id int64) ([]domain.Track, error) {
	var tracks []domain.Track
	err := t.db.Select(&tracks, `
		SELECT * FROM tracks 
		WHERE artist = (
    		SELECT artist FROM tracks 
    		GROUP BY artist 
    		HAVING MIN(id) = $1
)
`, id)
	return tracks, err
}

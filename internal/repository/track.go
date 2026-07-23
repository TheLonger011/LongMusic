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
	err := t.db.Select(&artists, `
		SELECT
			hashtext(trim(a.artist_name))::bigint as id,
			trim(a.artist_name) as name,
			COUNT(*) as track_count,
			MIN(t.id) as cover_track_id
		FROM tracks t,
			unnest(string_to_array(t.artist, ',')) AS a(artist_name)
		WHERE trim(a.artist_name) <> ''
		GROUP BY trim(a.artist_name)
	`)
	return artists, err
}

func (t *TrackRepository) GetTracksByArtistID(id int64) ([]domain.Track, error) {
	var tracks []domain.Track
	err := t.db.Select(&tracks, `
		SELECT DISTINCT t.* FROM tracks t,
			unnest(string_to_array(t.artist, ',')) AS a(artist_name)
		WHERE hashtext(trim(a.artist_name))::bigint = $1
	`, id)
	return tracks, err
}

func (t *TrackRepository) GetArtistNameByID(id int64) (string, error) {
	var name string
	err := t.db.Get(&name, `
		SELECT trim(a.artist_name) FROM tracks t,
			unnest(string_to_array(t.artist, ',')) AS a(artist_name)
		WHERE hashtext(trim(a.artist_name))::bigint = $1
		LIMIT 1
	`, id)
	return name, err
}

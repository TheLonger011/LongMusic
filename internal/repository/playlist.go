package repository

import (
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/jmoiron/sqlx"
)

type PlaylistRepository struct {
	db *sqlx.DB
}

func NewPlaylistRepository(db *sqlx.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

func (t *PlaylistRepository) GetByUserID(userID int64) ([]domain.Playlist, error) {
	var playlists []domain.Playlist
	err := t.db.Select(&playlists, `
		SELECT p.*, COUNT(pt.track_id) AS track_count
		FROM playlists p
		LEFT JOIN playlist_tracks pt ON pt.playlist_id = p.id
		WHERE p.user_id = $1
		GROUP BY p.id
		ORDER BY p.created_at DESC`, userID)
	return playlists, err
}

func (t *PlaylistRepository) GetByID(playlistID int64) (domain.Playlist, error) {
	var playlist domain.Playlist
	err := t.db.Get(&playlist, "SELECT * FROM playlists WHERE id = $1", playlistID)
	return playlist, err
}

func (t *PlaylistRepository) Create(playlist domain.Playlist) (int64, error) {
	var id int64
	err := t.db.QueryRow(`
INSERT INTO playlists (user_id, name) VALUES ($1, $2) RETURNING id`,
		playlist.UserID, playlist.Name).Scan(&id)
	return id, err
}

func (t *PlaylistRepository) AddTrack(playlistID int64, trackID int64) error {
	_, err := t.db.Exec(
		`
INSERT INTO playlist_tracks (playlist_id, track_id) VALUES ($1, $2)`, playlistID, trackID)
	return err
}

func (t *PlaylistRepository) GetTracks(playlistID int64) ([]domain.Track, error) {
	var tracks []domain.Track

	err := t.db.Select(&tracks, `
	SELECT tracks.* FROM tracks
	JOIN playlist_tracks ON tracks.id = playlist_tracks.track_id
	WHERE playlist_tracks.playlist_id = $1`, playlistID)
	return tracks, err
}

func (t *PlaylistRepository) RemoveTrack(playlistID int64, trackID int64) error {
	_, err := t.db.Exec(`
		DELETE FROM playlist_tracks WHERE playlist_id = $1 AND track_id = $2`, playlistID, trackID)
	return err

}

func (t *PlaylistRepository) Public(playlistID int64) error {
	_, err := t.db.Exec(`
UPDATE playlists SET is_public = true WHERE id = $1`, playlistID)
	return err
}

func (t *PlaylistRepository) Private(playlistID int64) error {
	_, err := t.db.Exec(`
UPDATE playlists SET is_public = false WHERE id = $1`, playlistID)
	return err
}

func (t *PlaylistRepository) SearchPublic(query string) ([]domain.Playlist, error) {
	var playlists []domain.Playlist
	err := t.db.Select(&playlists, `
		SELECT p.*, COUNT(pt.track_id) AS track_count
		FROM playlists p
		LEFT JOIN playlist_tracks pt ON pt.playlist_id = p.id
		WHERE p.is_public = true AND p.name ILIKE $1
		GROUP BY p.id`, "%"+query+"%")
	return playlists, err
}

func (t *PlaylistRepository) Delete(playlistID int64) error {
	_, err := t.db.Exec(`DELETE FROM playlists WHERE id = $1`, playlistID)
	return err
}

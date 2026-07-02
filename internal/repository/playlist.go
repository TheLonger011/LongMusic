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
	err := t.db.Select(&playlists, "SELECT * FROM playlists WHERE user_id = $1", userID)
	return playlists, err
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

package service

import (
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/TheLonger011/LongMusic/internal/repository"
)

type PlaylistService struct {
	repo *repository.PlaylistRepository
}

func NewPlaylistService(repo *repository.PlaylistRepository) *PlaylistService {
	return &PlaylistService{repo: repo}
}

func (s *PlaylistService) Create(userID int64, name string) (int64, error) {
	playlist := domain.Playlist{
		UserID: userID,
		Name:   name,
	}
	return s.repo.Create(playlist)

}

func (s *PlaylistService) GetByUserID(userID int64) ([]domain.Playlist, error) {
	return s.repo.GetByUserID(userID)
}

func (s *PlaylistService) AddTrack(playlistID, track int64) error {
	return s.repo.AddTrack(playlistID, track)
}

func (s *PlaylistService) GetTracks(playlistID int64) ([]domain.Track, error) {
	return s.repo.GetTracks(playlistID)
}

func (s *PlaylistService) RemoveTrack(playlistID, trackID int64) error {
	return s.repo.RemoveTrack(playlistID, trackID)
}

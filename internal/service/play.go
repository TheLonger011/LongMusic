package service

import (
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/TheLonger011/LongMusic/internal/repository"
)

type PlayService struct {
	repo *repository.PlayRepository
}

func NewPlayService(repo *repository.PlayRepository) *PlayService {
	return &PlayService{repo: repo}
}

func (s *PlayService) RecordPlay(userID, trackID int64) error {
	play := domain.Play{
		UserID:  userID,
		TrackID: trackID,
	}
	_, err := s.repo.Create(play)
	return err
}

func (s *PlayService) GetHistory(userID int64) ([]domain.Play, error) {
	return s.repo.GetByUserID(userID)
}

func (s *PlayService) GetHistoryWithTracks(userID int64, limit int) ([]domain.Track, error) {
	return s.repo.GetHistoryWithTracks(userID, limit)
}

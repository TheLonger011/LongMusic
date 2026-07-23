package service

import (
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/TheLonger011/LongMusic/internal/repository"
)

type ChartService struct {
	repo *repository.ChartRepository
}

func NewChartService(repo *repository.ChartRepository) *ChartService {
	return &ChartService{repo: repo}
}

func (s *ChartService) GetTopTracks(limit int) ([]domain.Track, error) {
	return s.repo.GetTopTracks(limit)
}

func (s *ChartService) GetTopTracksToday(limit int) ([]domain.Track, error) {
	return s.repo.GetTopTracksToday(limit)
}

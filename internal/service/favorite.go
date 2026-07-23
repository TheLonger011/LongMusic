package service

import (
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/TheLonger011/LongMusic/internal/repository"
)

type FavoriteService struct {
	repo *repository.FavoriteRepository
}

func NewFavoriteService(repo *repository.FavoriteRepository) *FavoriteService {
	return &FavoriteService{repo: repo}
}

func (s *FavoriteService) Add(userID, trackID int64) error {
	return s.repo.Add(userID, trackID)
}

func (s *FavoriteService) Remove(userID, trackID int64) error {
	return s.repo.Remove(userID, trackID)
}

func (s *FavoriteService) GetByUserID(userID int64) ([]domain.Track, error) {
	return s.repo.GetByUserID(userID)
}

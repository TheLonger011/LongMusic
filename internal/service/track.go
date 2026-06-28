package service

import (
	"fmt"
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/TheLonger011/LongMusic/internal/repository"
	"io"
	"mime/multipart"
	"os"
)

type TrackService struct {
	repo *repository.TrackRepository
}

func NewTrackService(repo *repository.TrackRepository) *TrackService {
	return &TrackService{repo: repo}
}

func (s *TrackService) Upload(name, artist, album string, duration int64, file multipart.File, filename string) (int64, error) {
	err := os.MkdirAll("uploads", 0755)
	if err != nil {
		return 0, err
	}

	dst, err := os.Create(fmt.Sprintf("uploads/%s", filename))

	if err != nil {
		return 0, err
	}

	defer dst.Close()

	io.Copy(dst, file)

	track := domain.Track{
		Name:     name,
		Artist:   artist,
		Album:    album,
		Duration: duration,
		FilePath: fmt.Sprintf("uploads/%s", filename),
	}
	return s.repo.Create(track)
}

func (s *TrackService) GetByID(id int64) (domain.Track, error) {
	return s.repo.GetByID(id)
}

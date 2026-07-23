package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/TheLonger011/LongMusic/internal/repository"
	"github.com/bogem/id3v2/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tcolgate/mp3"
	"io"
	"mime/multipart"
	"os"
	"time"
)

type TrackService struct {
	repo  *repository.TrackRepository
	redis *redis.Client
}

func NewTrackService(repo *repository.TrackRepository, redis *redis.Client) *TrackService {
	return &TrackService{repo: repo, redis: redis}
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

	filePath := fmt.Sprintf("uploads/%s", filename)

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err == nil {
		if tag.Title() != "" {
			name = tag.Title()
		}
		if tag.Artist() != "" {
			artist = tag.Artist()
		}
		tag.Close()
	}

	f, _ := os.Open(filePath)
	defer f.Close()
	d := mp3.NewDecoder(f)
	var frame mp3.Frame
	var skipped int
	var totalDuration float64

	for {
		if err := d.Decode(&frame, &skipped); err != nil {
			break
		}
		totalDuration += frame.Duration().Seconds()
	}
	duration = int64(totalDuration)

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
	ctx := context.Background()
	cashKey := fmt.Sprintf("track:%d", id)

	cached, err := s.redis.Get(ctx, cashKey).Result()
	if err == nil {
		var track domain.Track
		json.Unmarshal([]byte(cached), &track)
		return track, nil
	}

	track, err := s.repo.GetByID(id)
	if err != nil {
		return domain.Track{}, err
	}

	data, _ := json.Marshal(track)
	s.redis.Set(ctx, cashKey, string(data), 5*time.Minute)
	return track, nil
}

func (s *TrackService) GetAll() ([]domain.Track, error) {
	return s.repo.GetAll()
}

func (s *TrackService) Search(query string) ([]domain.Track, error) {
	return s.repo.Search(query)
}

func (s *TrackService) GetArtists() ([]domain.Artist, error) {
	return s.repo.GetArtists()
}

func (s *TrackService) GetArtistsTracks(id int64) ([]domain.Track, error) {
	return s.repo.GetTracksByArtistID(id)
}

func (s *TrackService) GetArtistName(id int64) (string, error) {
	return s.repo.GetArtistNameByID(id)
}

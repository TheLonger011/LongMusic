package main

import (
	"fmt"
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/TheLonger011/LongMusic/internal/repository"
	"github.com/dhowden/tag"
	"github.com/jmoiron/sqlx"
	"github.com/tcolgate/mp3"
	"os"
	"path/filepath"
	"strings"
)

func seedTracks(db *sqlx.DB) {
	repo := repository.NewTrackRepository(db)

	if err := os.MkdirAll("uploads", 0755); err != nil {
		fmt.Println("Ошибка создания папки uploads:", err)
		return
	}

	entries, err := os.ReadDir("uploads")
	if err != nil {
		fmt.Println("Ошибка чтения папки uploads:", err)
		return
	}

	count := 0
	skipped := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".mp3" && ext != ".m4a" && ext != ".flac" && ext != ".wav" && ext != ".ogg" {
			continue
		}

		filePath := fmt.Sprintf("uploads/%s", entry.Name())

		var existing int
		db.QueryRow("SELECT COUNT(*) FROM tracks WHERE file_path = $1", filePath).Scan(&existing)
		if existing > 0 {
			skipped++
			continue
		}

		title, artist, album := readTags(filePath)
		duration := readDuration(filePath)

		if title == "" {
			title = strings.TrimSuffix(entry.Name(), ext)
		}
		if artist == "" {
			artist = "Неизвестен"
		}
		if artist == "" {
			artist = "Неизвестен"
		}

		track := domain.Track{
			Name:     title,
			Artist:   artist,
			Album:    album,
			Duration: int64(duration),
			FilePath: filePath,
		}

		id, err := repo.Create(track)
		if err != nil {
			fmt.Printf("Ошибка добавления трека %s: %v\n", entry.Name(), err)
			continue
		}

		fmt.Printf("✅ Добавлен трек: %s - %s (ID: %d)\n", artist, title, id)
		count++
	}

	fmt.Printf("\n📊 Итог: Добавлено треков: %d, Пропущено: %d\n", count, skipped)
}

func readDuration(filePath string) int64 {
	f, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer f.Close()

	d := mp3.NewDecoder(f)
	var frame mp3.Frame
	var skipped int
	var total float64
	for {
		if err := d.Decode(&frame, &skipped); err != nil {
			break
		}
		total += frame.Duration().Seconds()
	}
	return int64(total)
}

func readTags(filePath string) (title, artist, album string) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", ""
	}
	defer file.Close()

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return "", "", ""
	}
	return metadata.Title(), metadata.Artist(), metadata.Album()
}

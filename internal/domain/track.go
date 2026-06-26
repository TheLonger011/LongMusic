package domain

import "time"

type Track struct {
	ID        int64
	Name      string
	Artist    string
	Album     string
	Duration  int64
	FilePath  string
	CreatedAt time.Time
}

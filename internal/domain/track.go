package domain

import "time"

type Track struct {
	ID        int64      `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Artist    string     `db:"artist" json:"artist"`
	Album     string     `db:"album" json:"album"`
	Duration  int64      `db:"duration" json:"duration"`
	FilePath  string     `db:"file_path" json:"file_path"`
	CreatedAt *time.Time `db:"created_at" json:"created_at"`
}

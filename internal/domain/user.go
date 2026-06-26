package domain

import "time"

type User struct {
	ID        int64
	Email     string
	Password  string
	CreatedAt time.Time
}

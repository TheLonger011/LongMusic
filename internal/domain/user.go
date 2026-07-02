package domain

import "time"

type User struct {
	ID           int64      `db:"id"`
	Email        string     `db:"email"`
	Password     string     `db:"password"`
	CreatedAt    *time.Time `db:"created_at"`
	Login        string     `db:"login"`
	Role         string     `db:"role"`
	Subscription string     `db:"subscription"`
}

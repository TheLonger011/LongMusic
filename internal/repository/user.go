package repository

import (
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user domain.User) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		`INSERT INTO users(email, password) VALUES ($1, $2) RETURNING id`,
		user.Email, user.Password,
	).Scan(&id)

	return id, err
}

func (r *UserRepository) GetByEmail(email string) (domain.User, error) {
	var user domain.User
	err := r.db.Get(&user, `SELECT * FROM users WHERE email = $1`, email)
	return user, err
}

func (r *UserRepository) GetByID(id int64) (domain.User, error) {
	var user domain.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
	return user, err
}

package repository

import (
	"fmt"
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
		`INSERT INTO users(email, login, password) VALUES ($1, $2, $3) RETURNING id`,
		user.Email, user.Login, user.Password,
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

func (r *UserRepository) UpdateUsername(userID int64, username string) error {
	_, err := r.db.Exec(`
	UPDATE users SET login = $1 WHERE id = $2`, username, userID)
	return err
}

func (r *UserRepository) GetByLogin(login string) (domain.User, error) {
	var user domain.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE login = $1", login)
	if err != nil {
		return domain.User{}, fmt.Errorf("пользователь с логином '%s' не найден: %w", login, err)
	}
	return user, nil
}

func (r *UserRepository) UpdateAvatar(userID int64, avatarPath string) error {
	_, err := r.db.Exec(`UPDATE users SET avatar = $1 WHERE id = $2`, avatarPath, userID)
	return err
}

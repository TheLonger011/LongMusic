package service

import (
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/TheLonger011/LongMusic/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(email, password string) (int64, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	user := domain.User{
		Email:    email,
		Password: string(hashedPass),
	}
	id, err := s.repo.Create(user)
	return id, err
}

func (s *UserService) Login(email, password string) (string, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add((24 * time.Hour)).Unix(),
	})

	tokenString, err := token.SignedString([]byte("secret_key"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

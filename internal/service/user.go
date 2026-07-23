package service

import (
	"errors"
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

func (s *UserService) Register(email, login, password string) (int64, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	user := domain.User{
		Email:    email,
		Password: string(hashedPass),
		Login:    login,
	}
	id, err := s.repo.Create(user)
	return id, err
}

func (s *UserService) Login(login, password string) (string, domain.User, error) {
	user, err := s.repo.GetByLogin(login)
	if err != nil {
		return "", domain.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", domain.User{}, err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})
	tokenString, err := token.SignedString([]byte("secret_key"))
	if err != nil {
		return "", domain.User{}, err
	}
	return tokenString, user, nil
}

func (s *UserService) GetProfile(id int64) (domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) UpdateUsername(userID int64, username string) error {
	existing, err := s.repo.GetByLogin(username)
	if err == nil && existing.ID != 0 && existing.ID != userID {
		return errors.New("логин уже занят")
	}
	return s.repo.UpdateUsername(userID, username)
}

func (s *UserService) UpdateAvatar(id int64, avatar string) error {
	return s.repo.UpdateAvatar(id, avatar)
}

func (s *UserService) GetPublicProfile(login string) (domain.User, error) {
	return s.repo.GetByLogin(login)
}

package service

import (
	"context"
	"errors"

	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/models"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/repository"
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hashedPass, password string) error
}

type Service struct {
	repo   *repository.Repository
	hasher Hasher
}

func New(repo *repository.Repository, hasher Hasher) *Service {
	return &Service{
		repo:   repo,
		hasher: hasher,
	}
}

func (s *Service) CreateUser(ctx context.Context, u *models.User) (int64, error) {
	if u.Age < 18 {
		return 0, errors.New("user must be at least 18 years old")
	}
	if len(u.Name) < 2 {
		return 0, errors.New("name is too short")
	}

	hashedPassword, err := s.hasher.Hash(u.Password)
	if err != nil {
		return 0, err
	}
	u.Password = hashedPassword

	id, err := s.repo.CreateUser(ctx, u)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Service) GetUser(ctx context.Context, id int64) (*models.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *Service) UpdateUser(ctx context.Context, u *models.User) error {
	return s.repo.UpdateUser(ctx, u)
}

func (s *Service) Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	err = s.hasher.Compare(user.Password, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/cache"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/models"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/repository"
	"go.uber.org/zap"
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hashedPass, password string) error
}

type Service struct {
	repo   *repository.Repository
	cache  cache.ICache
	hasher Hasher
	log    *zap.Logger
}

func New(repo *repository.Repository, cache cache.ICache, hasher Hasher, log *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		cache:  cache,
		hasher: hasher,
		log:    log,
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
		s.log.Error("failed to hash password", zap.Error(err))
		return 0, err
	}
	u.Password = hashedPassword

	id, err := s.repo.CreateUser(ctx, u)
	if err != nil {
		s.log.Error("failed to create user in db", zap.Error(err))
		return 0, err
	}
	s.log.Info("user created", zap.Int64("id", id))
	return id, nil
}

func (s *Service) GetUser(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	key := fmt.Sprintf("user:%d", id)

	err := s.cache.Get(ctx, key, &user)
	if err == nil {
		s.log.Info("user found in cache", zap.Int64("id", id))
		return &user, nil
	}

	u, err := s.repo.GetUser(ctx, id)
	if err != nil {
		s.log.Error("failed to get user from db", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	err = s.cache.Save(ctx, key, u, time.Hour)
	if err != nil {
		s.log.Error("failed to save user to cache", zap.Error(err))
	}

	return u, nil
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

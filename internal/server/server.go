package server

import (
	"context"

	userv1 "github.com/s-usmonalizoda25/protoCinemaService/gen/user"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/models"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/service"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/token"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	log *zap.Logger
	svc *service.Service
}

func New(log *zap.Logger, svc *service.Service) *Server {
	return &Server{
		log: log,
		svc: svc,
	}
}

func (s *Server) Add(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		Age:      req.Age,
	}

	id, err := s.svc.CreateUser(ctx, user)
	if err != nil {
		if err.Error() == "email already exists" {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.CreateUserResponse{Id: id}, nil
}

func (s *Server) GetByID(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := s.svc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &userv1.GetUserResponse{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
		Age:   user.Age,
	}, nil
}

func (s *Server) Update(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	user := &models.User{
		ID:    req.Id,
		Name:  req.Name,
		Phone: req.Phone,
	}

	err := s.svc.UpdateUser(ctx, user)
	if err != nil {
		return &userv1.UpdateUserResponse{Code: 500, Message: "failed to update"}, err
	}
	return &userv1.UpdateUserResponse{Code: 200, Message: "success"}, nil
}

func (s *Server) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	user, err := s.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	accessToken, refreshToken, err := token.GenerateTokens(user.ID)
	if err != nil {
		s.log.Error("failed to generate tokens", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	return &userv1.LoginResponse{
		Id:           user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

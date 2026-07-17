package server

import (
	"context"
	"errors"
	"strings"

	userpb "github.com/s-usmonalizoda25/protoCinemaService/gen/user"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/models"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/repository"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/service"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/token"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	userpb.UnimplementedUserServiceServer
	log *zap.Logger
	svc *service.Service
}

func New(log *zap.Logger, svc *service.Service) *Server {
	return &Server{
		log: log,
		svc: svc,
	}
}

func (s *Server) Add(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		Age:      req.Age,
		Role:     1,
	}

	id, err := s.svc.CreateUser(ctx, user)
	if err != nil {
		if err.Error() == "email already exists" {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userpb.CreateUserResponse{Id: id}, nil
}

func (s *Server) GetByID(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	user, err := s.svc.GetUser(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &userpb.GetUserResponse{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
		Age:   user.Age,
		Role:  userpb.UserRole(user.Role),
	}, nil
}

func (s *Server) Update(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	user := &models.User{
		ID:    req.Id,
		Name:  req.Name,
		Phone: req.Phone,
	}

	err := s.svc.UpdateUser(ctx, user)
	if err != nil {
		return &userpb.UpdateUserResponse{Code: 500, Message: "failed to update"}, err
	}
	return &userpb.UpdateUserResponse{Code: 200, Message: "success"}, nil
}

func (s *Server) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	user, err := s.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	accessToken, refreshToken, err := token.GenerateTokens(user.ID, user.Role)
	if err != nil {
		s.log.Error("failed to generate tokens", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	return &userpb.LoginResponse{
		Id:           user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Role:         user.Role,
	}, nil
}

func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/user.v1.UserService/Add" || info.FullMethod == "/user.v1.UserService/Login" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	tokenString := strings.TrimPrefix(values[0], "Bearer ")

	claims, err := token.ValidateToken(tokenString)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	newCtx := context.WithValue(ctx, "user_id", claims.UserID)

	return handler(newCtx, req)
}

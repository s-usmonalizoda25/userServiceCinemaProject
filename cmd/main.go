package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	userv1 "github.com/s-usmonalizoda25/protoCinemaService/gen/user"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/cache"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/db"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/logger"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/repository"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/security"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/server"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	err := godotenv.Load("config/config.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	l := logger.New()
	defer l.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := os.Getenv("SERVER_PORT")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		l.Fatal("failed to listen", zap.Error(err))
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	dbPool, err := db.New(ctx, dsn)
	if err != nil {
		l.Fatal("failed to connect to db", zap.Error(err))
	}
	defer dbPool.Close()

	redisCache, err := cache.New(ctx, os.Getenv("REDIS_ADDR"))
	if err != nil {
		l.Fatal("failed to init cache", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(server.AuthInterceptor),
	)
	hasher := security.NewBcryptHasher(12)
	repo := repository.New(dbPool)

	svc := service.New(repo, redisCache, hasher, l.Logger)

	userServer := server.New(l.Logger, svc)

	userv1.RegisterUserServiceServer(grpcServer, userServer)
	reflection.Register(grpcServer)

	go func() {
		l.Info("server started", zap.String("port", port))
		if err := grpcServer.Serve(lis); err != nil {
			l.Fatal("failed to serve", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	_, cancelShutdown := context.WithTimeout(ctx, 10*time.Second)
	defer cancelShutdown()

	l.Info("shutting down server...")
	grpcServer.GracefulStop()
	l.Info("server stopped")
}

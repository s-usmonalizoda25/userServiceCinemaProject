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
	log := logger.New()
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := os.Getenv("SERVER_PORT")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
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
		log.Fatal("failed to connect to db", zap.Error(err))
	}
	defer dbPool.Close()

	grpcServer := grpc.NewServer()
	hasher := security.NewBcryptHasher(12)

	repo := repository.New(dbPool)
	svc := service.New(repo, hasher)
	userServer := server.New(log.Logger, svc)

	userv1.RegisterUserServiceServer(grpcServer, userServer)

	reflection.Register(grpcServer)

	go func() {
		log.Info("server started", zap.String("port", port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Info("shutting down server...")
	grpcServer.GracefulStop()
	log.Info("server stopped")
}

// Server is the gRPC entry point.
// Its only responsibility is wiring dependencies and starting the server.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/belayhun-arage/image-blur-service/api"
	"github.com/belayhun-arage/image-blur-service/internal/config"
	"github.com/belayhun-arage/image-blur-service/internal/queue"
	"github.com/belayhun-arage/image-blur-service/internal/storage"
	"github.com/subosito/gotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	pb.UnimplementedImageServiceServer
	queue   queue.Queue
	storage storage.Storage
}

func (s *grpcServer) QueueImage(ctx context.Context, req *pb.ImageRequest) (*pb.ImageResponse, error) {
	if req.ImageId == "" {
		return nil, status.Error(codes.InvalidArgument, "image_id must not be empty")
	}
	if err := s.queue.Enqueue(ctx, req.ImageId); err != nil {
		return nil, status.Errorf(codes.Internal, "enqueue: %v", err)
	}
	return &pb.ImageResponse{Status: "queued"}, nil
}

func (s *grpcServer) DeleteImage(ctx context.Context, req *pb.DeleteImageRequest) (*pb.DeleteImageResponse, error) {
	if req.SourceImgId == "" {
		return nil, status.Error(codes.InvalidArgument, "source_img_id must not be empty")
	}
	rows, err := s.storage.DeleteBlurredImage(ctx, req.SourceImgId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "delete: %v", err)
	}
	if rows == 0 {
		return nil, status.Errorf(codes.NotFound, "no record found for source_img_id %q", req.SourceImgId)
	}
	return &pb.DeleteImageResponse{Success: true}, nil
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := gotenv.Load(".env"); err != nil {
		log.Warn("could not load .env file", "err", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	q, err := queue.NewRedisQueue(cfg.RedisAddr)
	if err != nil {
		log.Error("failed to connect to Redis", "err", err)
		os.Exit(1)
	}

	store, err := storage.NewPostgresStorage(cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to Postgres", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Error("failed to listen", "port", cfg.GRPCPort, "err", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()
	pb.RegisterImageServiceServer(srv, &grpcServer{queue: q, storage: store})
	log.Info("gRPC server listening", "port", cfg.GRPCPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info("shutdown signal received")
		srv.GracefulStop()
	}()

	if err := srv.Serve(lis); err != nil {
		log.Error("server error", "err", err)
		os.Exit(1)
	}
}

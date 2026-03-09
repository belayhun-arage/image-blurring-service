// gRPC server that receives image queue requests
package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/belayhun-arage/image-blur-service/api"
	"github.com/belayhun-arage/image-blur-service/internal/queue"
	_ "github.com/lib/pq"
	"github.com/subosito/gotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedImageServiceServer
	db *sql.DB
}

func (s *server) QueueImage(ctx context.Context, req *pb.ImageRequest) (*pb.ImageResponse, error) {
	log.Println("Received request to queue image:", req.ImageId)
	if req.ImageId == "" {
		return nil, status.Error(codes.InvalidArgument, "image_id must not be empty")
	}
	if err := queue.Enqueue(ctx, req.ImageId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to enqueue image: %v", err)
	}
	return &pb.ImageResponse{Status: "queued"}, nil
}

func (s *server) DeleteImage(ctx context.Context, req *pb.DeleteImageRequest) (*pb.DeleteImageResponse, error) {
	if req.SourceImgId == "" {
		return nil, status.Error(codes.InvalidArgument, "source_img_id must not be empty")
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM blurred_img WHERE source_img_id = $1`, req.SourceImgId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete image: %v", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, status.Errorf(codes.NotFound, "no record found for source_img_id: %s", req.SourceImgId)
	}
	return &pb.DeleteImageResponse{Success: true}, nil
}

func main() {
	if err := gotenv.Load(".env"); err != nil {
		log.Println("Warning: could not load .env file:", err)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if err := queue.InitRedisQueue(redisAddr); err != nil {
		log.Fatalf("failed to initialize Redis queue: %v", err)
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterImageServiceServer(s, &server{db: db})
	log.Printf("gRPC server running on :%s", grpcPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down gRPC server...")
		s.GracefulStop()
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

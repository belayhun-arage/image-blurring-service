// gRPC server that receives image queue requests
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/belayhun-arage/image-blur-service/api"
	"github.com/belayhun-arage/image-blur-service/internal/queue"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedImageServiceServer
}

func (s *server) QueueImage(ctx context.Context, req *pb.ImageRequest) (*pb.ImageResponse, error) {
	fmt.Println("Received request to queue image:", req.ImageId)
	if req.ImageId == "" {
		return nil, fmt.Errorf("invalid image ID")
	}
	err := queue.Enqueue(req.ImageId)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue image: %w", err)
	}
	return &pb.ImageResponse{Status: "queued"}, nil
}

func main() {
	// Initialize Redis queue
	err := queue.InitRedisQueue("localhost:6379")
	if err != nil {
		log.Fatalf("failed to initialize Redis queue: %v", err)
	}
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterImageServiceServer(s, &server{})
	log.Println("gRPC server running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

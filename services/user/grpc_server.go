package user

import (
	"context"
	"database/sql"

	pb "github.com/genin6382/go-grpc-microservices-benchmark/pb/user"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	pb.UnimplementedUserServiceServer
	DB *sql.DB
}

func (s * Server) CheckUserExists(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	log.Infof("Received CheckUserExists request for user ID: %s", req.UserId)

	exists , err := UserExistsByID(s.DB, ctx , req.UserId )

	if err != nil {
		log.Errorf("Error checking user existence: %v", err)
		return &pb.UserResponse{}, err
	}
	return &pb.UserResponse{Exists: exists}, nil
}
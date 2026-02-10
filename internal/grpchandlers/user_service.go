package grpchandlers

import (
	"context"
	"time"

	"github.com/pasanAbeysekara/collaborative-editor/internal/auth"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
	pb "github.com/pasanAbeysekara/collaborative-editor/proto"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	Store storage.Store
}

func (s *UserServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := s.Store.CreateUser(req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.AlreadyExists, "user already exists: %v", err)
	}

	return &pb.RegisterResponse{
		Id:    user.ID,
		Email: user.Email,
	}, nil
}

func (s *UserServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.Store.GetUserByEmail(req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	token, err := auth.CreateJWT(user.ID, 24*time.Hour)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not create token: %v", err)
	}

	return &pb.LoginResponse{Token: token}, nil
}

func (s *UserServiceServer) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.User, error) {
	user, err := s.Store.GetUserByEmail(req.Email)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.User{
		Id:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}, nil
}

func (s *UserServiceServer) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.User, error) {
	user, err := s.Store.GetUserByID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.User{
		Id:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}, nil
}

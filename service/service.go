package service

import (
	"context"
	"log"

	"github.com/golang/protobuf/ptypes/empty"

	. "github.com/violin0622/grpc-apple/api/operation"
	. "github.com/violin0622/grpc-apple/api"
)

// protoc-gen-go-grpc v1.20 之后, 实现服务的时候必须嵌入 UnimplementedAppleServiceServer 以保证能够向后兼容.
type AppleService struct {
	UnimplementedAppleServiceServer
}

func (*AppleService) DescribeApple(ctx context.Context, req *DescribeAppleRequest) (*Apple, error) {
	return &Apple{}, nil
}
func (*AppleService) CreateApple(ctx context.Context, req *CreateAppleRequest) (*Apple, error) {
	log.Println(req)
	return &Apple{}, nil
}
func (*AppleService) UpdateApple(ctx context.Context, req *UpdateAppleRequest) (*Apple, error) {
	return &Apple{}, nil
}
func (*AppleService) ModifyApple(ctx context.Context, req *ModifyAppleRequest) (*Apple, error) {
	return &Apple{}, nil
}
func (*AppleService) DestroyApple(ctx context.Context, req *DestroyAppleRequest) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}


package grpcCilent

import (
	"context"
	pb "core/app/grpc"
	"time"
	authClient "march-inventory/cmd/app/common/grpc"
)

func GetPermission(shopIds string) (*pb.GetPermissionResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return authClient.AuthGrpcClient.GetPermission(ctx, &pb.GetPermissionrRequest{ShopsId: shopIds})
}

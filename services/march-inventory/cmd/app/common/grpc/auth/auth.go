package grpcCilent

import (
	"context"
	pb "core/app/grpc"
	authClient "march-inventory/cmd/app/common/grpc"

	// "time"

	"google.golang.org/grpc/metadata"
)

func GetPermission(shopIds string, auth string) (*pb.GetPermissionResponse, error) {
	md := metadata.New(map[string]string{"authorization": auth})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	return authClient.AuthGrpcClient.GetPermission(ctx, &pb.GetPermissionrRequest{ShopsId: shopIds})
}

package grpcCilent

import (
	"context"
	pb "core/app/grpc"
	grpcClient "march-inventory/cmd/app/common/grpc"
	"time"
)

func HelperTest() (*pb.InventoryName, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return grpcClient.UserGrpcClient.HelperTest(ctx, nil)
}

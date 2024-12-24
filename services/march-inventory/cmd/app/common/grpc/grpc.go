package grpcCilent

import (
	"log"

	pb "core/app/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var AuthGrpcClient pb.AuthGrpcServiceClient

func Init() *grpc.ClientConn {

	creds := insecure.NewCredentials()

	conn, err := grpc.NewClient("localhost:5005", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	AuthGrpcClient = pb.NewAuthGrpcServiceClient(conn)

	return conn
}

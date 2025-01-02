package grpcCilent

import (
	"log"

	pb "core/app/grpc"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var AuthGrpcClient pb.AuthGrpcServiceClient
var UserGrpcClient pb.UserGrpcServiceClient

func Init() []*grpc.ClientConn {
	authGrpcUrl := viper.GetString("auth.grpc.url")
	userGrpcUrl := viper.GetString("user.grpc.url")
	connectionUrls := []string{authGrpcUrl, userGrpcUrl}

	connections := []*grpc.ClientConn{}

	for _, connectionUrl := range connectionUrls {
		creds := insecure.NewCredentials()
		conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(creds))
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		connections = append(connections, conn)
	}

	AuthGrpcClient = pb.NewAuthGrpcServiceClient(connections[0])
	UserGrpcClient = pb.NewUserGrpcServiceClient(connections[1])

	return connections
}

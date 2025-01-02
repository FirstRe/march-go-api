package userManagementService

import (
	"context"
	// "core/app/helper"

	pb "core/app/grpc"
)

const ClassNameGrpc = "grpc_user"

type Server struct {
	pb.UnimplementedUserGrpcServiceServer
}

func (s *Server) HelperTest(context.Context, *pb.EmptyRequest) (*pb.InventoryName, error) {
	// logctx := helper.LogContext(ClassNameGrpc, "HelperTest")
	// logctx.Logger(in.ShopsId, "[error-api] ShopsIdssss")

	response := pb.InventoryName{
		Id:   "123",
		Name: "1234",
	}

	return &response, nil
}

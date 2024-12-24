package authService

import (
	"context"
	"core/app/helper"

	pb "core/app/grpc"
	gormDb "march-auth/cmd/app/common/gorm"
	"march-auth/cmd/app/graph/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const ClassNameGrpc = "grpc_auth"

type Server struct {
	pb.UnimplementedAuthGrpcServiceServer
}

func (s *Server) GetPermission(ctx context.Context, in *pb.GetPermissionrRequest) (*pb.GetPermissionResponse, error) {
	logctx := helper.LogContext(ClassNameGrpc, "UpsertInventory")
	logctx.Logger(in.ShopsId, "[error-api] ShopsIdssss")
	shop := model.Shop{}
	if err := gormDb.Repos.Where("id = ?", in.ShopsId).First(&shop).Error; err != nil {
		logctx.Logger(err.Error(), "[error-api] SignOut")
		return nil, status.Errorf(codes.Unimplemented, "method GetPermission not implemented")
	}

	response := pb.GetPermissionResponse{
		Shop: &pb.Shop{
			Id: shop.ID,
		},
	}

	return &response, nil
}

package auth

import (
	"context"
	"core/app/common/jwt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func authInterceptor(ctx context.Context) (context.Context, error) {
	url := os.Getenv("UAM_URL")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "missing metadata")
	}

	tokens := md["authorization"]

	if len(tokens) == 0 {
		return nil, status.Errorf(codes.PermissionDenied, "missing metadata")
	}
	token := tokens[0]

	validate, err := jwt.Verify(token)
	if err != nil || !validate.Valid {
		return nil, status.Errorf(codes.PermissionDenied, "Unauthorized")
	}

	userInfo, err := jwt.Decode(token)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "Unauthorized")
	}

	deviceId, err := DeviceIdCheckPost(url, token)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "Unauthorized Device")
	}

	if userInfo.DeviceId != *deviceId {
		return nil, status.Errorf(codes.PermissionDenied, "Unauthorized Device2")
	}

	return ctx, nil
}

func UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	newCtx, err := authInterceptor(ctx)
	if err != nil {
		return nil, err
	}
	return handler(newCtx, req)
}

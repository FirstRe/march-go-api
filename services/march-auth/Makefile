# Makefile

# Define variables
PRISMA_CMD = go run github.com/steebchen/prisma-client-go

init:
	go mod init

tidy:
	go mod tidy	

dev:
	nodemon --exec "go run" cmd/app/server.go --signal SIGTERM

start:
	go run cmd/app/server.go

gen:
	go run github.com/99designs/gqlgen generate

proto-gen:
	protoc -I=../../packages/core/app/proto --go_out=. --go-grpc_out=. ../../packages/core/app/proto/auth/*.proto

prisma-push:
	go run github.com/steebchen/prisma-client-go db push --schema=./cmd/app/prisma/schema.prisma

prisma-gen:
	go run github.com/steebchen/prisma-client-go generate --schema=./cmd/app/prisma/schema.prisma

prisma-migrate:
	go run github.com/steebchen/prisma-client-go migrate dev --schema=./cmd/app/prisma/schema.prisma
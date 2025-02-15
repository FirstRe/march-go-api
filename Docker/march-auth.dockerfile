# Use the official Golang image as a base image
FROM golang:1.22.1-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules and vendor files for the march-auth service
# COPY ./services/march-auth/go.mod ./services/march-auth/go.sum ./

COPY ../ .

# Copy the core package into the container
# COPY ./packages/core /app/packages/core

WORKDIR /app/packages/core

RUN go mod tidy

WORKDIR /app/services/march-auth

COPY services/march-auth .

# Download dependencies
RUN go mod tidy

# Copy the source code for the march-auth service
# COPY ./services/march-auth /app


WORKDIR /app/services/march-auth/cmd/app

RUN go build -o server server.go
# Build the Go application


# RUN go build -o /app/march-auth ./cmd/app

# Use a smaller image to run the built application
FROM alpine:latest

# Install required dependencies (e.g., ca-certificates)
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder image
COPY --from=builder /app /app/

COPY services/march-auth/.env /app/services/march-auth/cmd/app/
# COPY services/march-auth/config.yaml  /app/services/march-auth/cmd/app

WORKDIR /app/services/march-auth/cmd/app
# Expose the port on which your service runs
EXPOSE 8080

# Command to run the application
CMD ["./server"]

# Use the official Go image as a base image
FROM golang:1.22.1 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.work ./
COPY packages/core/go.mod packages/core/go.sum ./packages/core/
COPY services/march-auth/go.mod services/march-auth/go.sum ./services/march-auth/

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go binaries
RUN go build -o services/march-auth/cmd/app ./services/march-auth/cmd/app

# Start a new stage from scratch
FROM debian:bullseye-slim

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary files from the previous stage
COPY --from=builder /app/services/march-auth ./services/march-auth

# Ensure the binary has execution permissions
RUN chmod +x services/march-auth/cmd/app/app

EXPOSE 8080

# Command to run the executable
CMD ["./services/march-auth/cmd/app/app"]

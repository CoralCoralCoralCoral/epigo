# Stage 1: Build
FROM golang:1.23 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app-binary main.go

# Stage 2: Run
FROM alpine:latest

# Install required certificates for HTTPS (if needed)
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/app-binary .

# Command to run the executable
CMD ["./app-binary"]

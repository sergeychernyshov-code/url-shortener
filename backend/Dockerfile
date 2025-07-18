# Builder stage: use official Go 1.24.4 image to build the binary
FROM golang:1.24.4 as builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies first (cache optimization)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY main.go ./

# Build the Go binary for Linux (no CGO)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go


# Final stage: Use AWS Lambda Go runtime base image
FROM public.ecr.aws/lambda/go:1

# Copy the binary from builder stage into Lambda runtime
COPY --from=builder /app/main /var/task/

# Set the Lambda entrypoint
CMD ["main"]
# Stage 1: Builder - use official Golang 1.24 image
FROM golang:1.24 AS builder

WORKDIR /app

# Copy go.mod and go.sum first to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY main.go ./

# Build static Linux binary (CGO disabled)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

# Stage 2: Runtime - AWS Lambda Go runtime
FROM public.ecr.aws/lambda/go:1

# Copy binary from builder
COPY --from=builder /app/main ${LAMBDA_TASK_ROOT}/main

# Set CMD to your executable
CMD ["main"]
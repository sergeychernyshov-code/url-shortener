# Stage 1: Builder - use official Go 1.24 SDK image
FROM golang:1.24 as builder

WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY main.go ./

# Build statically linked binary for Linux AMD64 (CGO disabled)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

# Stage 2: Runtime - use AWS Lambda Go runtime image (Go 1.x)
FROM public.ecr.aws/lambda/go:1

# Copy the binary from the builder stage
COPY --from=builder /app/main ${LAMBDA_TASK_ROOT}/main

# Set the CMD to your built executable
CMD ["main"]
# Builder stage: AWS Lambda Go builder (Amazon Linux 2 + Go)
FROM public.ecr.aws/lambda/go-builder:1 as builder

WORKDIR /app

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY main.go ./

# Build static Linux binary for Lambda
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

# Runtime stage: AWS Lambda Go runtime
FROM public.ecr.aws/lambda/go:1

# Copy the built binary from builder stage
COPY --from=builder /app/main ${LAMBDA_TASK_ROOT}/main

# Set entrypoint to the Lambda handler
CMD ["main"]
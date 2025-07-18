# Stage 1: Builder
FROM golang:1.20 as builder

WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build a static Linux binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

# Stage 2: Runtime (AWS Lambda AL2023 provided runtime)
FROM public.ecr.aws/lambda/provided:al2023

# Copy the built binary into Lambda task root
COPY --from=builder /app/main ${LAMBDA_TASK_ROOT}/main

# Set the Lambda handler
CMD ["main"]
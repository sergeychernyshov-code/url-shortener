# Stage 1: Builder - Use official Go image to build the binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY main.go ./

# Build statically linked Go binary named 'bootstrap' for AWS Lambda
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

# Stage 2: Runtime - AWS Lambda Custom Runtime (provided.al2023)
FROM public.ecr.aws/lambda/provided:al2023

# Copy the 'bootstrap' binary to Lambda's task root
COPY --from=builder /app/bootstrap ${LAMBDA_TASK_ROOT}/bootstrap

# Ensure 'bootstrap' is executable (usually not needed, but good practice)
RUN chmod +x ${LAMBDA_TASK_ROOT}/bootstrap
# Stage 1: Builder - Use official Go image to build the binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build statically linked binary for Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

# Stage 2: Runtime using provided.al2023
FROM public.ecr.aws/lambda/provided:al2023

# Copy the compiled binary to Lambda's expected bootstrap location
COPY --from=builder /app/bootstrap /var/task/bootstrap

# Set handler name for Lambda container image expectations
CMD ["bootstrap"]
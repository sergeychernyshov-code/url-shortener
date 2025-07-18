# Builder: use official Golang 1.24.4 image
FROM golang:1.24.4 AS builder

WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY main.go ./

# Build statically linked binary named 'bootstrap'
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

# Runtime: AWS Lambda provided.al2023 base image (Amazon Linux 2023)
FROM public.ecr.aws/lambda/provided:al2023

# Copy the bootstrap binary to /var/task/bootstrap
COPY --from=builder /app/bootstrap /var/task/bootstrap

# Make sure bootstrap is executable
RUN chmod +x /var/task/bootstrap

# No CMD needed for provided runtime (Lambda will run /var/task/bootstrap by default)
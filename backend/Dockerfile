# Stage 1: Builder
FROM amazonlinux:2 as builder

# Clean yum cache
RUN yum clean all

# Update packages and install dependencies
RUN yum update -y
RUN yum install -y gcc glibc-devel curl tar make

# Download Go 1.24.4
RUN curl -LO https://go.dev/dl/go1.24.4.linux-amd64.tar.gz

# Extract Go to /usr/local
RUN tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz

# Remove the tarball
RUN rm go1.24.4.linux-amd64.tar.gz

# Add Go to PATH
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY main.go ./

# Build statically linked binary for Linux AMD64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

# Stage 2: Runtime
FROM public.ecr.aws/lambda/go:1

# Copy the binary from builder stage
COPY --from=builder /app/main ${LAMBDA_TASK_ROOT}/main

# Set entrypoint
CMD ["main"]
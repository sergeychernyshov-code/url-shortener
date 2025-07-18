# Stage 1: Builder - Amazon Linux 2 + Go 1.24 install
FROM amazonlinux:2 as builder

# Install required packages: gcc, glibc-devel, curl, tar
RUN yum update -y && \
    yum install -y gcc glibc-devel curl tar make && \
    curl -LO https://go.dev/dl/go1.24.4.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz && \
    rm go1.24.4.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /app

# Copy go.mod and go.sum, download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY main.go ./

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

# Stage 2: Lambda runtime
FROM public.ecr.aws/lambda/go:1

COPY --from=builder /app/main ${LAMBDA_TASK_ROOT}/main

CMD ["main"]
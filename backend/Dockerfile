# Stage 1: Builder (Amazon Linux 2 with Go installed)
FROM public.ecr.aws/amazonlinux/amazonlinux:2 as builder

# Install Go compiler
RUN yum install -y golang git

WORKDIR /app

# Copy source code to builder
COPY main.go .

# Build the static Go binary
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go

# Stage 2: Lambda runtime image
FROM public.ecr.aws/lambda/go:1

# Copy the compiled binary from builder
COPY --from=builder /app/main ${LAMBDA_TASK_ROOT}/main

# Lambda entrypoint
CMD ["main"]
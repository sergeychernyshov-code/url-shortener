FROM public.ecr.aws/amazonlinux/amazonlinux:2 as builder

RUN yum install -y golang git

WORKDIR /app

# Copy Go module files first for caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy rest of the source code
COPY main.go .

# Build the Go binary
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go

FROM public.ecr.aws/lambda/go:1

COPY --from=builder /app/main ${LAMBDA_TASK_ROOT}/main

CMD ["main"]
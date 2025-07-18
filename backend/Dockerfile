FROM public.ecr.aws/lambda/go:1

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY main.go ./

# Build the binary for Linux with CGO disabled (compatible with Lambda runtime)
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go

# Set the CMD to your binary (Lambda entrypoint)
CMD [ "main" ]
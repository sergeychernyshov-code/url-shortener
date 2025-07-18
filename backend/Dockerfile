FROM public.ecr.aws/lambda/go:1

# Copy source code
COPY main.go .

# Install Go compiler
RUN yum install -y golang

# Build the Go binary
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go

# Lambda entrypoint
CMD [ "main" ]
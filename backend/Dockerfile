# Use AWS Lambda Go base image
FROM public.ecr.aws/lambda/go:1

# Copy executable to Lambda working directory
COPY main ${LAMBDA_TASK_ROOT}/main

# Set the CMD to the executable
CMD ["main"]


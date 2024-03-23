# Use an official Go runtime as a base image
FROM golang:1.18 as builder

# Install git.
# Git is required for fetching the dependencies.
RUN apt-get update && apt-get install -y git

# Accept the repository URL as a build argument.
# This allows you to specify which repository to clone when building the image.
ARG GITHUB_REPO
ARG BRANCH=main

# Set the working directory in the container
WORKDIR /app

# Clone the specified repository and checkout the specified branch (defaults to main)
RUN git clone -b ${BRANCH} http://github.com/${GITHUB_REPO} .

# Fetch dependencies using go mod.
RUN go mod download
RUN go mod verify

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -v -o server

# Use a Docker multi-stage build to create a lean production image.
FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .

# Run the server executable.
CMD ["./server"]

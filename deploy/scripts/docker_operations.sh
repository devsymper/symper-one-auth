#!/bin/bash

set -e

# Function to handle errors
function handle_error() {
    local error_message=$1
    echo "Error: $error_message"
    exit 1
}

# Function to perform Docker login
function docker_login() {
    echo "Attempting Docker login..."
    if ! echo "$DOCKER_REGISTRY_PWD" | docker login -u "$DOCKER_REGISTRY_USER" --password-stdin localhost:5000; then
        handle_error "Docker login failed"
    fi
    echo "Docker login successful"
}

# Function to build Docker image
function docker_build() {
    echo "Building Docker image..."
    if ! docker build --build-arg GITHUB_TOKEN="$GITHUB_TOKEN" -t "localhost:5000/$IMAGE_TAG_NAME" .; then
        docker_cleanup
        handle_error "Docker build failed"
    fi
    echo "Docker build image [$IMAGE_TAG_NAME] successful"
}

# Function to push Docker image
function docker_push() {
    echo "Pushing Docker image..."
    if ! docker push "localhost:5000/$IMAGE_TAG_NAME"; then
        docker_cleanup
        handle_error "Docker push failed"
    fi
    echo "Docker push image [$IMAGE_TAG_NAME] successful"
}

# Function to clean up Docker images
function docker_cleanup() {
    echo "Cleaning up Docker images..."
    docker image rm "localhost:5000/$IMAGE_TAG_NAME" || true
    docker image prune -f || true
    echo "Docker cleanup completed"
}

# Main execution
# Check required environment variables
if [ -z "$DOCKER_REGISTRY_USER" ] || [ -z "$DOCKER_REGISTRY_PWD" ]; then
    handle_error "Docker registry credentials not provided"
fi

if [ -z "$IMAGE_TAG_NAME" ]; then
    handle_error "IMAGE_TAG_NAME not provided"
fi

if [ -z "$GITHUB_TOKEN" ]; then
    handle_error "GITHUB_TOKEN not provided"
fi

# Execute Docker operations
docker_login
docker_build
docker_push
docker_cleanup
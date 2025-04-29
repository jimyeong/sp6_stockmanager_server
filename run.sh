#!/bin/bash

# Default to development if no argument is provided
ENV=development

# Set the environment variable
# export APP_ENV=$ENV

# Run the Go project with the corresponding .env.{ENV} file
echo "Running with environment: $ENV"
ENV=$ENV go run main.go
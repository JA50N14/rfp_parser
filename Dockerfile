# Stage 1 — Build the Go function
FROM golang:1.24.1-bookworm AS builder

WORKDIR /src

# Install git (needed for go mod download if using GitHub deps)
RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build Linux binary (Azure runs Linux)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o function


# Stage 2 — Azure Functions Runtime
FROM mcr.microsoft.com/azure-functions/base:4.0

# Install poppler-utils for pdftotext
RUN apt-get update && apt-get install -y poppler-utils ca-certificates && rm -rf /var/lib/apt/lists/*

# Set working directory required by Azure Functions
WORKDIR /home/site/wwwroot

# Copy built binary from builder stage
COPY --from=builder /src/function .

# Copy function metadata (host.json + function folders)
COPY host.json ./
COPY rfp_parser_timer_function ./rfp_parser_timer_function

# Set Azure Functions environment variables
ENV AzureWebJobsScriptRoot=/home/site/wwwroot \
    AzureFunctionsJobHost__Logging__Console__IsEnabled=true

# Copy the kpiDefinitions.json file
COPY --from=builder /src/parser ./parser/

# Azure Functions runtime starts automatically



# #/////////////////////////////////////////////////////////////////////////////////
# #This DockerFile below is used to create an Docker Image Locally and then create a Container locally - For Testing the code inside a Container Locally

# # Stage 1: Build the Go binary
# FROM golang:1.24.1-alpine AS builder

# # Install necessary build tools
# RUN apk add --no-cache git build-base

# # Set the working directory
# WORKDIR /app

# # Copy go.mod and go.sum first (for caching dependencies)
# COPY go.mod go.sum ./
# RUN go mod download

# # Copy the rest of the source code
# COPY . .

# # Build the Go binary for Linux
# RUN go build -o rfp-parser .

# # Stage 2: Create the minimal runtime image
# FROM alpine:3.18

# # Install poppler-utils (provides pdftotext)
# RUN apk add --no-cache poppler-utils bash ca-certificates

# # Set the working directory
# WORKDIR /app

# # Copy the Go binary from the builder stage
# COPY --from=builder /app/rfp-parser .

# # Make the binary executable
# RUN chmod +x rfp-parser

# # Copy the kpiDefinitions.json file
# COPY --from=builder /app/parser ./parser

# # Copy the .env file - ***Remove this line when creating a Production Image***
# COPY .env .env

# # Set the entrypoint (for Azure Functions Timer, this will be executed on container start)
# ENTRYPOINT ["./rfp-parser"]

# # To run this local docker image:
# # docker run -d -p 8080:8080 -e ENV=local --name parser-test -v $(pwd)/certs:/app/certs:ro parser-test:latest
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
FROM mcr.microsoft.com/azure-functions/go:4-go1.24

# Install poppler-utils for pdftotext
RUN apt-get update && apt-get install -y poppler-utils ca-certificates && rm -rf /var/lib/apt/lists/*

# Set working directory required by Azure Functions
WORKDIR /home/site/wwwroot

# Copy built binary from builder stage
COPY --from=builder /src/function .

# Copy function metadata (host.json + function folders)
COPY host.json ./
COPY timer_function ./timer_function

# Set Azure Functions environment variables
ENV AzureWebJobsScriptRoot=/home/site/wwwroot \
    AzureFunctionsJobHost__Logging__Console__IsEnabled=true

# Azure Functions runtime starts automatically

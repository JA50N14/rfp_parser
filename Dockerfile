# Stage 1: Build the Go binary
FROM golang:1.24.1-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git build-base

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first (for caching dependencies)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary for Linux
RUN go build -o rfp-parser .

# Stage 2: Create the minimal runtime image
FROM alpine:3.18

# Install poppler-utils (provides pdftotext)
RUN apk add --no-cache poppler-utils bash ca-certificates

# Set the working directory
WORKDIR /app

# Copy the Go binary from the builder stage
COPY --from=builder /app/rfp-parser .

# Make the binary executable
RUN chmod +x rfp-parser

# Set the entrypoint (for Azure Functions Timer, this will be executed on container start)
ENTRYPOINT ["./rfp-parser"]

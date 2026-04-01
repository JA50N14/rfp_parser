# Stage 1 — Build the Go app
FROM golang:1.24.1-bookworm AS builder

WORKDIR /src

# Install git (needed for go mod download if using GitHub deps)
RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o parserbinary

# Stage 2 — Minimal runtime
FROM debian:bookworm-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y poppler-utils ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the binary
COPY --from=builder /src/parserbinary ./parserbinary

# Copy the parser folder (so kpiDefinitions.json is available)
COPY --from=builder /src/parser ./parser/

# Make sure the binary is executable
RUN chmod +x ./parserbinary

# Run the binary
CMD ["./parserbinary"]



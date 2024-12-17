FROM golang:1.22 AS builder

# Set environment variables for Go build
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /app

# Copy Go modules and source code
COPY . .
RUN go mod download

# Build the server binary
RUN go build -o server ./cmd/main.go

# Stage 2: Final container
FROM registry.access.redhat.com/ubi8/ubi-minimal

# Set working directory and ensure it is writable for USER 1001
WORKDIR /app
RUN mkdir -p /app && chmod -R 777 /app

# Copy server binary
COPY --from=builder /app/server /app/server

# Copy the internal Containerfile to a location where the builder expects it
COPY Containerfile.vddk /app/Containerfile.vddk

# Set user
USER 1001

# Expose port and entrypoint
EXPOSE 8443
ENTRYPOINT ["/app/server"]

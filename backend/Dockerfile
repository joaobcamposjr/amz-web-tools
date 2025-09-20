# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Production stage
FROM oraclelinux:8-slim

# Install Oracle Instant Client dependencies
RUN microdnf install -y libaio && \
    microdnf clean all

# Create oracle directory
RUN mkdir -p /opt/oracle/instantclient_21_13

# Copy Oracle Instant Client (you'll need to download and place this file)
# COPY oracle-instantclient-basic-linux.x64-21.13.0.0.0dbru.zip /opt/oracle/

# Install Oracle Instant Client (uncomment when you have the file)
# RUN cd /opt/oracle && \
#     unzip oracle-instantclient-basic-linux.x64-21.13.0.0.0dbru.zip && \
#     rm oracle-instantclient-basic-linux.x64-21.13.0.0.0dbru.zip

# Set Oracle environment variables
ENV LD_LIBRARY_PATH=/opt/oracle/instantclient_21_13:$LD_LIBRARY_PATH
ENV TNS_ADMIN=/opt/oracle/instantclient_21_13/network/admin

# Create app directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Create var.sh for Oracle client initialization
RUN mkdir -p /opt/oracle/var.sh

# Copy environment file (will be mounted in docker-compose)
# COPY .env .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]

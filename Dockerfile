# ======================
# 1) Builder Stage
# ======================
FROM golang:1.21-alpine AS builder

# Install git (often needed for go mod downloads)
RUN apk add --no-cache git

WORKDIR /app

# Copy all your source code into the container
COPY . ./

# Initialize modules (if no local go.mod) but won't fail if it already exists
RUN go mod init github.com/kartikeya55555/fetch-assignment || true
RUN go mod tidy

# Build a single binary from your unified main (e.g. cmd/main.go)
RUN go build -o /app/fetch-assignment cmd/main.go

# ======================
# 2) Final Stage
# ======================
FROM alpine:3.18

# Copy the compiled binary from the builder stage
COPY --from=builder /app/fetch-assignment /usr/local/bin/

# Expose port 8080 if your service listens there
EXPOSE 8080

# By default, run the combined binary
ENTRYPOINT ["fetch-assignment"]

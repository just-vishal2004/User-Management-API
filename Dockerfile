# ─── Stage 1: Builder ─────────────────────────────────────────────────────────
# Use the official Go image to compile the application.
# We use a separate build stage so the final image doesn't
# contain the Go compiler, source code, or build tools.
FROM golang:1.26-alpine AS builder

# Install git — required by some Go modules during download
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first and download modules.
# Docker caches each layer — if go.mod and go.sum haven't changed,
# this layer is reused and modules aren't re-downloaded on every build.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary.
# CGO_ENABLED=0 — disables CGO for a fully static binary (no external libs needed)
# GOOS=linux    — target Linux (the container OS), even if building on Mac
# -ldflags="-w -s" — strips debug info, reduces binary size significantly
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o bin/ainyx-backend \
    cmd/server/main.go

# ─── Stage 2: Runner ──────────────────────────────────────────────────────────
# Use a minimal Alpine image for the final container.
# Alpine is only ~5MB vs ~300MB for the full Go image.
FROM alpine:latest

# Install CA certificates — required for HTTPS connections
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user to run the application.
# Running as root inside a container is a security risk.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy only the compiled binary from the builder stage.
# The final image has NO source code, NO Go compiler, NO build tools.
COPY --from=builder /app/bin/ainyx-backend .

# Switch to the non-root user
USER appuser

# Expose the port the app listens on
EXPOSE 3000

# Run the binary
CMD ["./ainyx-backend"]
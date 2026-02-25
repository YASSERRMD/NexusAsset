FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies required for migrating and building
RUN apk add --no-cache git

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the main server executable
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go
# Build the migration tool executable
RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate/main.go
# Build the seeder executable
RUN CGO_ENABLED=0 GOOS=linux go build -o seeder ./db/seed/seed.go

# Final minimal image
FROM alpine:3.19

WORKDIR /app

# Install CA certificates and bash
RUN apk add --no-cache ca-certificates bash curl

# Create a non-root user
RUN adduser -D staticuser
USER staticuser

# Copy over executables
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/seeder .
COPY --from=builder /app/config.yaml .
# We also need the migration files if we run migration on start
COPY --from=builder /app/db/migrations ./db/migrations

EXPOSE 8080

# To run migrations, you can override the command, but by default we run the server
CMD ["./main"]

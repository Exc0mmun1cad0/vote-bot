# Build stage
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /bot

# Copy Go module files
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy source code into the container
COPY . .

# Build app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o bot ./cmd/bot/main.go


# Run stage
FROM alpine:3.21.3 AS runner

# Update packages withing the image
RUN apk update && apk upgrade

# use non-root user
RUN adduser -D appuser
USER appuser

# Set the working directory inside the container
WORKDIR /bot

# Copy binary from buld stage
COPY --from=builder /bot/bot .

# Copy app config into the container 
COPY .env .env

EXPOSE 3302

# Run app itself
CMD ["./bot"]

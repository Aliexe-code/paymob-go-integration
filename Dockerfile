FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /paymob-demo ./cmd/server

# Production stage
FROM alpine:latest

WORKDIR /app

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary
COPY --from=builder /paymob-demo .

# Copy templates (embedded in binary but kept for reference)
COPY --from=builder /app/internal/views/templates ./templates/

# Copy static files
COPY --from=builder /app/static ./static

# Expose port
EXPOSE 3000

# Environment variables
ENV PORT=3000
ENV PAYMOBI_API_KEY=your_api_key
ENV PAYMOBI_MERCHANT_ID=your_merchant_id

# Run the application
CMD ["./paymob-demo"]

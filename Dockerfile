FROM golang:1.23 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o hn-bot .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -tags prod -o hn-bot .

FROM gcr.io/distroless/static-debian12

# Copy the binary
WORKDIR /app
COPY --from=builder /app/hn-bot .

# Run the application
CMD ["./hn-bot"]

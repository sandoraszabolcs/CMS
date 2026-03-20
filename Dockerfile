FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /server .
COPY frontend/ ./frontend/
COPY migrations/ ./migrations/
EXPOSE 8080
CMD ["./server"]

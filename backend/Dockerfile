# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /server ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata wget
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /server /app/server

RUN mkdir -p /app/data

ENV DATA_DIR=/app/data
ENV PORT=8099
EXPOSE 8099

CMD ["/app/server"]

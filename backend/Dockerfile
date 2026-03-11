# ---- Build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ciptr-backend .

# ---- Run stage ----
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/ciptr-backend .

# schema.sql is embedded in the binary via go:embed, no need to copy it

EXPOSE 8080

CMD ["./ciptr-backend"]

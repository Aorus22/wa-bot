# Stage 1: Build (Gunakan Alpine agar kompatibel dengan runtime)
FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o wa-bot

FROM alpine:latest

WORKDIR /root/

RUN apk add --no-cache sqlite-libs

COPY --from=builder /app/wa-bot .

RUN chmod +x ./wa-bot

CMD ["./wa-bot"]

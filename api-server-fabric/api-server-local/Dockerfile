FROM golang:1.24.3 AS builder

WORKDIR /app

COPY go.mod ./
COPY src/ ./src/
RUN go mod tidy

CMD ["go", "run", "./src/."]

# CMD ["./go-huma-api-server"]

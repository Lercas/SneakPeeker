FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o sneakpeeker .

FROM alpine:3.20

WORKDIR /root/
COPY --from=builder /app/sneakpeeker /usr/local/bin/sneakpeeker

ENTRYPOINT ["sneakpeeker"]
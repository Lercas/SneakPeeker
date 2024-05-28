FROM golang:1.21.5-alpine AS builder
WORKDIR /app
COPY go.mod ./

RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o sneakpeeker .


FROM gcr.io/distroless/static-debian11
WORKDIR /root/
COPY --from=builder /app/sneakpeeker .

ENTRYPOINT ["./sneakpeeker"]

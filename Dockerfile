FROM golang:alpine AS builder

WORKDIR /build

ADD go.mod .
COPY . .
RUN go build -o webhook ./cmd/webhook/main.go


FROM alpine

WORKDIR /build
COPY --from=builder /build/webhook /build/webhook

CMD ["./webhook"]
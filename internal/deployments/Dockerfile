# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/vitalis ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    addgroup -S vitalis && adduser -S vitalis -G vitalis

WORKDIR /app

COPY --from=builder /out/vitalis ./vitalis

USER vitalis

EXPOSE 8080

ENTRYPOINT ["./vitalis"]

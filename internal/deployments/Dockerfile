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
COPY --from=builder /src/internal/templates ./internal/templates
COPY --from=builder /src/internal/static ./internal/static
COPY --from=builder /src/cmd/scss/bin ./cmd/scss/bin

USER vitalis

EXPOSE 8080

ENTRYPOINT ["./vitalis"]

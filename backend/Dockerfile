# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20

RUN adduser -D -H -u 10001 appuser \
	&& mkdir -p /app \
	&& chown -R appuser:appuser /app

WORKDIR /app

COPY --from=build --chown=appuser:appuser /out/server /app/server

USER appuser

ENV ADDR=:8080

EXPOSE 8080

CMD ["/app/server"]

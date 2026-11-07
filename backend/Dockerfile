# syntax=docker/dockerfile:1
#
# EXPERIMENT: intentionally old Go/Alpine for Trivy failure testing. Do not merge to main.

FROM golang:1.18.10-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.15

RUN adduser -D -H -u 10001 appuser \
	&& mkdir -p /app/data \
	&& chown -R appuser:appuser /app

WORKDIR /app

COPY --from=build --chown=appuser:appuser /out/server /app/server
COPY --chown=appuser:appuser data/messages.csv /app/data/messages.csv

USER appuser

ENV ADDR=:8080
ENV DATA_CSV_PATH=/app/data/messages.csv

EXPOSE 8080

CMD ["/app/server"]

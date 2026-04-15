# syntax=docker/dockerfile:1

FROM node:22-alpine AS web-builder
WORKDIR /build/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS go-builder
WORKDIR /build
COPY go.mod ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM alpine:3.21 AS runtime
RUN addgroup -S pbgw && adduser -S -G pbgw pbgw
WORKDIR /app
COPY --from=go-builder /out/server /app/server
COPY config/ /app/config/
COPY --from=web-builder /build/web/dist /app/web/dist

EXPOSE 8080
USER pbgw
ENTRYPOINT ["/app/server"]


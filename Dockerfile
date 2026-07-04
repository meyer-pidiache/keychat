FROM node:22-alpine AS js-builder
WORKDIR /build/web
RUN corepack enable && corepack prepare pnpm@10 --activate
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm run build

FROM golang:1.26-alpine AS go-builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=js-builder /build/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -o /keychat ./cmd/keychat/...

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
RUN mkdir -p /web/static /web/dist /data && adduser -D -S -h /data keychat
USER keychat
WORKDIR /data
COPY --chown=keychat:keychat --from=go-builder /keychat /usr/local/bin/keychat
COPY --chown=keychat:keychat --from=go-builder /build/web/static /web/static
COPY --chown=keychat:keychat --from=go-builder /build/web/dist /web/dist
EXPOSE 8080
ENV PORT=8080 DB_PATH=/data/keychat.db WEB_DIR=/web
CMD ["keychat"]

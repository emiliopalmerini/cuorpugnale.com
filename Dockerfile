FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tini tzdata && \
    addgroup -S cp && adduser -S -G cp -H -s /sbin/nologin cp

ENV TZ=Europe/Rome

WORKDIR /app

COPY --from=build /out/server /app/server
RUN chown -R cp:cp /app

USER cp

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/server"]

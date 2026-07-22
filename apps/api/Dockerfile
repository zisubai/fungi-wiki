FROM golang:1.26-alpine AS builder
WORKDIR /src/apps/api
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/fungi-api ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && addgroup -S fungi && adduser -S -G fungi fungi
COPY --from=builder /out/fungi-api /usr/local/bin/fungi-api
USER fungi
EXPOSE 8080
ENTRYPOINT ["fungi-api"]

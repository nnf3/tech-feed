FROM golang:1.24-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/feed ./cmd/server

FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /bin/feed /bin/feed

EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=5s --retries=10 --start-period=10s \
  CMD curl -fsS http://localhost:8080/health || exit 1

CMD ["/bin/feed"]

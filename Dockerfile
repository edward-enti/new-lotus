# SonLotus — multi-stage build. Go version parity: golang:1.23.12
# matches go.mod (go 1.23.12). Tailwind CLI is downloaded pinned in
# the build stage (PHASE2-PINS.md) — never the host ~/local/bin binary.
# No binary is committed; both binaries are built here.
FROM golang:1.23.12-bookworm AS build
WORKDIR /src
ARG TAILWIND_VERSION=3.4.17
ARG TAILWIND_SHA256=7d24f7fa191d2193b78cd5f5a42a6093e14409521908529f42d80b11fde1f1d4
RUN curl -fsSL -o /usr/local/bin/tailwindcss \
      https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/tailwindcss-linux-x64 \
    && echo "${TAILWIND_SHA256}  /usr/local/bin/tailwindcss" | sha256sum -c - \
    && chmod +x /usr/local/bin/tailwindcss
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make css && go build -o /out/server ./cmd/server && go build -o /out/migrator ./cmd/migrator

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd -r sonlotus && useradd -r -g sonlotus sonlotus \
    && mkdir -p /app/data/uploads && chown -R sonlotus:sonlotus /app
WORKDIR /app
COPY --from=build /out/server /out/migrator /app/
USER sonlotus
ENV PORT=":8080" DATA_DIR="data" DB_PATH="data/sonlotus.db" APP_ENV="production"
EXPOSE 8080
CMD ["/app/server"]

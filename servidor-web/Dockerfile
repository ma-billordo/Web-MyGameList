# Este Dockerfile tiene 3 etapas. Cada servicio del docker-compose.yml elige
# cuál construir con 'target:'.
#   base   -> Go, sqlc y dependencias del proyecto (sin el código)
#   test   -> base + código                          (servicio 'test')
#   server -> base + Hurl + código                   (servicio 'server')


# ---------- base: lo común a los dos servicios ----------
FROM golang:1.26 AS base

WORKDIR /app

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

COPY go.mod go.sum ./
RUN go mod download


# ---------- test: para 'make test' ----------
FROM base AS test

COPY . .


# ---------- server: para 'make run' y 'make api-test' ----------
FROM base AS server

# Instalar Hurl (se hace antes de copiar el código, para que Docker no lo
# vuelva a instalar cada vez que cambia un archivo del proyecto)
ARG HURL_VERSION=8.0.1
RUN curl -fsSL -o /tmp/hurl.deb \
      https://github.com/Orange-OpenSource/hurl/releases/download/${HURL_VERSION}/hurl_${HURL_VERSION}_$(dpkg --print-architecture).deb \
 && apt-get update \
 && apt-get install -y /tmp/hurl.deb \
 && rm -rf /var/lib/apt/lists/* /tmp/hurl.deb

COPY . .
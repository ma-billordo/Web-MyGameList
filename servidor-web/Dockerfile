# Imagen única para los servicios 'test' y 'server' del docker-compose.yml.
# Incluye Go, sqlc y templ, y las dependencias del proyecto.
FROM golang:1.27.4

WORKDIR /app

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN go install github.com/a-h/templ/cmd/templ@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .
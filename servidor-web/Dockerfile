# Imagen única para los servicios 'test' y 'server' del docker-compose.yml.
# Incluye Go, sqlc y templ, y las dependencias del proyecto.
FROM golang:1.27.4

WORKDIR /app

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1100

COPY go.mod go.sum ./
RUN go mod download

COPY . .
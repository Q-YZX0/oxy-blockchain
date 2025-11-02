# Dockerfile para Oxy•gen Blockchain (Go)
# Multi-stage build para imagen optimizada

# ============================================
# Stage 1: Build
# ============================================
FROM golang:1.21-alpine AS builder

# Instalar dependencias del sistema necesarias
RUN apk add --no-cache git make gcc musl-dev

# Establecer directorio de trabajo
WORKDIR /build

# Copiar go.mod y go.sum
COPY go/go.mod go/go.sum ./
RUN go mod download

# Copiar código fuente
COPY go/ ./

# Construir el binario
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o oxy-blockchain ./cmd/oxy-blockchain/main.go

# ============================================
# Stage 2: Runtime
# ============================================
FROM alpine:latest

# Instalar dependencias runtime
RUN apk --no-cache add ca-certificates tzdata

# Crear usuario no-root
RUN addgroup -g 1000 blockchain && \
    adduser -D -u 1000 -G blockchain blockchain

# Establecer directorio de trabajo
WORKDIR /app

# Copiar binario desde builder
COPY --from=builder /build/oxy-blockchain .

# Crear directorio de datos
RUN mkdir -p /app/data && \
    chown -R blockchain:blockchain /app

# Usar usuario no-root
USER blockchain

# Variables de entorno por defecto
ENV OXY_DATA_DIR=/app/data
ENV OXY_LOG_LEVEL=info
ENV BLOCKCHAIN_API_ENABLED=true
ENV BLOCKCHAIN_API_HOST=localhost
ENV BLOCKCHAIN_API_PORT=8080

# Exponer puerto del API REST (solo localhost en producción)
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Comando por defecto
CMD ["./oxy-blockchain"]


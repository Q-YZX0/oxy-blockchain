# Guía de Testing - Oxy•gen Blockchain

## Tests Disponibles

### Tests Unitarios

```bash
# Ejecutar todos los tests
go test ./... -v

# Tests específicos
go test ./internal/consensus -v
go test ./internal/crypto -v
go test ./internal/storage -v
```

### Tests de Integración

```bash
# Tests de flujo completo ABCI
go test ./internal/consensus -run TestABCIApp -v

# Tests de verificación de firmas
go test ./internal/crypto -run TestVerify -v
```

### Coverage

```bash
# Generar coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Testing Manual

### 1. Iniciar el nodo

```bash
# Con Make
make run

# O directamente
go run ./cmd/oxy-blockchain/main.go
```

### 2. Verificar Health Check

```bash
curl http://localhost:8080/health
```

### 3. Verificar Métricas

```bash
curl http://localhost:8080/metrics
```

### 4. Probar Endpoints del API

```bash
# Obtener altura actual
curl http://localhost:8080/api/v1/blocks/latest

# Obtener transacción (requiere hash real)
curl http://localhost:8080/api/v1/transactions/0x...
```

## Testing con Docker

### Build y Run

```bash
# Build
make docker-build

# Run
make docker-run
```

### Docker Compose (desarrollo)

```bash
cd ..
docker-compose up -d
```

### Docker Compose (producción)

```bash
cd ..
docker-compose -f docker-compose.prod.yml up -d
```

## Variables de Entorno para Testing

```bash
export OXY_DATA_DIR=./test-data
export OXY_LOG_LEVEL=debug
export OXY_LOG_JSON=false
export BLOCKCHAIN_API_ENABLED=true
export BLOCKCHAIN_API_PORT=8080
export BLOCKCHAIN_API_HOST=localhost
```

## Troubleshooting

### Error: "cometbft: command not found"
- CometBFT no está instalado o no está en PATH
- El código intentará crear configuración manual si no encuentra el comando

### Error: "port already in use"
- Cambiar `BLOCKCHAIN_API_PORT` a otro puerto (ej: 8081)

### Error: "permission denied" en data/
- Verificar permisos del directorio de datos
- Usar `chmod 755` en el directorio


# Listo para Testing - Oxy•gen Blockchain

## ✅ Lo que tenemos implementado

### Core Blockchain
- ✅ Motor de consenso (CometBFT)
- ✅ Motor de ejecución EVM (go-ethereum compatible)
- ✅ Storage persistente (LevelDB)
- ✅ Sistema de validadores con staking
- ✅ Slash automático por faltar bloques
- ✅ Rate limiting y protección anti-spam

### API REST Local
- ✅ Endpoints completos:
  - `GET /health` - Health check
  - `GET /metrics` - Métricas
  - `GET /api/v1/blocks/latest` - Último bloque
  - `GET /api/v1/blocks/{height}` - Bloque por altura
  - `GET /api/v1/transactions/{hash}` - Transacción por hash
  - `GET /api/v1/accounts/{address}` - Estado de cuenta
  - `POST /api/v1/submit-tx` - Enviar transacción

### Red P2P
- ✅ Integración con oxygen-sdk mesh network
- ✅ MeshBridge para comunicación Go ↔ TypeScript
- ✅ Query handler para queries descentralizadas
- ✅ Discovery automático de validadores

### Tests Unitarios Existentes
- ✅ `internal/consensus/abci_app_test.go` - Tests ABCI
- ✅ `internal/crypto/signer_test.go` - Tests de firmas
- ✅ `internal/storage/db_test.go` - Tests de storage
- ✅ `internal/network/mesh_test.go` - Tests de mesh

## 🎯 Para Testear Necesitamos

### 1. Dependencias del Sistema

#### Requerido
- ✅ **Go 1.21+** - Ya configurado en `go.mod`
- ⚠️ **CometBFT** (opcional) - El código intenta crearlo manualmente si no está instalado

#### Instalación de CometBFT (Opcional)
```bash
# Opción 1: Usando Go
go install github.com/cometbft/cometbft/cmd/cometbft@latest

# Opción 2: Descargar binario
# Ver: https://github.com/cometbft/cometbft/releases
```

**Nota**: El código puede funcionar sin CometBFT instalado si genera la configuración manualmente.

### 2. Configuración Mínima

#### Variables de Entorno
```bash
# Mínimo necesario para testing
export OXY_DATA_DIR=./test-data
export OXY_CHAIN_ID=test-chain
export OXY_LOG_LEVEL=debug
export BLOCKCHAIN_API_ENABLED=true
export BLOCKCHAIN_API_PORT=8080
export BLOCKCHAIN_API_HOST=localhost
```

#### Archivo `.env` (opcional)
```bash
OXY_DATA_DIR=./test-data
OXY_CHAIN_ID=test-chain
OXY_LOG_LEVEL=debug
OXY_LOG_JSON=false

OXY_MESH_ENDPOINT=ws://localhost:3001

BLOCKCHAIN_API_ENABLED=true
BLOCKCHAIN_API_PORT=8080
BLOCKCHAIN_API_HOST=localhost
```

### 3. Preparación para Testing

#### Paso 1: Instalar Dependencias Go
```bash
cd oxy-blockchain/go
go mod download
go mod tidy
```

#### Paso 2: Compilar
```bash
make build
# o
go build -o bin/oxy-blockchain ./cmd/oxy-blockchain/main.go
```

#### Paso 3: Verificar Tests Unitarios
```bash
# Ejecutar todos los tests
make test
# o
go test ./... -v

# Tests con coverage
make test-coverage
```

## 🧪 Tests que Podemos Ejecutar Ahora

### Tests Unitarios
```bash
# Tests de consenso (ABCI)
go test ./internal/consensus -v

# Tests de firmas criptográficas
go test ./internal/crypto -v

# Tests de storage
go test ./internal/storage -v

# Tests de network/mesh
go test ./internal/network -v
```

### Tests de Integración

#### Test 1: Iniciar Nodo y Verificar Health
```bash
# Terminal 1: Iniciar nodo
make run

# Terminal 2: Verificar health
curl http://localhost:8080/health

# Debería retornar:
# {
#   "status": "healthy",
#   "storage": true,
#   "evm": true,
#   "consensus": true,
#   "mesh": true
# }
```

#### Test 2: Verificar API REST
```bash
# Verificar métricas
curl http://localhost:8080/metrics

# Obtener último bloque
curl http://localhost:8080/api/v1/blocks/latest

# Obtener bloque por altura (si existe)
curl http://localhost:8080/api/v1/blocks/0
```

#### Test 3: Enviar Transacción
```bash
# Preparar transacción (necesita hash y firma válidos)
curl -X POST http://localhost:8080/api/v1/submit-tx \
  -H "Content-Type: application/json" \
  -d '{
    "hash": "0x...",
    "from": "0x...",
    "to": "0x...",
    "value": "1000000000000000000",
    "gasLimit": 21000,
    "gasPrice": "1000000000",
    "nonce": 0,
    "signature": "..."
  }'
```

### Tests Manuales Recomendados

#### Test Básico: Nodo en Modo Standalone
1. ✅ Compilar el nodo
2. ✅ Iniciar con configuración mínima
3. ✅ Verificar que el API REST está disponible
4. ✅ Verificar health check
5. ✅ Verificar que CometBFT se inicializa correctamente
6. ✅ Verificar que el EVM se inicia correctamente

#### Test Intermedio: Múltiples Nodos
1. ⚠️ Iniciar 2+ nodos con diferentes puertos
2. ⚠️ Conectar a mesh network (requiere oxygen-sdk corriendo)
3. ⚠️ Verificar discovery de validadores
4. ⚠️ Verificar transmisión de transacciones entre nodos

#### Test Avanzado: Red Completa
1. ⚠️ Configurar 3+ validadores
2. ⚠️ Producir bloques
3. ⚠️ Verificar consenso
4. ⚠️ Verificar slash automático

## ⚠️ Lo que Falta para Testing Completo

### 1. Herramientas de Testing

#### Scripts de Testing
- [ ] Script para generar transacciones de prueba con firmas válidas
- [ ] Script para crear múltiples nodos de prueba
- [ ] Script para testing de red completa

#### Utilities de Testing
- [ ] Helper para crear transacciones firmadas
- [ ] Helper para crear bloques de prueba
- [ ] Mock del mesh network para testing sin oxygen-sdk

### 2. Integración con oxygen-sdk

#### Para Testing Completo Necesitamos:
- [ ] **oxygen-sdk corriendo** en el puerto 3001 (WebSocket relay)
  - O configurar el nodo para funcionar sin mesh endpoint
- [ ] Script de setup para testing híbrido (Go + TypeScript)

### 3. Tests de Integración Completos

#### Tests Pendientes:
- [ ] Test de flujo completo: Transacción → Bloque → Receipt
- [ ] Test de múltiples nodos produciendo bloques
- [ ] Test de consenso con validadores
- [ ] Test de slash automático
- [ ] Test de rate limiting
- [ ] Test de queries por mesh network

### 4. Docker para Testing

#### Configuración Docker:
- [ ] Dockerfile actualizado y funcionando
- [ ] docker-compose para testing con múltiples nodos
- [ ] Scripts de testing en Docker

## 🚀 Comenzar Testing Ahora

### Opción 1: Testing Mínimo (Solo Go)

```bash
# 1. Instalar dependencias
cd oxy-blockchain/go
go mod download

# 2. Ejecutar tests unitarios
go test ./... -v

# 3. Compilar
go build -o bin/oxy-blockchain ./cmd/oxy-blockchain/main.go

# 4. Iniciar nodo (sin mesh, solo API REST)
export OXY_DATA_DIR=./test-data
export OXY_MESH_ENDPOINT=""  # Deshabilitar mesh para testing simple
export BLOCKCHAIN_API_ENABLED=true
export BLOCKCHAIN_API_PORT=8080

./bin/oxy-blockchain

# 5. En otra terminal, probar API
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/blocks/latest
```

### Opción 2: Testing con Mesh (Requiere oxygen-sdk)

```bash
# 1. Iniciar oxygen-sdk mesh network (en otra terminal)
cd oxygen-sdk
npm install
npm run dev
# Configurar WebSocket relay en puerto 3001

# 2. Iniciar nodo blockchain (en otra terminal)
cd oxy-blockchain/go
export OXY_MESH_ENDPOINT=ws://localhost:3001
./bin/oxy-blockchain

# 3. Probar integración
# El nodo debería conectarse a la mesh network
```

## 📋 Checklist Pre-Testing

Antes de comenzar testing, verificar:

- [ ] Go 1.21+ instalado: `go version`
- [ ] Dependencias Go instaladas: `go mod download`
- [ ] Tests unitarios pasan: `go test ./...`
- [ ] El nodo compila: `go build ./cmd/oxy-blockchain/main.go`
- [ ] Variables de entorno configuradas (mínimas)
- [ ] Puerto 8080 disponible (o cambiar `BLOCKCHAIN_API_PORT`)
- [ ] (Opcional) oxygen-sdk instalado y corriendo para testing de mesh

## 🎯 Próximos Pasos Sugeridos

1. **Primero**: Ejecutar tests unitarios para verificar que todo compila
2. **Segundo**: Iniciar nodo standalone y probar API REST básico
3. **Tercero**: Crear script para generar transacciones de prueba
4. **Cuarto**: Configurar testing con múltiples nodos
5. **Quinto**: Integrar con oxygen-sdk para testing completo de mesh

## 📝 Notas Importantes

- **CometBFT opcional**: El código puede funcionar sin tener CometBFT instalado
- **Mesh opcional**: Puedes testear sin mesh network configurando `OXY_MESH_ENDPOINT=""`
- **API REST**: Siempre disponible si `BLOCKCHAIN_API_ENABLED=true`
- **Datos de prueba**: Usar `OXY_DATA_DIR=./test-data` para no afectar datos de producción


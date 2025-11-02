# Instrucciones de Build - Oxy•gen Blockchain

## Requisitos

### 1. Go
- **Versión**: Go 1.21 o superior
- **Descarga**: https://golang.org/dl/
- **Verificar instalación**:
  ```bash
  go version
  ```

### 2. CometBFT (Opcional para desarrollo)
- CometBFT se puede usar como binario externo o como biblioteca Go
- Si se usa como binario:
  ```bash
  # Instalar desde https://github.com/cometbft/cometbft/releases
  # O desde source:
  git clone https://github.com/cometbft/cometbft.git
  cd cometbft
  make install
  ```

### 3. Node.js (para API Layer)
- **Versión**: Node.js 18+
- **Descarga**: https://nodejs.org/

## Build del Core (Go)

### 1. Instalar dependencias

```bash
cd oxy-blockchain/go
go mod download
go mod tidy
```

### 2. Compilar

```bash
# Usando Makefile
make build

# O manualmente
go build -o bin/oxy-blockchain ./cmd/oxy-blockchain/main.go
```

### 3. Ejecutar

```bash
# Configurar variables de entorno (opcional)
export OXY_DATA_DIR=./data
export OXY_CHAIN_ID=oxy-gen-chain
export OXY_MESH_ENDPOINT=ws://localhost:3001

# Ejecutar
./bin/oxy-blockchain
```

## Build del API Layer (TypeScript)

### 1. Instalar dependencias

```bash
cd oxy-blockchain/api
npm install
```

### 2. Compilar

```bash
npm run build
```

### 3. Ejecutar tests

```bash
npm test
```

## Problemas Comunes

### Error: "go: command not found"
- **Solución**: Instalar Go y agregarlo al PATH
- **Windows**: Descargar desde golang.org y seguir instalador
- **Verificar PATH**: Agregar `C:\Program Files\Go\bin` al PATH

### Error: "package github.com/cometbft/cometbft not found"
- **Solución**: 
  ```bash
  cd oxy-blockchain/go
  go mod download
  go mod tidy
  ```

### Error: "cometbft: command not found"
- **Solución**: CometBFT no está instalado como binario
- **Nota**: El código intenta usar el binario si está disponible, pero también puede inicializarse manualmente

### Error: "cannot find package"
- **Solución**: Verificar que estás en el directorio correcto (`oxy-blockchain/go`)
- Ejecutar: `go mod download` y `go mod tidy`

## Testing

### Ejecutar tests Go

```bash
cd oxy-blockchain/go
go test ./...
```

### Tests específicos

```bash
# Tests de storage
go test ./internal/storage/...

# Tests de network
go test ./internal/network/...
```

## Desarrollo

### Modo desarrollo con hot-reload

```bash
# Usar air o similar para hot-reload
# Instalar air:
go install github.com/cosmtrek/air@latest

# Ejecutar con air
cd oxy-blockchain/go
air
```

## Verificación de Build

Después de compilar, verificar que el binario existe:

```bash
ls -lh oxy-blockchain/go/bin/oxy-blockchain
# O en Windows:
dir oxy-blockchain\go\bin\oxy-blockchain.exe
```

Si el build es exitoso, deberías ver un archivo ejecutable.


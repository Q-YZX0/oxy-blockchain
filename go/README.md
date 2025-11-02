# Oxy•gen Blockchain - Core (Go)

Core de la blockchain nativa de Oxy•gen escrita en Go.

## Stack Tecnológico

- **Consenso**: CometBFT (anteriormente Tendermint Core)
- **Ejecución**: EVMone (EVM compatible)
- **Storage**: LevelDB
- **Lenguaje**: Go 1.21+

## Arquitectura

```
go/
├── main.go                    # Punto de entrada
├── internal/
│   ├── consensus/            # Motor de consenso (CometBFT)
│   ├── execution/            # Motor de ejecución (EVMone)
│   ├── storage/              # Storage de blockchain
│   └── network/              # Red P2P (integración con oxygen-sdk)
```

## Desarrollo

### Requisitos

- Go 1.21 o superior
- CometBFT
- EVMone (integración futura)

### Instalación

```bash
# Instalar dependencias
make deps
# o
go mod download
```

### Build

```bash
# Compilar
make build
# o
go build -o bin/oxy-blockchain ./main.go
```

### Ejecutar

```bash
# Ejecutar nodo
make run
# o
./bin/oxy-blockchain
```

### Configuración

Copia `.env.example` a `.env` y configura las variables necesarias:

```bash
cp .env.example .env
# Edita .env con tu configuración
```

## Integración con oxygen-sdk

El nodo se integra con `oxygen-sdk` para:
- Usar mesh network para transmisión de transacciones/bloques
- Discovery de otros validadores por la mesh
- Comunicación P2P entre nodos

## Testing

Ver `test/README.md` para guía completa de testing.

### Ejecutar Tests

```bash
# Opción 1: Script automático (Windows)
.\test\scripts\test.bat

# Opción 2: Manual
go test ./... -v

# Opción 3: Con Make
make test
```

Los tests están organizados en `test/`:
- `test/scripts/` - Scripts de testing
- `test/unit/` - Tests unitarios adicionales (futuro)
- `test/integration/` - Tests de integración (futuro)
- Tests unitarios principales están junto al código (convención Go)

## Estructura del Proyecto

```
go/
├── cmd/oxy-blockchain/    # Punto de entrada principal
├── internal/              # Código interno
│   ├── consensus/        # Motor de consenso (CometBFT)
│   ├── execution/        # Motor de ejecución (EVM)
│   ├── storage/          # Storage de blockchain
│   ├── network/          # Red P2P (integración con oxygen-sdk)
│   └── ...
├── test/                  # Tests y documentación de testing
│   ├── scripts/          # Scripts de testing
│   ├── integration/      # Tests de integración (futuro)
│   └── unit/             # Tests unitarios adicionales (futuro)
├── Makefile              # Comandos de build
└── README.md             # Este archivo
```

## Notas

- La blockchain funciona completamente sin Internet
- Usa mesh network (WiFi/LoRa) para comunicación entre nodos
- Internet/satélite solo como puente opcional cuando meshes están desconectadas
- Los tests unitarios están junto al código (convención Go estándar)


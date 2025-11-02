# Tests - Oxy•gen Blockchain

Carpeta organizada para todos los tests y documentación relacionada.

## Estructura

```
test/
├── README.md              # Este archivo
├── scripts/               # Scripts de testing
│   ├── test.bat          # Script de testing (Windows)
│   ├── test.sh           # Script de testing (Linux/Mac)
│   └── check-build.bat   # Script de verificación de compilación
├── unit/                  # Tests unitarios (futuro - cuando se organice)
├── integration/           # Tests de integración (futuro)
├── TESTING.md             # Guía de testing completa
├── TESTING_READY.md       # Estado de readiness para testing
├── TESTING_STATUS.md      # Estado actual del testing
├── QUICK_TEST.md          # Guía rápida de testing
└── INSTALL_GO.md          # Guía de instalación de Go
```

## Tests Unitarios

Los tests unitarios actuales están en sus respectivos paquetes (convención Go):

- `internal/crypto/signer_test.go` - Tests de firmas criptográficas
- `internal/storage/db_test.go` - Tests de storage (LevelDB)
- `internal/consensus/abci_app_test.go` - Tests de consenso (ABCI)
- `internal/network/mesh_test.go` - Tests de mesh network

### Ejecutar Tests Unitarios

```powershell
# Desde oxy-blockchain/go/
.\test\scripts\test.bat
```

O manualmente:

```powershell
go test ./... -v
```

## Tests de Integración

Los tests de integración futuros irán en `test/integration/`.

Estos tests requieren:
- Nodo blockchain corriendo
- Configuración de red (opcional)
- API REST disponible

## Scripts de Testing

### `scripts/test.bat` (Windows)
Script completo que:
- Verifica instalación de Go
- Instala dependencias
- Ejecuta todos los tests
- Genera reporte de coverage

### `scripts/test.sh` (Linux/Mac)
Versión Unix del script de testing.

### `scripts/check-build.bat` (Windows)
Script para verificar compilación:
- Verifica sintaxis del código
- Compila el binario
- Reporta errores

## Guías de Testing

- **TESTING.md** - Guía completa de testing
- **TESTING_READY.md** - Qué necesitas para empezar a testear
- **TESTING_STATUS.md** - Estado actual del testing
- **QUICK_TEST.md** - Guía rápida de inicio
- **INSTALL_GO.md** - Cómo instalar Go en Windows

## Ejecutar Tests

### Opción 1: Script Automático (Recomendado)

```powershell
# Desde oxy-blockchain/go/
.\test\scripts\test.bat
```

### Opción 2: Manual

```powershell
# Instalar dependencias
go mod download
go mod tidy

# Ejecutar tests
go test ./... -v

# Solo tests de un paquete
go test ./internal/crypto -v
go test ./internal/storage -v
go test ./internal/consensus -v
go test ./internal/network -v
```

### Opción 3: Coverage

```powershell
# Generar coverage
go test ./... -coverprofile=test/coverage.out
go tool cover -html=test/coverage.out -o test/coverage.html

# Abrir en navegador
start test/coverage.html
```

## Organización Futura

En el futuro, podemos organizar tests adicionales:

- `test/unit/` - Tests unitarios adicionales si es necesario
- `test/integration/` - Tests de integración completos
- `test/benchmarks/` - Benchmarks de performance
- `test/mocks/` - Mocks para testing

Los tests que están junto al código (convención Go) se mantienen allí para mejor cohesión.



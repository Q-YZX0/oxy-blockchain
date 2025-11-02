# Guía Rápida de Testing - Oxy•gen Blockchain

## ⚠️ Requisito Previo

**Necesitas tener Go instalado** para ejecutar los tests.

### Instalación de Go (Windows)

1. Descarga Go desde: https://golang.org/dl/
2. Instala el instalador `.msi`
3. Verifica instalación: Abre PowerShell y ejecuta `go version`

## 🚀 Testing Rápido

### Opción 1: Script Automático (Windows)

```powershell
# Ejecutar desde: oxy-blockchain/go/
.\test.bat
```

Este script:
- ✅ Verifica que Go esté instalado
- ✅ Instala dependencias
- ✅ Ejecuta todos los tests unitarios
- ✅ Genera reporte de coverage

### Opción 2: Script Manual

```powershell
# 1. Instalar dependencias
go mod download
go mod tidy

# 2. Ejecutar tests unitarios
go test ./... -v

# 3. Verificar compilación
go build -o bin/oxy-blockchain.exe ./cmd/oxy-blockchain/main.go
```

### Opción 3: Verificar Solo Compilación

```powershell
# Ejecutar desde: oxy-blockchain/go/
.\check-build.bat
```

Este script:
- ✅ Verifica sintaxis del código
- ✅ Compila el binario
- ✅ Reporta si hay errores

## 🧪 Tests Disponibles

### Tests Unitarios

```powershell
# Tests de firmas criptográficas
go test ./internal/crypto -v

# Tests de storage (LevelDB)
go test ./internal/storage -v

# Tests de consenso (ABCI)
go test ./internal/consensus -v

# Tests de network
go test ./internal/network -v
```

### Tests de Integración

Los tests de integración requieren un nodo corriendo. Ver `TESTING.md` para más detalles.

## 📊 Coverage

```powershell
# Generar coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Abrir coverage.html en el navegador
start coverage.html
```

## ✅ Verificación Pre-Testing

Antes de ejecutar tests, verifica:

1. **Go instalado**: `go version` debe mostrar Go 1.21+
2. **Dependencias**: `go mod download` debe completarse sin errores
3. **Puerto 8080 libre**: Si vas a iniciar el nodo, verifica que el puerto esté disponible

## 🔧 Troubleshooting

### Error: "go: command not found"
- Go no está instalado o no está en PATH
- Solución: Instalar Go y reiniciar la terminal

### Error: "cannot find package"
- Dependencias no instaladas
- Solución: `go mod download && go mod tidy`

### Error: "port already in use"
- Puerto 8080 en uso
- Solución: Cambiar `BLOCKCHAIN_API_PORT` a otro puerto (ej: 8081)

### Error: "permission denied"
- Problemas de permisos en Windows
- Solución: Ejecutar PowerShell como Administrador

## 📝 Notas

- Los tests usan directorios temporales (`/tmp/test-*` o `./test_data`)
- Algunos tests pueden requerir CGO habilitado para LevelDB
- Los tests de integración requieren CometBFT inicializado (opcional)


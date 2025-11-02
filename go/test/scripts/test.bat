@echo off
REM Script de Testing - Oxy•gen Blockchain (Windows)
REM Ejecutar desde: oxy-blockchain/go/

echo 🧪 Iniciando tests de Oxy•gen Blockchain...
echo.

REM Verificar que Go está instalado
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Error: Go no está instalado o no está en PATH
    echo Por favor instala Go 1.21+ desde: https://golang.org/dl/
    exit /b 1
)

echo ✅ Go encontrado
go version
echo.

REM Instalar dependencias
echo 📦 Instalando dependencias...
go mod download
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Error instalando dependencias
    exit /b 1
)
echo ✅ Dependencias instaladas
echo.

REM Ejecutar tests unitarios
echo 🧪 Ejecutando tests unitarios...
echo.

REM Cambiar al directorio raíz de go/ si estamos en test/scripts/
if exist "..\..\go.mod" (
    cd ..\..
)

echo --- Tests de Crypto (Firmas) ---
go test ./internal/crypto -v
echo.

echo --- Tests de Storage ---
go test ./internal/storage -v
echo.

echo --- Tests de Consensus (ABCI) ---
go test ./internal/consensus -v
echo.

echo --- Tests de Network ---
go test ./internal/network -v
echo.

REM Ejecutar todos los tests juntos
echo --- Ejecutando todos los tests ---
go test ./... -v
if %ERRORLEVEL% NEQ 0 (
    echo ⚠️ Algunos tests fallaron
)

REM Generar coverage (ya estamos en raíz de go/)
echo.
echo 📊 Generando reporte de coverage...
go test ./... -coverprofile=test\coverage.out
go tool cover -html=test\coverage.out -o test\coverage.html
if %ERRORLEVEL% EQU 0 (
    echo ✅ Coverage reporte generado: test\coverage.html
) else (
    echo ⚠️ No se pudo generar coverage (requiere CGO habilitado para algunos tests)
)

echo.
echo 🎉 Testing completado!
pause


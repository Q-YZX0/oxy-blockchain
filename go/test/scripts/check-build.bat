@echo off
REM Script para verificar que el código compila correctamente (Windows)
REM Ejecutar desde: oxy-blockchain/go/ o oxy-blockchain/go/test/scripts/

REM Cambiar al directorio raíz de go/ si estamos en test/scripts/
if exist "..\..\go.mod" (
    cd ..\..
)

echo 🔍 Verificando compilación de Oxy•gen Blockchain...
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
echo.

REM Verificar sintaxis sin compilar
echo 🔍 Verificando sintaxis del código...
go vet ./...
if %ERRORLEVEL% NEQ 0 (
    echo ⚠️ Advertencias encontradas en el código
)
echo.

REM Intentar compilar
echo 🔨 Compilando binario...
go build -o bin/oxy-blockchain.exe ./cmd/oxy-blockchain/main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ Compilación exitosa: bin/oxy-blockchain.exe
    echo.
    echo El binario está listo para ejecutar.
    echo Ejecuta: bin\oxy-blockchain.exe
) else (
    echo ❌ Error de compilación
    echo Revisa los errores arriba
    exit /b 1
)

echo.
echo 🎉 Verificación completada!
pause


@echo off
REM Script principal para ejecutar tests
REM Ejecutar desde: oxy-blockchain/go/

echo 🧪 Oxy•gen Blockchain - Ejecutar Tests
echo.

REM Cambiar al directorio del script
cd /d "%~dp0"

REM Ejecutar script de testing
call scripts\test.bat



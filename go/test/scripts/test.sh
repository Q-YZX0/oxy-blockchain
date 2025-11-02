#!/bin/bash
# Script de Testing - Oxy•gen Blockchain
# Ejecutar desde: oxy-blockchain/go/

set -e

echo "🧪 Iniciando tests de Oxy•gen Blockchain..."
echo ""

# Colores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Verificar que Go está instalado
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Error: Go no está instalado o no está en PATH${NC}"
    echo "Por favor instala Go 1.21+ desde: https://golang.org/dl/"
    exit 1
fi

echo -e "${GREEN}✅ Go encontrado: $(go version)${NC}"
echo ""

# Verificar versión de Go (requiere 1.21+)
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo -e "${RED}❌ Error: Se requiere Go 1.21 o superior. Versión actual: $GO_VERSION${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Versión de Go compatible${NC}"
echo ""

# Instalar dependencias
echo "📦 Instalando dependencias..."
go mod download
go mod tidy
echo -e "${GREEN}✅ Dependencias instaladas${NC}"
echo ""

# Ejecutar tests unitarios
echo "🧪 Ejecutando tests unitarios..."
echo ""

echo "--- Tests de Crypto (Firmas) ---"
go test ./internal/crypto -v || echo -e "${YELLOW}⚠️ Algunos tests de crypto fallaron${NC}"
echo ""

echo "--- Tests de Storage ---"
go test ./internal/storage -v || echo -e "${YELLOW}⚠️ Algunos tests de storage fallaron${NC}"
echo ""

echo "--- Tests de Consensus (ABCI) ---"
go test ./internal/consensus -v || echo -e "${YELLOW}⚠️ Algunos tests de consensus fallaron${NC}"
echo ""

echo "--- Tests de Network ---"
go test ./internal/network -v || echo -e "${YELLOW}⚠️ Algunos tests de network fallaron${NC}"
echo ""

# Ejecutar todos los tests juntos
echo "--- Ejecutando todos los tests ---"
go test ./... -v
TEST_RESULT=$?

echo ""
if [ $TEST_RESULT -eq 0 ]; then
    echo -e "${GREEN}✅ Todos los tests pasaron${NC}"
else
    echo -e "${YELLOW}⚠️ Algunos tests fallaron (exit code: $TEST_RESULT)${NC}"
fi

# Generar coverage
echo ""
echo "📊 Generando reporte de coverage..."
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}✅ Coverage reporte generado: coverage.html${NC}"

echo ""
echo "🎉 Testing completado!"


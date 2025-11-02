# Estructura de Tests - Oxy•gen Blockchain

## ✅ Organización Completada

Los tests y documentación están ahora organizados en `test/`.

## Estructura Final

```
go/
├── test/                      # 📁 Carpeta principal de tests
│   ├── README.md              # Documentación principal
│   ├── run-tests.bat          # Script principal para ejecutar tests
│   │
│   ├── scripts/               # 📁 Scripts de testing
│   │   ├── test.bat           # Script completo de testing (Windows)
│   │   ├── test.sh            # Script completo de testing (Linux/Mac)
│   │   └── check-build.bat    # Script de verificación de compilación
│   │
│   ├── unit/                  # 📁 Tests unitarios adicionales (futuro)
│   │   └── (reservado para tests unitarios adicionales)
│   │
│   ├── integration/           # 📁 Tests de integración (futuro)
│   │   └── (reservado para tests de integración completos)
│   │
│   └── docs/                  # 📁 Documentación de testing
│       ├── TESTING.md         # Guía completa de testing
│       ├── TESTING_READY.md   # Estado de readiness
│       ├── TESTING_STATUS.md  # Estado actual
│       ├── QUICK_TEST.md      # Guía rápida
│       └── INSTALL_GO.md      # Instrucciones de instalación
│
├── internal/                  # Código fuente (tests unitarios aquí - convención Go)
│   ├── consensus/
│   │   └── abci_app_test.go   # Tests unitarios junto al código
│   ├── crypto/
│   │   └── signer_test.go     # Tests unitarios junto al código
│   ├── storage/
│   │   └── db_test.go         # Tests unitarios junto al código
│   └── network/
│       └── mesh_test.go       # Tests unitarios junto al código
│
├── Makefile                   # Actualizado para usar test/
└── README.md                   # Actualizado con referencia a test/
```

## ¿Por qué esta organización?

### Tests unitarios junto al código (`internal/*/*_test.go`)
- ✅ **Convención estándar de Go**: Los tests unitarios van junto al código
- ✅ **Mejor cohesión**: Fácil encontrar tests relacionados con el código
- ✅ **Importaciones simples**: Los tests pueden importar desde el mismo paquete

### Scripts y documentación en `test/`
- ✅ **Organización clara**: Todo lo relacionado con testing en un lugar
- ✅ **Separación de concerns**: Scripts y docs no mezclados con código fuente
- ✅ **Fácil de encontrar**: Un solo lugar para todo lo de testing

## Uso

### Ejecutar Tests

```powershell
# Opción 1: Desde test/ (más directo)
cd test
.\scripts\test.bat

# Opción 2: Desde raíz de go/
.\test\scripts\test.bat

# Opción 3: Con Make
make test
```

### Ver Documentación

```powershell
# Ver guía completa
type test\TESTING.md

# Ver guía rápida
type test\QUICK_TEST.md
```

## Convención Go vs. Organización Manual

### Tests Unitarios (Convención Go)
- **Ubicación**: Junto al código (`internal/*/*_test.go`)
- **Razón**: Convención estándar de Go, mejor cohesión
- **Ejecutar**: `go test ./internal/crypto -v`

### Tests de Integración (Carpeta test/)
- **Ubicación**: `test/integration/` (futuro)
- **Razón**: Tests más complejos que requieren setup especial
- **Ejecutar**: `go test ./test/integration -v`

### Scripts y Documentación (Carpeta test/)
- **Ubicación**: `test/scripts/` y `test/docs/`
- **Razón**: Organización clara, no mezclado con código

## Beneficios de esta Organización

✅ **Raíz limpia**: Menos archivos en la raíz de `go/`
✅ **Fácil de encontrar**: Todo lo de testing en `test/`
✅ **Convención Go respetada**: Tests unitarios junto al código
✅ **Escalable**: Fácil agregar tests de integración en el futuro
✅ **Documentación centralizada**: Todas las guías en un lugar



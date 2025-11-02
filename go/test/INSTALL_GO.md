# Instalación de Go para Testing

## Windows

### Paso 1: Descargar Go

1. Visita: https://golang.org/dl/
2. Descarga el instalador para Windows (`.msi`)
3. Ejecuta el instalador

### Paso 2: Verificar Instalación

Abre PowerShell (como Administrador) y ejecuta:

```powershell
go version
```

Deberías ver algo como: `go version go1.21.x windows/amd64`

### Paso 3: Configurar PATH (si es necesario)

Si `go version` no funciona, el PATH puede no estar configurado. Verifica:

1. Abre "Variables de entorno" en Windows
2. En "Variables del sistema", busca `Path`
3. Debe incluir: `C:\Program Files\Go\bin`
4. Si no está, agrégalo y reinicia la terminal

### Paso 4: Verificar que funciona

```powershell
# Desde oxy-blockchain/go/
cd oxy-blockchain\go

# Verificar Go
go version

# Instalar dependencias
go mod download

# Ejecutar tests básicos
go test ./internal/crypto -v
```

## Alternativa: Usar Chocolatey

Si tienes Chocolatey instalado:

```powershell
choco install golang
```

## Verificación Final

Una vez Go esté instalado, ejecuta:

```powershell
# Desde oxy-blockchain/go/
.\test.bat
```

Este script ejecutará todos los tests automáticamente.


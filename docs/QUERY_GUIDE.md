# Guía de Queries - Oxy•gen Blockchain

El ABCI App soporta las siguientes queries:

## Queries Disponibles

### 1. Obtener Altura Actual
```
Path: "height"
Response: String con la altura del último bloque
```

### 2. Obtener Balance de Cuenta
```
Path: "balance/{address}"
Ejemplo: "balance/0x1234567890123456789012345678901234567890"
Response: JSON con address y balance
```

### 3. Obtener Estado Completo de Cuenta
```
Path: "account/{address}"
Ejemplo: "account/0x1234567890123456789012345678901234567890"
Response: JSON con AccountState completo (address, balance, nonce, codeHash, storage)
```

### 4. Obtener Transacción por Hash
```
Path: "tx/{hash}"
Ejemplo: "tx/0xabcdef1234567890abcdef1234567890abcdef12"
Response: JSON con datos de la transacción
```

### 5. Obtener Bloque por Altura
```
Path: "block/{height}"
Ejemplo: "block/42"
Response: JSON con datos completos del bloque
```

## Uso desde CometBFT CLI

```bash
# Consultar altura
cometbft query --path="height" --chain-id=oxy-gen-chain

# Consultar balance
cometbft query --path="balance/0x1234567890123456789012345678901234567890" --chain-id=oxy-gen-chain

# Consultar cuenta
cometbft query --path="account/0x1234567890123456789012345678901234567890" --chain-id=oxy-gen-chain

# Consultar transacción
cometbft query --path="tx/0xabcdef1234567890abcdef1234567890abcdef12" --chain-id=oxy-gen-chain

# Consultar bloque
cometbft query --path="block/42" --chain-id=oxy-gen-chain
```

## Uso desde Go

```go
import "github.com/oxygen-economy/oxy-blockchain/internal/consensus"

// Desde ABCI App
response := app.Query(abcitypes.RequestQuery{
    Path: "balance/0x1234567890123456789012345678901234567890",
    Data: nil,
})

if response.Code == 0 {
    // Parsear respuesta JSON
    var result map[string]interface{}
    json.Unmarshal(response.Value, &result)
}
```


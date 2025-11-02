# Oxy•gen Blockchain Nativa

Blockchain nativa del sistema Oxy•gen que corre en nodos físicos. Completamente independiente de Internet, funciona mediante mesh network (WiFi/LoRa).

## Stack Tecnológico

- **Consenso**: CometBFT (anteriormente Tendermint Core)
- **Ejecución**: EVMone (EVM compatible)
- **Lenguaje Base**: Go
- **SDK/API Layer**: Node.js + TypeScript

## Arquitectura

```
oxy-blockchain/
├── go/                    # Core de la blockchain (Go)
│   ├── consensus/        # CometBFT integration
│   ├── execution/         # EVMone integration
│   ├── storage/          # LevelDB/RocksDB
│   └── network/          # P2P usando oxygen-sdk mesh
# Nota: El cliente TypeScript está ahora en `oxygen-sdk/src/blockchain/`
# Ver: `oxygen-sdk/src/blockchain/native-client.ts`
└── contracts/            # Contratos inteligentes
    ├── OXG.sol
    ├── GreenPool.sol
    └── DAO.sol
```

## Características

- ✅ Completamente independiente de Internet
- ✅ Funciona mediante mesh network (WiFi/LoRa)
- ✅ EVM compatible (contratos Solidity)
- ✅ Tolerante a particiones de red
- ✅ Consenso Proof of Stake (PoS)

## Desarrollo

### Requisitos

- Go 1.21+
- Node.js 18+
- CometBFT
- EVMone

### Build

```bash
# Build del core (Go)
cd go
go build

# Nota: El cliente TypeScript está en oxygen-sdk
# Para usarlo, instalar oxygen-sdk: npm install @oxygen/sdk
```

## Integración con oxygen-sdk

El nodo físico usa `oxygen-sdk` para:
- Mesh network (transmisión de transacciones/bloques)
- Discovery de validadores
- Comunicación P2P entre nodos


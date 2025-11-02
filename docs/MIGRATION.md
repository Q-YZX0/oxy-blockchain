# Migración: Cliente TypeScript movido al SDK

## Cambio Realizado

El cliente TypeScript para interactuar con la blockchain nativa ha sido movido de `oxy-blockchain/api/` al SDK principal `oxygen-sdk/src/blockchain/`.

## Razón del Cambio

- ✅ Evita redundancia: El SDK ya maneja blockchain (externa y mesh)
- ✅ Cohesión: Toda la funcionalidad blockchain está junta
- ✅ Dependencias: El cliente ya dependía del SDK, ahora está integrado
- ✅ Experiencia del desarrollador: Todo desde un solo lugar (`@oxygen/sdk`)

## Cambios en el Código

### Antes

```typescript
// ❌ Importar de dos lugares diferentes
import { Mesh } from '@oxygen/sdk';
import { BlockchainClient } from '@oxygen/blockchain-api'; // Separado

const blockchain = new BlockchainClient({ mesh, chainId: 'oxy-gen-chain' });
```

### Ahora

```typescript
// ✅ Todo desde un solo lugar
import { Mesh, NativeBlockchainClient } from '@oxygen/sdk';

const blockchain = new NativeBlockchainClient({ 
  mesh, 
  chainId: 'oxy-gen-chain',
  rpcEndpoint: 'http://localhost:8080' // Opcional: API REST local
});
```

## Nueva Estructura

### En el SDK

```
oxygen-sdk/src/blockchain/
├── index.ts              # Exportaciones principales
├── native-client.ts       # NativeBlockchainClient (antes BlockchainClient)
├── mesh-bridge.ts         # NativeBlockchainMeshBridge (antes MeshBridge)
└── types.ts              # Tipos TypeScript
```

### Exportaciones

El cliente ahora está disponible desde el SDK principal:

```typescript
import { 
  NativeBlockchainClient,
  NativeBlockchainMeshBridge,
  // Tipos
  Block,
  Transaction,
  TransactionReceipt,
  AccountState,
  ValidatorInfo,
  BlockchainStatus,
  BroadcastOptions,
} from '@oxygen/sdk';
```

## Estrategia Híbrida

El cliente ahora implementa una estrategia híbrida mejorada:

1. **API REST local** (prioridad): Si `rpcEndpoint` está configurado, usa el API REST del nodo Go local
2. **Mesh network** (fallback): Si el API REST no está disponible o falla, usa la mesh network

Esto permite:
- Desarrollo rápido usando API REST local
- Funcionamiento descentralizado usando mesh network cuando el nodo local no está disponible

## Archivos Eliminados

Los siguientes archivos/carpetas fueron eliminados:
- `oxy-blockchain/api/` (completo)
- `oxy-blockchain/api/README.md`
- `oxy-blockchain/api/src/client.ts`
- `oxy-blockchain/api/src/types.ts`
- `oxy-blockchain/api/src/mesh-bridge.ts`
- `oxy-blockchain/api/src/index.ts`
- `oxy-blockchain/api/package.json`
- `oxy-blockchain/api/tsconfig.json`

## Compatibilidad

⚠️ **Breaking Change**: Este cambio rompe la compatibilidad con código que importaba desde `@oxygen/blockchain-api`.

**Migración requerida:**
1. Actualizar imports: De `@oxygen/blockchain-api` a `@oxygen/sdk`
2. Cambiar nombre de clase: `BlockchainClient` → `NativeBlockchainClient`
3. Instalar solo SDK: `npm install @oxygen/sdk` (ya no necesita `@oxygen/blockchain-api`)


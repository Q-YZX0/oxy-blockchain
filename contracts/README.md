# Contratos Inteligentes Oxy•gen

Contratos inteligentes para la blockchain nativa de Oxy•gen.

## Contratos

### OXG.sol

Token ERC-20 nativo de Oxy•gen Economy.

**Características**:
- Token ERC-20 compatible
- Fee automático (5%) hacia GreenPool
- Quema automática (burn mechanism)
- Distribución de recompensas a validadores
- Exenciones de fee para direcciones específicas

**Funciones principales**:
- `transfer()`: Transferencia con fee automático
- `setGreenPoolWallet()`: Cambiar dirección del GreenPool (solo owner)
- `setFeeExemption()`: Exentar direcciones de fee
- `distributeValidatorRewards()`: Distribuir recompensas a validadores

### GreenPool.sol

Fondo ambiental que recibe contribuciones automáticas del token OXG.

**Características**:
- Recibe fees automáticos de transacciones OXG
- Gestión de proyectos ambientales
- Distribución de fondos a proyectos aprobados
- Integración con DAO para aprobación

**Funciones principales**:
- `createProject()`: Crear nuevo proyecto ambiental
- `approveProject()`: Aprobar proyecto (DAO)
- `distributeToProject()`: Distribuir fondos a proyecto aprobado
- `getBalance()`: Obtener balance del GreenPool

### DAO.sol

Organización Autónoma Descentralizada para gobernanza del ecosistema.

**Características**:
- Staking de OXG para participar en votaciones
- Creación y votación de propuestas
- Múltiples tipos de propuestas:
  - Aprobar proyectos en GreenPool
  - Cambiar parámetros del sistema
  - Gastos de tesorería
  - Actualizaciones de contratos

**Funciones principales**:
- `depositStake()`: Depositar OXG para votar
- `createProposal()`: Crear nueva propuesta
- `vote()`: Votar en propuesta
- `executeProposal()`: Ejecutar propuesta aprobada

## Despliegue

Los contratos se despliegan en la blockchain nativa usando el EVM compatible (EVMone).

## Dependencias

- OpenZeppelin Contracts (ERC20, Ownable)
- Solidity ^0.8.20

## Compilación

```bash
# Instalar dependencias
npm install @openzeppelin/contracts

# Compilar
npx hardhat compile
# o
solc --version
solc contracts/*.sol
```

## Notas

- Los contratos corren nativamente en los nodos físicos
- No dependen de blockchains externas
- Compatible con herramientas Web3 estándar (Metamask, ethers.js, etc.)


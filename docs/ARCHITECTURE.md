# Arquitectura de Oxy•gen Blockchain

## Visión General

Oxy•gen Blockchain es una blockchain nativa completamente descentralizada que corre en nodos físicos. Funciona sin Internet, usando mesh network (WiFi/LoRa) para comunicación entre nodos.

## Stack Tecnológico

- **Consenso**: CometBFT (BFT tolerante a particiones)
- **Ejecución**: EVMone (EVM compatible)
- **Core**: Go 1.21+
- **API Layer**: Node.js + TypeScript
- **Storage**: LevelDB
- **Red P2P**: Integración con `oxygen-sdk` mesh network

## Arquitectura en Capas

### 1. Capa de Consenso (CometBFT)

**Responsabilidades**:
- Validación de transacciones
- Producción de bloques
- Consenso distribuido entre validadores
- Tolerancia a particiones de red

**Características**:
- Proof of Stake (PoS)
- Validadores rotan para producir bloques
- Cuando meshes se desconectan, cada una mantiene su propia cadena
- Al reconectarse, sincronizan estado

### 2. Capa de Ejecución (EVMone)

**Responsabilidades**:
- Ejecutar transacciones
- Ejecutar contratos inteligentes (Solidity)
- Mantener estado de cuentas y contratos
- Compatibilidad con EVM

**Características**:
- Compatible con contratos Web3 estándar
- Soporta Solidity
- Compatible con herramientas Web3 (Metamask, ethers.js, etc.)

### 3. Capa de Storage (LevelDB)

**Responsabilidades**:
- Almacenar bloques
- Almacenar estado de la blockchain
- Almacenar transacciones
- Pruning de bloques antiguos

### 4. Capa de Red (oxygen-sdk Mesh)

**Responsabilidades**:
- Transmitir transacciones por la mesh
- Transmitir bloques por la mesh
- Discovery de otros validadores
- Comunicación P2P entre nodos

**Integración**:
- Usa `oxygen-sdk` para mesh network
- Transmite por WiFi/LoRa
- No requiere Internet

### 5. Capa de API (Node.js/TypeScript)

**Responsabilidades**:
- Interfaz para `oxygen-sdk` y aplicaciones
- Cliente TypeScript para interactuar con blockchain
- Tipos TypeScript
- Suscripciones a eventos

## Flujo de Transacciones

1. **Usuario/DApp** envía transacción
2. **API Layer** recibe transacción
3. **Mesh Bridge** transmite transacción por la mesh
4. **Otros nodos** reciben transacción
5. **Validadores** validan transacción
6. **Consenso** produce bloque con transacciones
7. **Ejecución** ejecuta transacciones en EVM
8. **Storage** guarda bloque y estado
9. **Mesh** transmite bloque a otros nodos

## Consenso y Validadores

### Sistema de Validadores

- Los validadores deben tener stake de OXG
- Rotación para producir bloques
- Validación distribuida
- Recompensas por participación

### Tolerancia a Particiones

Cuando la mesh se divide en subredes:
- Cada subred mantiene su propia cadena
- Los bloques se validan dentro de cada subred
- Cuando las subredes se reconectan:
  - Se sincronizan estado
  - Se resuelven conflictos (fork recovery)
  - Se unifica la cadena

## Integración con oxygen-sdk

El nodo físico usa `oxygen-sdk` para:

1. **Mesh Network**:
   - Transmitir transacciones
   - Transmitir bloques
   - Recibir eventos de otros nodos

2. **Discovery**:
   - Encontrar otros validadores
   - Conectarse a peers

3. **Routing**:
   - Elegir mejor ruta para transmisión
   - Edge AI para optimización

## Internet como Puente Opcional

Internet/satélite se usa solo cuando:
- Dos meshes están completamente desconectadas
- Necesitan sincronizar estado
- Actúa como puente temporal

No es necesario para el funcionamiento normal de la red.

## Seguridad

- Validadores con stake (PoS)
- Firma criptográfica de transacciones
- Validación distribuida
- Tolerancia a nodos maliciosos (hasta 1/3)
- Contratos auditables on-chain

## Escalabilidad

- Bloque size limitado
- Pruning de bloques antiguos
- Optimización para nodos con recursos limitados
- TPS objetivo definido

## Estado Actual

⚠️ **Work in Progress**

La implementación está en desarrollo. Componentes básicos creados:
- ✅ Estructura de proyecto
- ✅ Integración básica con CometBFT (estructura)
- ✅ Integración básica con EVMone (estructura)
- ✅ API Layer TypeScript
- ✅ Contratos inteligentes (OXG, GreenPool, DAO)
- ⏳ Integración completa pendiente


package execution

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/Q-YZX0/oxy-blockchain/internal/storage"
)

// EVMExecutor ejecuta transacciones usando go-ethereum (EVM compatible)
// Nota: go-ethereum es compatible con EVM y puede usarse como alternativa a EVMone
type EVMExecutor struct {
	storage          *storage.BlockchainDB
	stateManager     *StateManager
	stateDB          *state.StateDB
	chainConfig      *params.ChainConfig
	currentHeight    uint64
	currentTimestamp int64
	running          bool
}

// NewEVMExecutor crea una nueva instancia del ejecutor EVM
func NewEVMExecutor(storage *storage.BlockchainDB) *EVMExecutor {
	// Configurar chain config para Oxy•gen
	chainConfig := &params.ChainConfig{
		ChainID:             big.NewInt(999), // Chain ID de Oxy•gen (temporal)
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:        big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
	}
	
	stateManager := NewStateManager(storage, "./data")
	
	return &EVMExecutor{
		storage:      storage,
		stateManager: stateManager,
		chainConfig:  chainConfig,
		running:      false,
	}
}

// Start inicia el ejecutor EVM
func (e *EVMExecutor) Start() error {
	// Cargar estado desde storage
	stateDB, err := e.stateManager.LoadState()
	if err != nil {
		return fmt.Errorf("error cargando estado: %w", err)
	}
	
	e.stateDB = stateDB
	e.running = true
	log.Println("Ejecutor EVM iniciado")
	return nil
}

// Stop detiene el ejecutor EVM
func (e *EVMExecutor) Stop() error {
	if !e.running {
		return nil
	}
	
	// Guardar estado antes de detener
	if err := e.stateManager.SaveState(); err != nil {
		log.Printf("Advertencia: error guardando estado: %v", err)
	}
	
	// Cerrar gestor de estado
	if err := e.stateManager.Close(); err != nil {
		log.Printf("Advertencia: error cerrando gestor de estado: %v", err)
	}
	
	e.running = false
	log.Println("Ejecutor EVM detenido")
	return nil
}

// SetCurrentBlockInfo establece la información del bloque actual
func (e *EVMExecutor) SetCurrentBlockInfo(height uint64, timestamp int64) {
	e.currentHeight = height
	e.currentTimestamp = timestamp
}

// ExecuteTransaction ejecuta una transacción y actualiza el estado
func (e *EVMExecutor) ExecuteTransaction(tx *Transaction) (*ExecutionResult, error) {
	if !e.running {
		return nil, fmt.Errorf("ejecutor EVM no está corriendo")
	}

	// Convertir transacción a formato go-ethereum
	from := common.HexToAddress(tx.From)
	to := common.HexToAddress(tx.To)
	value, ok := new(big.Int).SetString(tx.Value, 10)
	if !ok {
		return nil, fmt.Errorf("valor inválido: %s", tx.Value)
	}
	gasPrice, ok := new(big.Int).SetString(tx.GasPrice, 10)
	if !ok {
		return nil, fmt.Errorf("gas price inválido: %s", tx.GasPrice)
	}
	
	// Obtener nonce actual si no se proporcionó
	nonce := tx.Nonce
	if nonce == 0 {
		stateDB := e.getStateDB()
		if stateDB != nil {
			nonce = stateDB.GetNonce(from)
		}
	}
	
	// Crear transacción Ethereum
	ethereumTx := types.NewTransaction(
		nonce,
		to,
		value,
		tx.GasLimit,
		gasPrice,
		tx.Data,
	)
	
	// Preparar header del bloque con valores reales
	header := &types.Header{
		Number:   big.NewInt(int64(e.currentHeight)),
		GasLimit: tx.GasLimit,
		Time:     uint64(e.currentTimestamp),
	}
	
	// Preparar contexto de ejecución
	blockContext := core.NewEVMBlockContext(header, nil, nil)
	
	// Crear EVM
	txContext := core.NewEVMTxContext(&types.Transaction{})
	evm := vm.NewEVM(blockContext, txContext, e.getStateDB(), e.chainConfig, vm.Config{})
	
	// Crear message para ejecutar
	msg := types.NewMessage(
		from,
		&to,
		tx.Nonce,
		value,
		tx.GasLimit,
		gasPrice,
		big.NewInt(0),
		tx.Data,
		nil,
		false,
	)
	
	// Ejecutar transacción
	result, err := core.ApplyMessage(evm, msg, new(core.GasPool).AddGas(tx.GasLimit))
	
	if err != nil {
		return &ExecutionResult{
			Success: false,
			GasUsed: result.UsedGas,
			Error:   err.Error(),
		}, nil
	}
	
	// Convertir logs
	logs := make([]Log, len(result.Logs))
	for i, log := range result.Logs {
		topics := make([]string, len(log.Topics))
		for j, topic := range log.Topics {
			topics[j] = topic.Hex()
		}
		logs[i] = Log{
			Address: log.Address.Hex(),
			Topics:  topics,
			Data:    log.Data,
		}
	}
	
	// Si la ejecución fue exitosa, guardar estado intermedio
	if err == nil && !result.Failed() {
		// Finalizar el StateDB para aplicar cambios
		e.stateDB.Finalise(true)
	}
	
	return &ExecutionResult{
		Success:    err == nil && result.Failed() == false,
		GasUsed:    result.UsedGas,
		ReturnData: result.ReturnData,
		Logs:       logs,
		Error:      "",
	}, nil
}

// getStateDB obtiene o crea el StateDB
func (e *EVMExecutor) getStateDB() *state.StateDB {
	if e.stateDB == nil {
		// Si no hay StateDB, cargar desde manager
		if e.stateManager != nil {
			e.stateDB, _ = e.stateManager.LoadState()
		}
	}
	return e.stateDB
}

// GetState retorna el estado actual de una cuenta
func (e *EVMExecutor) GetState(address string) (*AccountState, error) {
	if !e.running {
		return nil, fmt.Errorf("ejecutor EVM no está corriendo")
	}
	
	addr := common.HexToAddress(address)
	stateDB := e.getStateDB()
	
	balance := stateDB.GetBalance(addr)
	nonce := stateDB.GetNonce(addr)
	codeHash := stateDB.GetCodeHash(addr)
	
	// Obtener storage (primeras 100 slots como ejemplo)
	storage := make(map[string]string)
	for i := 0; i < 100; i++ {
		key := common.BigToHash(big.NewInt(int64(i)))
		value := stateDB.GetState(addr, key)
		if value != (common.Hash{}) {
			storage[key.Hex()] = value.Hex()
		}
	}
	
	return &AccountState{
		Address:  address,
		Balance:  balance.String(),
		Nonce:    nonce,
		CodeHash: codeHash.Hex(),
		Storage:  storage,
	}, nil
}

// DeployContract despliega un contrato inteligente
func (e *EVMExecutor) DeployContract(
	from string,
	code []byte,
	constructorArgs []byte,
	gasLimit uint64,
	gasPrice string,
) (string, *ExecutionResult, error) {
	if !e.running {
		return "", nil, fmt.Errorf("ejecutor EVM no está corriendo")
	}
	
	fromAddr := common.HexToAddress(from)
	
	// Combinar bytecode con constructor args
	contractData := append(code, constructorArgs...)
	
	// Obtener nonce actual
	stateDB := e.getStateDB()
	nonce := uint64(0)
	if stateDB != nil {
		nonce = stateDB.GetNonce(common.HexToAddress(from))
	}
	
	// Crear transacción de deployment (To es nil)
	tx := &Transaction{
		From:     from,
		To:       "", // Empty para deployment
		Value:    "0",
		Data:     contractData,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Nonce:    nonce,
	}
	
	// Ejecutar transacción
	result, err := e.ExecuteTransaction(tx)
	if err != nil {
		return "", nil, fmt.Errorf("error ejecutando deployment: %w", err)
	}
	
	if !result.Success {
		return "", result, fmt.Errorf("deployment falló: %s", result.Error)
	}
	
	// Calcular dirección del contrato
	stateDB := e.getStateDB()
	nonce := stateDB.GetNonce(fromAddr)
	contractAddr := crypto.CreateAddress(fromAddr, nonce).Hex()
	
	return contractAddr, result, nil
}

// CallContract ejecuta una llamada a un contrato (sin modificar estado)
func (e *EVMExecutor) CallContract(
	from string,
	contractAddr string,
	data []byte,
	gasLimit uint64,
) ([]byte, error) {
	if !e.running {
		return nil, fmt.Errorf("ejecutor EVM no está corriendo")
	}
	
	// Obtener nonce actual
	stateDB := e.getStateDB()
	nonce := uint64(0)
	if stateDB != nil {
		nonce = stateDB.GetNonce(common.HexToAddress(from))
	}
	
	// Crear transacción de llamada
	tx := &Transaction{
		From:     from,
		To:       contractAddr,
		Value:    "0",
		Data:     data,
		GasLimit: gasLimit,
		GasPrice: "0", // Sin costo para calls
		Nonce:    nonce,
	}
	
	// Ejecutar transacción
	result, err := e.ExecuteTransaction(tx)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando call: %w", err)
	}
	
	if !result.Success {
		return nil, fmt.Errorf("call falló: %s", result.Error)
	}
	
	return result.ReturnData, nil
}

// GetStateManager retorna el StateManager (para uso interno de consensus)
func (e *EVMExecutor) GetStateManager() *StateManager {
	return e.stateManager
}

// SaveState guarda el estado actual del EVM
func (e *EVMExecutor) SaveState() error {
	if e.stateManager == nil {
		return fmt.Errorf("stateManager no está inicializado")
	}
	return e.stateManager.SaveState()
}

// SaveStateAtHeight guarda el estado en una altura específica
func (e *EVMExecutor) SaveStateAtHeight(height uint64) error {
	if e.stateManager == nil {
		return fmt.Errorf("stateManager no está inicializado")
	}
	return e.stateManager.SaveStateAtHeight(height)
}

// Transaction representa una transacción a ejecutar
type Transaction struct {
	Hash     string
	From     string
	To       string
	Value    string
	Data     []byte
	GasLimit uint64
	GasPrice string
	Nonce    uint64
}

// ExecutionResult contiene el resultado de ejecutar una transacción
type ExecutionResult struct {
	Success    bool
	GasUsed    uint64
	ReturnData []byte
	Logs       []Log
	Error      string
}

// AccountState representa el estado de una cuenta
type AccountState struct {
	Address    string
	Balance    string
	Nonce      uint64
	CodeHash   string
	Storage    map[string]string
}

// Log representa un evento emitido por un contrato
type Log struct {
	Address string
	Topics  []string
	Data    []byte
}


package consensus

import (
	"bytes"
	"encoding/json"
	"fmt"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	execution "github.com/Q-YZX0/oxy-blockchain/internal/execution"
	"github.com/Q-YZX0/oxy-blockchain/internal/storage"
	"github.com/Q-YZX0/oxy-blockchain/internal/crypto/signer"
	"github.com/Q-YZX0/oxy-blockchain/internal/logger"
	"time"
	"math/big"
)

// ABCIApp implementa la interfaz ABCI de CometBFT
// Esta es la aplicación que corre sobre CometBFT
type ABCIApp struct {
	storage            *storage.BlockchainDB
	executor           *execution.EVMExecutor
	validators         *ValidatorSet
	state              *AppState
	currentBlockHeight uint64
	currentBlockTime   int64
	currentBlockTxs    []*Transaction
	currentBlockReceipts []*TransactionReceipt
	chainID            string
}

// AppState mantiene el estado de la aplicación
type AppState struct {
	Height      int64
	AppHash     []byte
	Validators  []abcitypes.ValidatorUpdate
}

// NewABCIApp crea una nueva aplicación ABCI
func NewABCIApp(storage *storage.BlockchainDB, executor *execution.EVMExecutor, validators *ValidatorSet, chainID string) *ABCIApp {
	return &ABCIApp{
		storage:    storage,
		executor:   executor,
		validators: validators,
		chainID:    chainID,
		state: &AppState{
			Height:     0,
			AppHash:    make([]byte, 32),
			Validators: []abcitypes.ValidatorUpdate{},
		},
		currentBlockTxs:     make([]*Transaction, 0),
		currentBlockReceipts: make([]*TransactionReceipt, 0),
	}
}

// Info retorna información sobre el estado de la aplicación
func (app *ABCIApp) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{
		Data:             fmt.Sprintf("oxy-blockchain-v0.1.0"),
		Version:          "0.1.0",
		AppVersion:       1,
		LastBlockHeight:  app.state.Height,
		LastBlockAppHash: app.state.AppHash,
	}
}

// InitChain inicializa la blockchain
func (app *ABCIApp) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	logger.Info("Inicializando blockchain")
	
	// Cargar validadores guardados
	if app.validators != nil {
		if err := app.validators.LoadValidators(); err != nil {
			logger.Warn().Err(err).Msg("Error cargando validadores")
		}
		
		// Usar validadores del set en lugar de los del genesis
		// Si no hay validadores guardados, usar los del genesis
		if len(app.validators.GetActiveValidators()) == 0 {
			// Convertir validadores del genesis al formato interno
			genesisValidators := make([]GenesisValidator, 0, len(req.Validators))
			for _, v := range req.Validators {
				// Extraer dirección desde la clave pública (simplificado)
				address := common.BytesToAddress(v.PubKey.Data).Hex()
				
				genesisValidators = append(genesisValidators, GenesisValidator{
					Address: address,
					PubKey:  v.PubKey.Data,
					Stake:   big.NewInt(int64(v.Power * 1e18)), // Convertir power a stake (aproximado)
				})
			}
			
			if err := app.validators.InitializeGenesisValidators(genesisValidators); err != nil {
				logger.Error().Err(err).Msg("Error inicializando validadores genesis")
			}
		}
		
		// Obtener validadores actualizados
		app.state.Validators = app.validators.ToCometBFTValidators()
	} else {
		// Fallback: usar validadores del genesis directamente
		app.state.Validators = req.Validators
	}
	
	return abcitypes.ResponseInitChain{
		Validators: app.state.Validators,
		AppHash:    app.state.AppHash,
	}
}

// BeginBlock se llama al inicio de cada bloque
func (app *ABCIApp) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	logger.Info().Uint64("height", uint64(req.Header.Height)).Msg("Iniciando bloque")
	
	// Guardar altura y timestamp actuales para uso en ejecución EVM
	app.state.Height = req.Header.Height
	app.currentBlockHeight = uint64(req.Header.Height)
	app.currentBlockTime = req.Header.Time.Unix()
	
	// Limpiar transacciones del bloque anterior
	app.currentBlockTxs = make([]*Transaction, 0)
	app.currentBlockReceipts = make([]*TransactionReceipt, 0)
	
	return abcitypes.ResponseBeginBlock{}
}

// DeliverTx procesa una transacción
func (app *ABCIApp) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	// Decodificar transacción
	var tx Transaction
	if err := json.Unmarshal(req.Tx, &tx); err != nil {
		return abcitypes.ResponseDeliverTx{
			Code: 1,
			Log:  fmt.Sprintf("Error decodificando transacción: %v", err),
		}
	}

	// Validar transacción básica
	if err := app.validateTransaction(&tx); err != nil {
		return abcitypes.ResponseDeliverTx{
			Code: 2,
			Log:  fmt.Sprintf("Transacción inválida: %v", err),
		}
	}

	// Convertir a formato execution.Transaction
	executionTx := &execution.Transaction{
		Hash:     tx.Hash,
		From:     tx.From,
		To:       tx.To,
		Value:    tx.Value,
		Data:     tx.Data,
		GasLimit: tx.GasLimit,
		GasPrice: tx.GasPrice,
		Nonce:    tx.Nonce,
	}

	// Establecer información del bloque actual en el ejecutor
	app.executor.SetCurrentBlockInfo(uint64(app.state.Height), app.currentBlockTime)

	// Ejecutar transacción con EVM
	result, err := app.executor.ExecuteTransaction(executionTx)
	if err != nil {
		return abcitypes.ResponseDeliverTx{
			Code: 3,
			Log:  fmt.Sprintf("Error ejecutando transacción: %v", err),
		}
	}

	if !result.Success {
		return abcitypes.ResponseDeliverTx{
			Code: 4,
			Log:  result.Error,
		}
	}

	// Guardar transacción
	txData, _ := json.Marshal(tx)
	app.storage.SaveTransaction(tx.Hash, txData)
	
	// Agregar transacción al bloque actual
	app.currentBlockTxs = append(app.currentBlockTxs, &tx)
	
	// Crear receipt de la transacción
	receipt := &TransactionReceipt{
		TransactionHash: tx.Hash,
		BlockNumber:     app.currentBlockHeight,
		GasUsed:         result.GasUsed,
		Status:          "success",
		Logs:            convertLogs(result.Logs),
		Error:           result.Error,
	}
	
	if !result.Success {
		receipt.Status = "failed"
	}
	
	app.currentBlockReceipts = append(app.currentBlockReceipts, receipt)
	
	// Guardar estado después de cada transacción exitosa
	if result.Success {
		// El estado se guardará al final del bloque en Commit
	}

	return abcitypes.ResponseDeliverTx{
		Code: 0,
		Log:  "OK",
		GasUsed: result.GasUsed,
		Events: app.buildEvents(result),
	}
}

// EndBlock se llama al final de cada bloque
func (app *ABCIApp) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	logger.Info().Int64("height", req.Height).Msg("Finalizando bloque")
	
	app.state.Height = req.Height
	
	// Rotar validadores periódicamente (cada 100 bloques)
	var validatorUpdates []abcitypes.ValidatorUpdate
	if req.Height > 0 && req.Height%100 == 0 {
		if app.validators != nil {
			updates, err := app.validators.RotateValidators()
			if err != nil {
				logger.Error().Err(err).Msg("Error rotando validadores")
			} else {
				validatorUpdates = updates
				logger.Info().Int("count", len(updates)).Msg("Validadores rotados")
			}
		}
	}
	
	return abcitypes.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
	}
}

// Commit confirma el bloque y retorna el AppHash
func (app *ABCIApp) Commit() abcitypes.ResponseCommit {
	// Guardar estado EVM completo (esto persiste el StateDB)
	if err := app.executor.SaveState(); err != nil {
		logger.Warn().Err(err).Msg("Error guardando estado EVM")
	}
	
	// Obtener root hash del StateDB
	stateRoot := app.executor.GetStateManager().GetRootHash()
	
	// Si no hay root, usar hash del estado de la aplicación
	var appHash []byte
	if stateRoot != (common.Hash{}) {
		appHash = stateRoot[:]
	} else {
		stateData, _ := json.Marshal(app.state)
		hash := crypto.Keccak256(stateData)
		appHash = hash[:32]
	}
	
	// Guardar metadata del estado
	stateData, _ := json.Marshal(map[string]interface{}{
		"root":      stateRoot.Hex(),
		"height":    app.state.Height,
		"app_hash": common.BytesToHash(appHash).Hex(),
	})
	app.storage.SaveState(stateData)
	
	// Guardar bloque completo
	if app.currentBlockHeight > 0 {
		if err := app.saveBlock(appHash); err != nil {
			logger.Warn().Err(err).Msg("Error guardando bloque")
		}
	}
	
	// Actualizar AppHash con root del StateDB
	copy(app.state.AppHash, appHash)
	
	return abcitypes.ResponseCommit{
		Data: appHash,
		RetainHeight: 0,
	}
}

// saveBlock guarda el bloque completo en storage
func (app *ABCIApp) saveBlock(blockHash []byte) error {
	// Calcular hash del bloque
	blockHashStr := common.BytesToHash(blockHash).Hex()
	
	// Obtener hash del bloque padre
	parentHash := ""
	if app.currentBlockHeight > 0 {
		parentBlockData, err := app.storage.GetBlock(app.currentBlockHeight - 1)
		if err == nil && parentBlockData != nil {
			var parentBlock Block
			if err := json.Unmarshal(parentBlockData, &parentBlock); err == nil {
				parentHash = parentBlock.Header.Hash
			}
		}
	}
	
	// Crear bloque completo
	block := &Block{
		Header: BlockHeader{
			Height:     app.currentBlockHeight,
			Hash:       blockHashStr,
			ParentHash: parentHash,
			Timestamp:  time.Unix(app.currentBlockTime, 0),
			ChainID:    app.chainID,
		},
		Transactions: app.currentBlockTxs,
		Receipts:     app.currentBlockReceipts,
	}
	
	// Guardar bloque
	blockData, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("error serializando bloque: %w", err)
	}
	
	if err := app.storage.SaveBlock(app.currentBlockHeight, blockData); err != nil {
		return fmt.Errorf("error guardando bloque: %w", err)
	}
	
	// Guardar altura del último bloque
	if err := app.storage.SaveLatestHeight(app.currentBlockHeight); err != nil {
		logger.Warn().Err(err).Msg("Error guardando altura")
	}
	
	logger.Info().
		Uint64("height", app.currentBlockHeight).
		Str("hash", blockHashStr[:8]).
		Int("transactions", len(app.currentBlockTxs)).
		Msg("Bloque guardado")
	
	return nil
}

// convertLogs convierte logs de execution a consensus
func convertLogs(execLogs []execution.Log) []Log {
	logs := make([]Log, len(execLogs))
	for i, log := range execLogs {
		logs[i] = Log{
			Address:     log.Address,
			Topics:      log.Topics,
			Data:        log.Data,
			BlockNumber: 0, // Se actualizará al guardar el bloque
			TxHash:      "",
		}
	}
	return logs
}

// Query permite consultar el estado de la aplicación
func (app *ABCIApp) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	// Parsear path del query
	// Formatos esperados:
	// - "balance/{address}" - Obtener balance de cuenta
	// - "account/{address}" - Obtener estado completo de cuenta
	// - "tx/{hash}" - Obtener transacción por hash
	// - "block/{height}" - Obtener bloque por altura
	// - "height" - Obtener altura actual
	
	path := string(req.Path)
	
	switch {
	case path == "height":
		height := uint64(app.state.Height)
		return abcitypes.ResponseQuery{
			Code:  0,
			Value: []byte(fmt.Sprintf("%d", height)),
		}
	
	case len(path) > 8 && path[:8] == "balance/":
		address := path[8:]
		accountState, err := app.executor.GetState(address)
		if err != nil {
			return abcitypes.ResponseQuery{
				Code:  1,
				Log:   fmt.Sprintf("Error obteniendo balance: %v", err),
			}
		}
		
		result := map[string]interface{}{
			"address": address,
			"balance": accountState.Balance,
		}
		
		resultData, _ := json.Marshal(result)
		return abcitypes.ResponseQuery{
			Code:  0,
			Value: resultData,
		}
	
	case len(path) > 7 && path[:7] == "account/":
		address := path[7:]
		accountState, err := app.executor.GetState(address)
		if err != nil {
			return abcitypes.ResponseQuery{
				Code:  1,
				Log:   fmt.Sprintf("Error obteniendo cuenta: %v", err),
			}
		}
		
		resultData, _ := json.Marshal(accountState)
		return abcitypes.ResponseQuery{
			Code:  0,
			Value: resultData,
		}
	
	case len(path) > 3 && path[:3] == "tx/":
		txHash := path[3:]
		txData, err := app.storage.GetTransaction(txHash)
		if err != nil {
			return abcitypes.ResponseQuery{
				Code:  1,
				Log:   fmt.Sprintf("Transacción no encontrada: %s", txHash),
			}
		}
		
		return abcitypes.ResponseQuery{
			Code:  0,
			Value: txData,
		}
	
	case len(path) > 6 && path[:6] == "block/":
		height := uint64(0)
		fmt.Sscanf(path[6:], "%d", &height)
		
		blockData, err := app.storage.GetBlock(height)
		if err != nil {
			return abcitypes.ResponseQuery{
				Code:  1,
				Log:   fmt.Sprintf("Bloque no encontrado: altura %d", height),
			}
		}
		
		return abcitypes.ResponseQuery{
			Code:  0,
			Value: blockData,
		}
	
	default:
		return abcitypes.ResponseQuery{
			Code:  1,
			Log:   fmt.Sprintf("Query path desconocido: %s", path),
		}
	}
}

// CheckTx valida una transacción sin ejecutarla
func (app *ABCIApp) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	var tx Transaction
	if err := json.Unmarshal(req.Tx, &tx); err != nil {
		return abcitypes.ResponseCheckTx{
			Code: 1,
			Log:  fmt.Sprintf("Error decodificando transacción: %v", err),
		}
	}

	// Validación completa de transacción
	if err := app.validateTransactionComplete(&tx); err != nil {
		return abcitypes.ResponseCheckTx{
			Code: 2,
			Log:  fmt.Sprintf("Transacción inválida: %v", err),
		}
	}

	return abcitypes.ResponseCheckTx{
		Code: 0,
		Log:  "OK",
	}
}

// validateTransaction valida una transacción básica
func (app *ABCIApp) validateTransaction(tx *Transaction) error {
	// Validaciones básicas
	if tx.From == "" {
		return fmt.Errorf("dirección remitente vacía")
	}
	
	if !common.IsHexAddress(tx.From) {
		return fmt.Errorf("dirección remitente inválida: %s", tx.From)
	}
	
	if tx.To != "" && !common.IsHexAddress(tx.To) {
		return fmt.Errorf("dirección destino inválida: %s", tx.To)
	}
	
	return nil
}

// validateTransactionComplete valida una transacción completamente (firma, nonce, balance)
func (app *ABCIApp) validateTransactionComplete(tx *Transaction) error {
	// Validaciones básicas primero
	if err := app.validateTransaction(tx); err != nil {
		return err
	}

	// Validar que tenga hash
	if tx.Hash == "" {
		return fmt.Errorf("transacción sin hash")
	}

	// Validar nonce (obtener nonce actual de la cuenta)
	accountState, err := app.executor.GetState(tx.From)
	if err == nil && accountState != nil {
		if tx.Nonce < accountState.Nonce {
			return fmt.Errorf("nonce inválido: esperado >= %d, tiene %d", accountState.Nonce, tx.Nonce)
		}
	}

	// Validar balance suficiente (si hay transferencia de valor)
	if tx.Value != "" && tx.Value != "0" {
		if accountState == nil {
			return fmt.Errorf("cuenta no encontrada: %s", tx.From)
		}

		// Parsear valor
		value, ok := new(big.Int).SetString(tx.Value, 10)
		if !ok {
			return fmt.Errorf("valor inválido: %s", tx.Value)
		}

		// Parsear balance
		balance, ok := new(big.Int).SetString(accountState.Balance, 10)
		if !ok {
			return fmt.Errorf("balance inválido: %s", accountState.Balance)
		}

		// Calcular gas cost
		gasPrice, ok := new(big.Int).SetString(tx.GasPrice, 10)
		if !ok {
			return fmt.Errorf("gas price inválido: %s", tx.GasPrice)
		}

		gasCost := new(big.Int).Mul(gasPrice, big.NewInt(int64(tx.GasLimit)))
		totalCost := new(big.Int).Add(value, gasCost)

		// Validar balance suficiente
		if balance.Cmp(totalCost) < 0 {
			return fmt.Errorf("balance insuficiente: tiene %s, necesita %s", balance.String(), totalCost.String())
		}
	}

	// Validar firma criptográfica
	if len(tx.Signature) == 0 {
		return fmt.Errorf("transacción sin firma")
	}

	// Convertir transacción a mapa para validación de firma
	txMap := map[string]interface{}{
		"hash":      tx.Hash,
		"from":      tx.From,
		"to":        tx.To,
		"value":     tx.Value,
		"data":      tx.Data,
		"gasLimit":  tx.GasLimit,
		"gasPrice":  tx.GasPrice,
		"nonce":     tx.Nonce,
		"signature": tx.Signature,
	}

	// Verificar firma
	_, err := signer.VerifyTransactionSignature(txMap)
	if err != nil {
		return fmt.Errorf("firma criptográfica inválida: %w", err)
	}

	// Verificar que el hash de la transacción sea correcto
	// Calcular hash esperado
	expectedHash, err := signer.CalculateTransactionHash(txMap)
	if err != nil {
		return fmt.Errorf("error calculando hash de transacción: %w", err)
	}

	// Comparar hash
	if tx.Hash != expectedHash.Hex() {
		return fmt.Errorf("hash de transacción inválido: esperado %s, tiene %s", expectedHash.Hex(), tx.Hash)
	}
	
	return nil
}

// buildEvents construye eventos a partir del resultado de ejecución
func (app *ABCIApp) buildEvents(result *execution.ExecutionResult) []abcitypes.Event {
	events := []abcitypes.Event{}
	
	// Evento de ejecución
	events = append(events, abcitypes.Event{
		Type: "execution",
		Attributes: []abcitypes.EventAttribute{
			{Key: "success", Value: fmt.Sprintf("%t", result.Success)},
			{Key: "gas_used", Value: fmt.Sprintf("%d", result.GasUsed)},
		},
	})
	
	// Eventos de logs de contratos
	for _, log := range result.Logs {
		events = append(events, abcitypes.Event{
			Type: "contract_log",
			Attributes: []abcitypes.EventAttribute{
				{Key: "address", Value: log.Address},
			},
		})
	}
	
	return events
}

// ListSnapshots retorna snapshots disponibles
func (app *ABCIApp) ListSnapshots(req abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

// OfferSnapshot ofrece un snapshot
func (app *ABCIApp) OfferSnapshot(req abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{Result: abcitypes.ResponseOfferSnapshot_REJECT}
}

// LoadSnapshotChunk carga un chunk de snapshot
func (app *ABCIApp) LoadSnapshotChunk(req abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

// ApplySnapshotChunk aplica un chunk de snapshot
func (app *ABCIApp) ApplySnapshotChunk(req abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{Result: abcitypes.ResponseApplySnapshotChunk_REJECT}
}


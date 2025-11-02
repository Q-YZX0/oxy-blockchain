package execution

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/Q-YZX0/oxy-blockchain/internal/storage"
)

// StateManager maneja el estado de la blockchain EVM
type StateManager struct {
	storage    *storage.BlockchainDB
	stateDB    *state.StateDB
	database   state.Database
	stateRoot  common.Hash
	dataDir    string
}

// NewStateManager crea un nuevo gestor de estado
func NewStateManager(storage *storage.BlockchainDB, dataDir string) *StateManager {
	return &StateManager{
		storage: storage,
		dataDir: dataDir,
	}
}

// LoadState carga el estado desde storage
func (sm *StateManager) LoadState() (*state.StateDB, error) {
	// Crear base de datos para StateDB usando LevelDB
	stateDBPath := filepath.Join(sm.dataDir, "evm_state")
	
	// Crear base de datos Ethereum compatible con LevelDB
	db, err := rawdb.NewLevelDBDatabase(stateDBPath, 0, 0, "", false)
	if err != nil {
		return nil, fmt.Errorf("error creando base de datos EVM: %w", err)
	}
	
	// Crear database wrapper para StateDB
	database := state.NewDatabase(db)
	
	// Intentar cargar root hash guardado
	stateData, err := sm.storage.GetState()
	var root common.Hash
	
	if err == nil && stateData != nil {
		// Parsear estado guardado
		var stateInfo map[string]interface{}
		if err := json.Unmarshal(stateData, &stateInfo); err == nil {
			if rootStr, ok := stateInfo["root"].(string); ok {
				root = common.HexToHash(rootStr)
			}
		}
	}
	
	// Si no hay root guardado, usar hash vacío (estado nuevo)
	if root == (common.Hash{}) {
		root = common.Hash{}
	}
	
	// Crear StateDB desde root (o estado nuevo)
	stateDB, err := state.New(root, database, nil)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error creando StateDB: %w", err)
	}
	
	sm.stateDB = stateDB
	sm.database = database
	sm.stateRoot = root
	
	return stateDB, nil
}

// SaveState guarda el estado completo en storage
func (sm *StateManager) SaveState() error {
	if sm.stateDB == nil {
		return fmt.Errorf("StateDB no está inicializado")
	}
	
	// Calcular root hash intermedio (commits todos los cambios)
	root := sm.stateDB.IntermediateRoot(true)
	
	// Commit el StateDB a la base de datos
	if err := sm.stateDB.Commit(true); err != nil {
		return fmt.Errorf("error haciendo commit del StateDB: %w", err)
	}
	
	// Guardar root hash en metadata storage
	stateData, err := json.Marshal(map[string]interface{}{
		"root":      root.Hex(),
		"height":    sm.getCurrentHeight(),
		"timestamp": sm.getCurrentTimestamp(),
	})
	if err != nil {
		return fmt.Errorf("error serializando estado: %w", err)
	}
	
	if err := sm.storage.SaveState(stateData); err != nil {
		return fmt.Errorf("error guardando estado: %w", err)
	}
	
	// Actualizar root hash local
	sm.stateRoot = root
	
	return nil
}

// SaveStateAtHeight guarda el estado en una altura específica
func (sm *StateManager) SaveStateAtHeight(height uint64) error {
	if sm.stateDB == nil {
		return fmt.Errorf("StateDB no está inicializado")
	}
	
	// Calcular root hash
	root := sm.stateDB.IntermediateRoot(true)
	
	// Commit el StateDB
	if err := sm.stateDB.Commit(true); err != nil {
		return fmt.Errorf("error haciendo commit del StateDB: %w", err)
	}
	
	// Guardar estado con altura
	stateData, err := json.Marshal(map[string]interface{}{
		"root":   root.Hex(),
		"height": height,
	})
	if err != nil {
		return fmt.Errorf("error serializando estado: %w", err)
	}
	
	// Guardar estado en altura específica usando método público de storage
	// Nota: Necesitamos agregar método SaveStateAtHeight a BlockchainDB
	// Por ahora, guardamos en metadata storage
	key := fmt.Sprintf("state:%d", height)
	if err := sm.storage.SaveAccount(key, stateData); err != nil {
		return fmt.Errorf("error guardando estado en altura %d: %w", height, err)
	}
	
	return nil
}

// LoadStateAtHeight carga el estado en una altura específica
func (sm *StateManager) LoadStateAtHeight(height uint64) (*state.StateDB, error) {
	// Obtener estado guardado en altura
	key := fmt.Sprintf("state:%d", height)
	stateData, err := sm.storage.GetAccount(key)
	if err != nil {
		return nil, fmt.Errorf("estado no encontrado en altura %d: %w", height, err)
	}
	
	// Parsear estado
	var stateInfo map[string]interface{}
	if err := json.Unmarshal(stateData, &stateInfo); err != nil {
		return nil, fmt.Errorf("error parseando estado: %w", err)
	}
	
	// Obtener root hash
	rootStr, ok := stateInfo["root"].(string)
	if !ok {
		return nil, fmt.Errorf("root hash no encontrado en estado")
	}
	
	root := common.HexToHash(rootStr)
	
	// Crear StateDB desde root
	stateDBPath := filepath.Join(sm.dataDir, "evm_state")
	db, err := rawdb.NewLevelDBDatabase(stateDBPath, 0, 0, "", false)
	if err != nil {
		return nil, fmt.Errorf("error creando base de datos EVM: %w", err)
	}
	
	database := state.NewDatabase(db)
	stateDB, err := state.New(root, database, nil)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error creando StateDB desde root: %w", err)
	}
	
	sm.stateDB = stateDB
	sm.database = database
	sm.stateRoot = root
	
	return stateDB, nil
}

// GetStateDB retorna el StateDB actual
func (sm *StateManager) GetStateDB() *state.StateDB {
	return sm.stateDB
}

// GetRootHash retorna el root hash actual
func (sm *StateManager) GetRootHash() common.Hash {
	if sm.stateDB == nil {
		return common.Hash{}
	}
	return sm.stateDB.IntermediateRoot(true)
}

// Close cierra la base de datos y libera recursos
func (sm *StateManager) Close() error {
	if sm.database != nil {
		if closer, ok := sm.database.(ethdb.Database); ok {
			return closer.Close()
		}
	}
	return nil
}

// getCurrentHeight obtiene la altura actual (helper)
func (sm *StateManager) getCurrentHeight() uint64 {
	height, err := sm.storage.GetLatestHeight()
	if err != nil {
		return 0
	}
	return height
}

// getCurrentTimestamp obtiene el timestamp actual (helper)
func (sm *StateManager) getCurrentTimestamp() int64 {
	// Obtener altura actual
	height, err := sm.storage.GetLatestHeight()
	if err != nil {
		return 0
	}

	// Obtener bloque más reciente
	blockData, err := sm.storage.GetBlock(height)
	if err != nil {
		return 0
	}

	// Parsear bloque para obtener timestamp
	// El timestamp se guarda como time.Time en el BlockHeader
	type BlockHeader struct {
		Timestamp time.Time `json:"timestamp"`
	}
	type Block struct {
		Header BlockHeader `json:"header"`
	}
	
	var block Block
	if err := json.Unmarshal(blockData, &block); err != nil {
		return 0
	}

	return block.Header.Timestamp.Unix()
}


package consensus

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"sort"
	"sync"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/ethereum/go-ethereum/common"
	"github.com/Q-YZX0/oxy-blockchain/internal/execution"
	"github.com/Q-YZX0/oxy-blockchain/internal/storage"
)

// Validator representa un validador en la red
type Validator struct {
	Address       string    // Dirección Ethereum del validador
	PubKey        []byte    // Clave pública CometBFT
	Stake         *big.Int  // Cantidad de OXG staked
	Power         int64     // Poder de voto (calculado del stake)
	DelegatedTo   string    // Dirección que tiene delegado el stake (opcional)
	Jailed        bool      // Si está en jail (slashed)
	JailedUntil   time.Time // Fecha hasta la que está en jail
	CreatedAt     time.Time // Fecha de creación
	LastActiveAt  time.Time // Última actividad
	MissedBlocks  int       // Bloques perdidos consecutivos
	TotalMissed   int       // Total de bloques perdidos
}

// ValidatorSet maneja el conjunto de validadores
type ValidatorSet struct {
	storage       *storage.BlockchainDB
	executor      *execution.EVMExecutor
	validators    map[string]*Validator
	mutex         sync.RWMutex
	minStake      *big.Int // Stake mínimo para ser validador
	maxValidators int      // Número máximo de validadores
}

// NewValidatorSet crea un nuevo conjunto de validadores
func NewValidatorSet(
	storage *storage.BlockchainDB,
	executor *execution.EVMExecutor,
	minStake *big.Int,
	maxValidators int,
) *ValidatorSet {
	return &ValidatorSet{
		storage:       storage,
		executor:      executor,
		validators:    make(map[string]*Validator),
		minStake:      minStake,
		maxValidators: maxValidators,
	}
}

// LoadValidators carga validadores desde storage
func (vs *ValidatorSet) LoadValidators() error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Cargar validadores guardados
	validatorsData, err := vs.storage.GetAccount("validators:set")
	if err != nil {
		// No hay validadores guardados, retornar nil
		log.Println("No hay validadores guardados, iniciando con set vacío")
		return nil
	}

	var validatorsList []*Validator
	if err := json.Unmarshal(validatorsData, &validatorsList); err != nil {
		return fmt.Errorf("error parseando validadores: %w", err)
	}

	vs.validators = make(map[string]*Validator)
	for _, v := range validatorsList {
		vs.validators[v.Address] = v
	}

	log.Printf("Cargados %d validadores desde storage", len(vs.validators))
	return nil
}

// SaveValidators guarda validadores en storage
func (vs *ValidatorSet) SaveValidators() error {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	validatorsList := make([]*Validator, 0, len(vs.validators))
	for _, v := range vs.validators {
		validatorsList = append(validatorsList, v)
	}

	validatorsData, err := json.Marshal(validatorsList)
	if err != nil {
		return fmt.Errorf("error serializando validadores: %w", err)
	}

	return vs.storage.SaveAccount("validators:set", validatorsData)
}

// RegisterValidator registra un nuevo validador
func (vs *ValidatorSet) RegisterValidator(
	address string,
	pubKey []byte,
	initialStake *big.Int,
) (*Validator, error) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Validar que no esté ya registrado
	if _, exists := vs.validators[address]; exists {
		return nil, fmt.Errorf("validador ya está registrado: %s", address)
	}

	// Validar stake mínimo
	if initialStake.Cmp(vs.minStake) < 0 {
		return nil, fmt.Errorf("stake insuficiente: requiere mínimo %s, tiene %s", vs.minStake.String(), initialStake.String())
	}

	// Validar que no se exceda el máximo
	if len(vs.validators) >= vs.maxValidators {
		// Encontrar validador con menor stake
		minStakeVal := vs.findLowestStakeValidator()
		if minStakeVal == nil || initialStake.Cmp(minStakeVal.Stake) <= 0 {
			return nil, fmt.Errorf("número máximo de validadores alcanzado y stake insuficiente para reemplazar al validador con menor stake")
		}

		// Remover validador con menor stake
		delete(vs.validators, minStakeVal.Address)
		log.Printf("Removido validador %s con stake %s para hacer espacio", minStakeVal.Address, minStakeVal.Stake.String())
	}

	// Crear nuevo validador
	validator := &Validator{
		Address:      address,
		PubKey:       pubKey,
		Stake:        new(big.Int).Set(initialStake),
		Power:        vs.calculatePower(initialStake),
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	vs.validators[address] = validator

	log.Printf("✅ Validador registrado: %s con stake %s", address, initialStake.String())

	// Guardar validadores
	if err := vs.SaveValidators(); err != nil {
		log.Printf("Advertencia: error guardando validadores: %v", err)
	}

	return validator, nil
}

// Stake aumenta el stake de un validador
func (vs *ValidatorSet) Stake(address string, amount *big.Int) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	validator, exists := vs.validators[address]
	if !exists {
		return fmt.Errorf("validador no encontrado: %s", address)
	}

	// Validar que no esté en jail
	if validator.Jailed {
		return fmt.Errorf("validador está en jail: %s", address)
	}

	// Actualizar stake
	validator.Stake.Add(validator.Stake, amount)
	validator.Power = vs.calculatePower(validator.Stake)
	validator.LastActiveAt = time.Now()

	log.Printf("✅ Stake actualizado para %s: %s (nuevo total: %s)", address, amount.String(), validator.Stake.String())

	// Guardar validadores
	if err := vs.SaveValidators(); err != nil {
		log.Printf("Advertencia: error guardando validadores: %v", err)
	}

	return nil
}

// Unstake reduce el stake de un validador
func (vs *ValidatorSet) Unstake(address string, amount *big.Int) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	validator, exists := vs.validators[address]
	if !exists {
		return fmt.Errorf("validador no encontrado: %s", address)
	}

	// Validar que no esté en jail
	if validator.Jailed {
		return fmt.Errorf("validador está en jail: %s", address)
	}

	// Validar que no baje del mínimo
	newStake := new(big.Int).Sub(validator.Stake, amount)
	if newStake.Cmp(vs.minStake) < 0 {
		return fmt.Errorf("no se puede unstake: quedaría con %s, requiere mínimo %s", newStake.String(), vs.minStake.String())
	}

	// Actualizar stake
	validator.Stake.Sub(validator.Stake, amount)
	validator.Power = vs.calculatePower(validator.Stake)
	validator.LastActiveAt = time.Now()

	log.Printf("✅ Stake reducido para %s: -%s (nuevo total: %s)", address, amount.String(), validator.Stake.String())

	// Si el stake es muy bajo, puede ser removido del set activo
	if validator.Stake.Cmp(vs.minStake) < 0 {
		delete(vs.validators, address)
		log.Printf("⚠️ Validador %s removido por stake insuficiente", address)
	}

	// Guardar validadores
	if err := vs.SaveValidators(); err != nil {
		log.Printf("Advertencia: error guardando validadores: %v", err)
	}

	return nil
}

// Slash penaliza a un validador por comportamiento malicioso
func (vs *ValidatorSet) Slash(address string, slashPercent int, jailDuration time.Duration) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	validator, exists := vs.validators[address]
	if !exists {
		return fmt.Errorf("validador no encontrado: %s", address)
	}

	// Calcular cantidad a slashear
	slashAmount := new(big.Int)
	slashAmount.Mul(validator.Stake, big.NewInt(int64(slashPercent)))
	slashAmount.Div(slashAmount, big.NewInt(100))

	// Reducir stake
	validator.Stake.Sub(validator.Stake, slashAmount)

	// Actualizar power
	validator.Power = vs.calculatePower(validator.Stake)

	// Enviar a jail
	validator.Jailed = true
	validator.JailedUntil = time.Now().Add(jailDuration)

	log.Printf("⚠️ Validador slasheado: %s -%s (%%%d)", address, slashAmount.String(), slashPercent)
	log.Printf("⛓️ Validador %s enviado a jail hasta %s", address, validator.JailedUntil.Format(time.RFC3339))

	// Si el stake es muy bajo después de slash, remover
	if validator.Stake.Cmp(vs.minStake) < 0 {
		delete(vs.validators, address)
		log.Printf("⚠️ Validador %s removido por stake insuficiente después de slash", address)
	}

	// Guardar validadores
	if err := vs.SaveValidators(); err != nil {
		log.Printf("Advertencia: error guardando validadores: %v", err)
	}

	return nil
}

// Unjail libera a un validador de jail
func (vs *ValidatorSet) Unjail(address string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	validator, exists := vs.validators[address]
	if !exists {
		return fmt.Errorf("validador no encontrado: %s", address)
	}

	if !validator.Jailed {
		return fmt.Errorf("validador no está en jail: %s", address)
	}

	if time.Now().Before(validator.JailedUntil) {
		return fmt.Errorf("validador aún está en jail hasta %s", validator.JailedUntil.Format(time.RFC3339))
	}

	// Liberar de jail
	validator.Jailed = false
	validator.JailedUntil = time.Time{}
	validator.MissedBlocks = 0

	log.Printf("✅ Validador %s liberado de jail", address)

	// Guardar validadores
	if err := vs.SaveValidators(); err != nil {
		log.Printf("Advertencia: error guardando validadores: %v", err)
	}

	return nil
}

// GetValidators retorna la lista de validadores
func (vs *ValidatorSet) GetValidators() []*Validator {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	validators := make([]*Validator, 0, len(vs.validators))
	for _, v := range vs.validators {
		if !v.Jailed {
			validators = append(validators, v)
		}
	}

	return validators
}

// GetActiveValidators retorna solo los validadores activos (no en jail)
func (vs *ValidatorSet) GetActiveValidators() []*Validator {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	validators := make([]*Validator, 0, len(vs.validators))
	for _, v := range vs.validators {
		if !v.Jailed && v.Stake.Cmp(vs.minStake) >= 0 {
			validators = append(validators, v)
		}
	}

	// Ordenar por stake (mayor primero)
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].Stake.Cmp(validators[j].Stake) > 0
	})

	return validators
}

// GetValidator retorna un validador por dirección
func (vs *ValidatorSet) GetValidator(address string) (*Validator, error) {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	validator, exists := vs.validators[address]
	if !exists {
		return nil, fmt.Errorf("validador no encontrado: %s", address)
	}

	return validator, nil
}

// ToCometBFTValidators convierte validadores a formato CometBFT
func (vs *ValidatorSet) ToCometBFTValidators() []abcitypes.ValidatorUpdate {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	activeValidators := vs.GetActiveValidators()
	updates := make([]abcitypes.ValidatorUpdate, 0, len(activeValidators))

	for _, v := range activeValidators {
		// Convertir clave pública a formato CometBFT
		var pubKey ed25519.PubKey
		if len(v.PubKey) == ed25519.PubKeySize {
			copy(pubKey[:], v.PubKey)
		} else {
			log.Printf("Advertencia: clave pública inválida para validador %s", v.Address)
			continue
		}

		updates = append(updates, abcitypes.ValidatorUpdate{
			PubKey: abcitypes.PubKey{
				Type: "/tm.PubKeyEd25519",
				Data: pubKey.Bytes(),
			},
			Power: v.Power,
		})
	}

	return updates
}

// calculatePower calcula el poder de voto basado en stake
func (vs *ValidatorSet) calculatePower(stake *big.Int) int64 {
	// Power máximo en CometBFT es 2^63 - 1
	// Usamos una proporción simple: 1 OXG = 1 power (con límites)
	maxPower := int64(1 << 30) // 1,073,741,824 (seguro para CometBFT)

	// Convertir stake a int64 (con límite)
	power := new(big.Int).Div(stake, big.NewInt(1e18)) // Dividir por 18 decimales
	if power.Cmp(big.NewInt(maxPower)) > 0 {
		return maxPower
	}

	return power.Int64()
}

// findLowestStakeValidator encuentra el validador con menor stake
func (vs *ValidatorSet) findLowestStakeValidator() *Validator {
	var lowest *Validator
	for _, v := range vs.validators {
		if lowest == nil || v.Stake.Cmp(lowest.Stake) < 0 {
			lowest = v
		}
	}
	return lowest
}

// UpdateValidatorActivity actualiza la última actividad de un validador
func (vs *ValidatorSet) UpdateValidatorActivity(address string, missedBlock bool) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	validator, exists := vs.validators[address]
	if !exists {
		return
	}

	validator.LastActiveAt = time.Now()

	if missedBlock {
		validator.MissedBlocks++
		validator.TotalMissed++

		// Si falla muchos bloques consecutivos, aplicar slash automático
		if validator.MissedBlocks >= 100 {
			log.Printf("⚠️ Validador %s ha fallado %d bloques consecutivos, aplicando slash", address, validator.MissedBlocks)
			// Slash del 5% del stake por cada 100 bloques perdidos
			slashPercentage := float64(validator.MissedBlocks/100) * 0.05 // 5% por cada 100 bloques
			if slashPercentage > 0.50 { // Máximo 50% de slash
				slashPercentage = 0.50
			}
			// Jail por 24 horas
			if err := vs.SlashValidator(address, slashPercentage, 24*time.Hour); err != nil {
				log.Printf("Error aplicando slash a validador %s: %v", address, err)
			} else {
				// Resetear contador después del slash
				validator.MissedBlocks = 0
			}
		}
	} else {
		validator.MissedBlocks = 0
	}
}

// RotateValidators rota los validadores según stake y actividad
func (vs *ValidatorSet) RotateValidators() ([]abcitypes.ValidatorUpdate, error) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Obtener validadores activos ordenados por stake
	activeValidators := vs.GetActiveValidators()

	// Limitar a máximo de validadores
	if len(activeValidators) > vs.maxValidators {
		activeValidators = activeValidators[:vs.maxValidators]
	}

	// Actualizar power de todos los validadores
	for _, v := range vs.validators {
		v.Power = vs.calculatePower(v.Stake)
	}

	// Convertir a formato CometBFT
	updates := vs.ToCometBFTValidators()

	log.Printf("🔄 Rotación de validadores: %d validadores activos", len(updates))

	return updates, nil
}

// InitializeGenesisValidators inicializa validadores desde genesis
func (vs *ValidatorSet) InitializeGenesisValidators(genesisValidators []GenesisValidator) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	for _, gv := range genesisValidators {
		validator := &Validator{
			Address:      gv.Address,
			PubKey:       gv.PubKey,
			Stake:        gv.Stake,
			Power:        vs.calculatePower(gv.Stake),
			CreatedAt:    time.Now(),
			LastActiveAt: time.Now(),
		}

		vs.validators[gv.Address] = validator
		log.Printf("✅ Validador genesis registrado: %s con stake %s", gv.Address, gv.Stake.String())
	}

	return vs.SaveValidators()
}

// GenesisValidator representa un validador en el genesis
type GenesisValidator struct {
	Address string
	PubKey  []byte
	Stake   *big.Int
}


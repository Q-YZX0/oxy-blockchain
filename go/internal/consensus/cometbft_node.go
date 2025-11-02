package consensus

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cometbft/cometbft/abci/server"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/Q-YZX0/oxy-blockchain/internal/execution"
	"github.com/Q-YZX0/oxy-blockchain/internal/storage"
)

// CometBFTNode maneja el nodo CometBFT
type CometBFTNode struct {
	node       *node.Node
	abciApp    *ABCIApp
	config     *Config
	running    bool
}

// NewCometBFTNode crea una nueva instancia del nodo CometBFT
func NewCometBFTNode(
	ctx context.Context,
	cfg *Config,
	storage *storage.BlockchainDB,
	executor *execution.EVMExecutor,
	validators *ValidatorSet,
) (*CometBFTNode, error) {
	
	// Crear aplicación ABCI con validators
	abciApp := NewABCIApp(storage, executor, validators, cfg.ChainID)
	
	// Crear configuración de CometBFT
	cometConfig := config.DefaultConfig()
	cometConfig.SetRoot(filepath.Join(cfg.DataDir, "cometbft"))
	
	// Asegurar que el directorio existe
	if err := os.MkdirAll(cometConfig.RootDir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio CometBFT: %w", err)
	}
	
	// Inicializar CometBFT si no existe
	if !isCometBFTInitialized(cometConfig) {
		if err := initializeCometBFT(cometConfig, cfg); err != nil {
			return nil, fmt.Errorf("error inicializando CometBFT: %w", err)
		}
	}
	
	// Cargar configuración
	if err := cometConfig.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("configuración inválida: %w", err)
	}
	
	// Cargar private validator
	pv, err := privval.LoadFilePV(
		cometConfig.PrivValidatorKeyFile(),
		cometConfig.PrivValidatorStateFile(),
	)
	if err != nil {
		return nil, fmt.Errorf("error cargando private validator: %w", err)
	}
	
	// Crear node key
	nodeKey, err := p2p.LoadNodeKey(cometConfig.NodeKeyFile())
	if err != nil {
		return nil, fmt.Errorf("error cargando node key: %w", err)
	}
	
	// Crear aplicación ABCI directamente (LocalClientCreator)
	// CometBFT usará LocalClientCreator para comunicarse in-process con la aplicación
	appCreator := proxy.NewLocalClientCreator(abciApp)
	
	// Crear nodo CometBFT
	cometNode, err := node.NewNode(
		cometConfig,
		pv,
		nodeKey,
		appCreator,
		node.DefaultGenesisDocProviderFunc(cometConfig),
		node.DefaultDBProvider,
		node.DefaultMetricsProvider(cometConfig.Instrumentation),
		log.New(os.Stderr, "", log.LstdFlags),
	)
	if err != nil {
		return nil, fmt.Errorf("error creando nodo CometBFT: %w", err)
	}
	
	return &CometBFTNode{
		node:    cometNode,
		abciApp: abciApp,
		config:  cfg,
		running: false,
	}, nil
}

// isCometBFTInitialized verifica si CometBFT ya está inicializado
func isCometBFTInitialized(cfg *config.Config) bool {
	genesisFile := filepath.Join(cfg.RootDir, "config", "genesis.json")
	_, err := os.Stat(genesisFile)
	return err == nil
}

// initializeCometBFT inicializa CometBFT usando el comando cometbft
func initializeCometBFT(cfg *config.Config, appConfig *Config) error {
	// Crear comando cometbft init
	cmd := exec.Command("cometbft", "init", "--home", cfg.RootDir)
	
	// Ejecutar comando
	if err := cmd.Run(); err != nil {
		// Si cometbft no está instalado, crear configuración manual
		return createCometBFTConfig(cfg, appConfig)
	}
	
	// Modificar genesis para incluir chain ID
	genesisFile := filepath.Join(cfg.RootDir, "config", "genesis.json")
	genesis, err := types.GenesisDocFromFile(genesisFile)
	if err != nil {
		return fmt.Errorf("error cargando genesis: %w", err)
	}
	
	genesis.ChainID = appConfig.ChainID
	
	if err := genesis.SaveAs(genesisFile); err != nil {
		return fmt.Errorf("error guardando genesis: %w", err)
	}
	
	return nil
}

// createCometBFTConfig crea configuración de CometBFT manualmente
func createCometBFTConfig(cfg *config.Config, appConfig *Config) error {
	// Crear directorios necesarios
	dirs := []string{
		filepath.Join(cfg.RootDir, "config"),
		filepath.Join(cfg.RootDir, "data"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creando directorio %s: %w", dir, err)
		}
	}
	
	// Crear genesis básico
	genesis := types.GenesisDoc{
		ChainID:         appConfig.ChainID,
		GenesisTime:     types.Now(),
		ConsensusParams: types.DefaultConsensusParams(),
	}
	
	genesisFile := filepath.Join(cfg.RootDir, "config", "genesis.json")
	if err := genesis.SaveAs(genesisFile); err != nil {
		return fmt.Errorf("error guardando genesis: %w", err)
	}
	
	// Generar claves si no existen
	if err := generateKeys(cfg); err != nil {
		return fmt.Errorf("error generando claves: %w", err)
	}
	
	return nil
}

// generateKeys genera claves para CometBFT si no existen
func generateKeys(cfg *config.Config) error {
	// TODO: Generar claves usando crypto de CometBFT
	// Por ahora, retornar nil si ya existen
	keyFile := cfg.PrivValidatorKeyFile()
	if _, err := os.Stat(keyFile); err == nil {
		return nil // Ya existe
	}
	
	// Generar nueva clave privada
	pv, err := privval.GenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
	if err != nil {
		return fmt.Errorf("error generando private validator: %w", err)
	}
	
	// Generar node key
	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return fmt.Errorf("error generando node key: %w", err)
	}
	
	_ = pv
	_ = nodeKey
	
	log.Println("Claves generadas para CometBFT")
	return nil
}


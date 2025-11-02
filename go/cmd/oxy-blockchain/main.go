package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"net/http"

	"github.com/Q-YZX0/oxy-blockchain/internal/config"
	"github.com/Q-YZX0/oxy-blockchain/internal/consensus"
	"github.com/Q-YZX0/oxy-blockchain/internal/execution"
	"github.com/Q-YZX0/oxy-blockchain/internal/storage"
	"github.com/Q-YZX0/oxy-blockchain/internal/network"
	"github.com/Q-YZX0/oxy-blockchain/internal/logger"
	"github.com/Q-YZX0/oxy-blockchain/internal/health"
	"github.com/Q-YZX0/oxy-blockchain/internal/metrics"
	"github.com/Q-YZX0/oxy-blockchain/internal/api"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configuración
	cfg := config.LoadConfig()

	// Inicializar logger estructurado
	useJSON := os.Getenv("OXY_LOG_JSON") == "true"
	logger.Init(cfg.LogLevel, useJSON)

	// Inicializar health checker y métricas
	healthChecker := health.NewHealthChecker()
	metricsInstance := metrics.NewMetrics()

	// Inicializar storage
	db, err := storage.NewBlockchainDB(cfg.DataDir)
	if err != nil {
		logger.Fatalf("Error inicializando storage: %v", err)
	}
	defer db.Close()

	// Reportar estado del storage al health checker
	healthChecker.SetStorageHealth(true)

	// Inicializar motor de ejecución (EVM)
	evm := execution.NewEVMExecutor(db, cfg.DataDir)
	
	// Iniciar ejecutor EVM
	if err := evm.Start(); err != nil {
		logger.Fatalf("Error iniciando ejecutor EVM: %v", err)
	}
	defer evm.Stop()

	// Reportar estado del EVM al health checker
	healthChecker.SetEVMHealth(true)

	// Inicializar conjunto de validadores
	minStake := big.NewInt(1000 * 1e18) // 1000 OXG mínimo (con 18 decimales)
	maxValidators := 100
	validators := consensus.NewValidatorSet(db, evm, minStake, maxValidators)
	
	// Cargar validadores guardados
	if err := validators.LoadValidators(); err != nil {
		logger.Warn().Err(err).Msg("Error cargando validadores")
	}

	// Inicializar consenso (CometBFT)
	consensusConfig := &consensus.Config{
		DataDir:       cfg.DataDir,
		ChainID:       cfg.ChainID,
		ValidatorAddr: cfg.ValidatorAddr,
		ValidatorKey:  cfg.ValidatorKey,
	}
	
	consensusEngine, err := consensus.NewCometBFT(ctx, consensusConfig, db, evm, validators)
	if err != nil {
		logger.Fatalf("Error inicializando consenso: %v", err)
	}

	// Reportar estado del consenso al health checker
	healthChecker.SetConsensusHealth(true)

	// Inicializar red P2P (integración con oxygen-sdk mesh)
	networkConfig := &network.Config{
		MeshEndpoint: cfg.MeshEndpoint,
		PeerID:       cfg.ValidatorAddr,
	}
	
	p2pNetwork, err := network.NewP2PNetwork(ctx, networkConfig, consensusEngine, db)
	if err != nil {
		logger.Fatalf("Error inicializando red P2P: %v", err)
	}

	// Iniciar componentes
	if err := consensusEngine.Start(); err != nil {
		logger.Fatalf("Error iniciando consenso: %v", err)
	}
	defer consensusEngine.Stop()

	if err := p2pNetwork.Start(); err != nil {
		logger.Fatalf("Error iniciando red P2P: %v", err)
	}
	defer p2pNetwork.Stop()

	// Reportar estado de la mesh network al health checker
	healthChecker.SetMeshHealth(true)

	// Iniciar servidor REST si está habilitado
	var restServer *api.RestServer
	if cfg.APIEnabled {
		restServer = api.NewRestServer(
			cfg.APIHost,
			cfg.APIPort,
			db,
			consensusEngine,
			healthChecker,
			metricsInstance,
			evm,
		)

		// Iniciar servidor REST en goroutine
		go func() {
			logger.Info().
				Str("host", cfg.APIHost).
				Str("port", cfg.APIPort).
				Msg("Iniciando servidor REST local")
			if err := restServer.Start(); err != nil && err != http.ErrServerClosed {
				logger.Error().Err(err).Msg("Error iniciando servidor REST")
			}
		}()

		defer func() {
			if restServer != nil {
				if err := restServer.Stop(); err != nil {
					logger.Error().Err(err).Msg("Error deteniendo servidor REST")
				}
			}
		}()
	}

	logger.Info().Msg("Oxy•gen Blockchain iniciada correctamente")

	// Manejar señales de terminación
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info().Msg("Deteniendo Oxy•gen Blockchain...")
}


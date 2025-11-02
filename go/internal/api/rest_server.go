package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/Q-YZX0/oxy-blockchain/internal/consensus"
	"github.com/Q-YZX0/oxy-blockchain/internal/execution"
	"github.com/Q-YZX0/oxy-blockchain/internal/health"
	"github.com/Q-YZX0/oxy-blockchain/internal/metrics"
	"github.com/Q-YZX0/oxy-blockchain/internal/storage"
)

// RestServer maneja el servidor HTTP REST local
type RestServer struct {
	host          string
	port          string
	storage       *storage.BlockchainDB
	consensus     *consensus.CometBFT
	healthChecker *health.HealthChecker
	metrics       *metrics.Metrics
	executor      *execution.EVMExecutor
	server        *http.Server
}

// NewRestServer crea un nuevo servidor REST
func NewRestServer(
	host string,
	port string,
	storage *storage.BlockchainDB,
	consensus *consensus.CometBFT,
	healthChecker *health.HealthChecker,
	metrics *metrics.Metrics,
	executor *execution.EVMExecutor,
) *RestServer {
	return &RestServer{
		host:          host,
		port:          port,
		storage:       storage,
		consensus:     consensus,
		healthChecker: healthChecker,
		metrics:       metrics,
		executor:      executor,
	}
}

// Start inicia el servidor REST
func (s *RestServer) Start() error {
	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/api/v1/blocks/", s.handleBlocks)
	mux.HandleFunc("/api/v1/transactions/", s.handleTransactions)
	mux.HandleFunc("/api/v1/accounts/", s.handleAccounts)
	mux.HandleFunc("/api/v1/submit-tx", s.handleSubmitTx)

	// Middleware CORS básico
	handler := s.corsMiddleware(mux)

	addr := s.host + ":" + s.port
	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s.server.ListenAndServe()
}

// Stop detiene el servidor REST
func (s *RestServer) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// corsMiddleware añade headers CORS
func (s *RestServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleHealth maneja el endpoint /health
func (s *RestServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := s.healthChecker.CheckHealth()
	
	w.Header().Set("Content-Type", "application/json")
	
	// Retornar código HTTP apropiado
	if status.Status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else if status.Status == "degraded" {
		w.WriteHeader(http.StatusOK) // 200 pero con advertencia
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(status)
}

// handleMetrics maneja el endpoint /metrics
func (s *RestServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	m := s.metrics.GetMetrics()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(m)
}

// handleBlocks maneja /api/v1/blocks/{height} o /api/v1/blocks/latest
func (s *RestServer) handleBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer height del path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/blocks/")
	
	var block *consensus.Block
	var err error

	if path == "latest" || path == "" {
		// Obtener último bloque
		block, err = s.consensus.GetLatestBlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Parsear altura
		height, parseErr := strconv.ParseUint(path, 10, 64)
		if parseErr != nil {
			http.Error(w, "Invalid block height", http.StatusBadRequest)
			return
		}

		// Obtener bloque por altura
		blockData, dbErr := s.storage.GetBlock(height)
		if dbErr != nil {
			http.Error(w, "Block not found", http.StatusNotFound)
			return
		}

		if err := json.Unmarshal(blockData, &block); err != nil {
			http.Error(w, "Error decoding block", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(block)
}

// handleTransactions maneja /api/v1/transactions/{hash}
func (s *RestServer) handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer hash del path
	txHash := r.URL.Path[len("/api/v1/transactions/"):]

	// Obtener transacción desde storage
	txData, err := s.storage.GetTransaction(txHash)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(txData)
}

// handleAccounts maneja /api/v1/accounts/{address}
func (s *RestServer) handleAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer dirección del path
	address := strings.TrimPrefix(r.URL.Path, "/api/v1/accounts/")
	
	// Validar dirección
	if !common.IsHexAddress(address) {
		http.Error(w, "Invalid Ethereum address", http.StatusBadRequest)
		return
	}

	// Obtener estado de cuenta desde el executor EVM
	if s.executor == nil {
		http.Error(w, "EVM executor not available", http.StatusServiceUnavailable)
		return
	}

	accountState, err := s.executor.GetState(address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting account state: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accountState)
}

// handleSubmitTx maneja /api/v1/submit-tx
func (s *RestServer) handleSubmitTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decodificar transacción del body
	var tx consensus.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, fmt.Sprintf("Invalid transaction format: %v", err), http.StatusBadRequest)
		return
	}

	// Validar transacción básica
	if tx.Hash == "" {
		http.Error(w, "Transaction hash required", http.StatusBadRequest)
		return
	}

	if tx.From == "" {
		http.Error(w, "Transaction from address required", http.StatusBadRequest)
		return
	}

	// Enviar transacción al consensus
	if err := s.consensus.SubmitTransaction(&tx); err != nil {
		http.Error(w, fmt.Sprintf("Error submitting transaction: %v", err), http.StatusBadRequest)
		return
	}

	// Retornar confirmación
	response := map[string]interface{}{
		"success": true,
		"hash":    tx.Hash,
		"message": "Transaction submitted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}


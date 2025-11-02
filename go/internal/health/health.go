package health

import (
	"sync"
	"time"
)

// HealthStatus representa el estado de salud del nodo
type HealthStatus struct {
	Status      string                 `json:"status"`      // "healthy", "degraded", "unhealthy"
	Timestamp   time.Time              `json:"timestamp"`
	Components  map[string]ComponentStatus `json:"components"`
	BlockHeight uint64                 `json:"block_height"`
	Peers       int                    `json:"peers"`
}

// ComponentStatus representa el estado de un componente
type ComponentStatus struct {
	Status    string    `json:"status"`    // "ok", "warning", "error"
	Message   string    `json:"message,omitempty"`
	LastCheck time.Time `json:"last_check"`
}

// HealthChecker maneja el estado de salud del nodo
type HealthChecker struct {
	mu              sync.RWMutex
	components      map[string]ComponentStatus
	blockHeight     uint64
	peers           int
	storageHealthy  bool
	evmHealthy      bool
	consensusHealthy bool
	meshHealthy     bool
}

// NewHealthChecker crea un nuevo verificador de salud
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		components: make(map[string]ComponentStatus),
	}
}

// CheckHealth retorna el estado de salud actual
func (h *HealthChecker) CheckHealth() HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	status := "healthy"
	
	// Verificar estado general
	allHealthy := true
	anyWarning := false
	
	for _, comp := range h.components {
		if comp.Status == "error" {
			allHealthy = false
			status = "unhealthy"
			break
		} else if comp.Status == "warning" {
			anyWarning = true
		}
	}

	if anyWarning && allHealthy {
		status = "degraded"
	}

	// Verificar componentes críticos
	if !h.storageHealthy || !h.evmHealthy || !h.consensusHealthy {
		status = "unhealthy"
	}

	return HealthStatus{
		Status:      status,
		Timestamp:   time.Now(),
		Components:  h.components,
		BlockHeight: h.blockHeight,
		Peers:       h.peers,
	}
}

// UpdateComponent actualiza el estado de un componente
func (h *HealthChecker) UpdateComponent(name string, status string, message string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.components[name] = ComponentStatus{
		Status:    status,
		Message:   message,
		LastCheck: time.Now(),
	}
}

// SetBlockHeight actualiza la altura del bloque
func (h *HealthChecker) SetBlockHeight(height uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.blockHeight = height
}

// SetPeers actualiza el número de peers
func (h *HealthChecker) SetPeers(count int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.peers = count
}

// SetStorageHealth actualiza el estado del storage
func (h *HealthChecker) SetStorageHealth(healthy bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.storageHealthy = healthy
	if healthy {
		h.UpdateComponent("storage", "ok", "Storage operativo")
	} else {
		h.UpdateComponent("storage", "error", "Storage no disponible")
	}
}

// SetEVMHealth actualiza el estado del EVM
func (h *HealthChecker) SetEVMHealth(healthy bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.evmHealthy = healthy
	if healthy {
		h.UpdateComponent("evm", "ok", "EVM operativo")
	} else {
		h.UpdateComponent("evm", "error", "EVM no disponible")
	}
}

// SetConsensusHealth actualiza el estado del consenso
func (h *HealthChecker) SetConsensusHealth(healthy bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.consensusHealthy = healthy
	if healthy {
		h.UpdateComponent("consensus", "ok", "Consenso operativo")
	} else {
		h.UpdateComponent("consensus", "error", "Consenso no disponible")
	}
}

// SetMeshHealth actualiza el estado de la mesh network
func (h *HealthChecker) SetMeshHealth(healthy bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.meshHealthy = healthy
	if healthy {
		h.UpdateComponent("mesh", "ok", "Mesh network operativa")
	} else {
		h.UpdateComponent("mesh", "warning", "Mesh network degradada")
	}
}

// IsHealthy retorna si el nodo está saludable
func (h *HealthChecker) IsHealthy() bool {
	status := h.CheckHealth()
	return status.Status == "healthy"
}


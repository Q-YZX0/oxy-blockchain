package metrics

import (
	"sync"
	"time"
)

// Metrics almacena métricas del nodo
type Metrics struct {
	mu sync.RWMutex

	// Métricas de bloques
	BlocksProcessed    uint64
	BlockProcessingTime time.Duration
	
	// Métricas de transacciones
	TransactionsProcessed uint64
	TransactionsRejected  uint64
	TransactionsPerSecond  float64
	
	// Métricas de red
	PeersConnected     int
	MessagesReceived   uint64
	MessagesSent       uint64
	
	// Métricas de estado
	CurrentBlockHeight uint64
	StateDBSize        uint64
	MempoolSize        int
	
	// Métricas de rendimiento
	AverageGasUsed     uint64
	TotalGasUsed       uint64
	
	// Timestamps
	LastBlockTime      time.Time
	Uptime             time.Duration
	StartTime          time.Time
}

// NewMetrics crea una nueva instancia de métricas
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
}

// GetMetrics retorna una copia de las métricas actuales
func (m *Metrics) GetMetrics() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Calcular uptime
	uptime := time.Since(m.StartTime)
	
	return Metrics{
		BlocksProcessed:       m.BlocksProcessed,
		BlockProcessingTime:   m.BlockProcessingTime,
		TransactionsProcessed: m.TransactionsProcessed,
		TransactionsRejected:  m.TransactionsRejected,
		TransactionsPerSecond: m.calculateTPS(),
		PeersConnected:        m.PeersConnected,
		MessagesReceived:      m.MessagesReceived,
		MessagesSent:          m.MessagesSent,
		CurrentBlockHeight:    m.CurrentBlockHeight,
		StateDBSize:           m.StateDBSize,
		MempoolSize:           m.MempoolSize,
		AverageGasUsed:        m.AverageGasUsed,
		TotalGasUsed:          m.TotalGasUsed,
		LastBlockTime:         m.LastBlockTime,
		Uptime:                uptime,
		StartTime:             m.StartTime,
	}
}

// IncrementBlocks incrementa el contador de bloques procesados
func (m *Metrics) IncrementBlocks() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BlocksProcessed++
	m.LastBlockTime = time.Now()
}

// AddBlockProcessingTime añade tiempo de procesamiento de bloque
func (m *Metrics) AddBlockProcessingTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.BlocksProcessed > 0 {
		m.BlockProcessingTime = (m.BlockProcessingTime + duration) / 2
	} else {
		m.BlockProcessingTime = duration
	}
}

// IncrementTransactions incrementa el contador de transacciones procesadas
func (m *Metrics) IncrementTransactions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TransactionsProcessed++
}

// IncrementRejectedTransactions incrementa el contador de transacciones rechazadas
func (m *Metrics) IncrementRejectedTransactions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TransactionsRejected++
}

// SetPeers actualiza el número de peers conectados
func (m *Metrics) SetPeers(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PeersConnected = count
}

// IncrementMessagesReceived incrementa el contador de mensajes recibidos
func (m *Metrics) IncrementMessagesReceived() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesReceived++
}

// IncrementMessagesSent incrementa el contador de mensajes enviados
func (m *Metrics) IncrementMessagesSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesSent++
}

// SetBlockHeight actualiza la altura del bloque actual
func (m *Metrics) SetBlockHeight(height uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CurrentBlockHeight = height
}

// SetStateDBSize actualiza el tamaño del StateDB
func (m *Metrics) SetStateDBSize(size uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StateDBSize = size
}

// SetMempoolSize actualiza el tamaño del mempool
func (m *Metrics) SetMempoolSize(size int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MempoolSize = size
}

// AddGasUsed añade gas usado a las métricas
func (m *Metrics) AddGasUsed(gas uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalGasUsed += gas
	if m.TransactionsProcessed > 0 {
		m.AverageGasUsed = m.TotalGasUsed / m.TransactionsProcessed
	}
}

// calculateTPS calcula transacciones por segundo
func (m *Metrics) calculateTPS() float64 {
	uptime := time.Since(m.StartTime).Seconds()
	if uptime > 0 {
		return float64(m.TransactionsProcessed) / uptime
	}
	return 0.0
}

// Reset resetea todas las métricas
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BlocksProcessed = 0
	m.TransactionsProcessed = 0
	m.TransactionsRejected = 0
	m.MessagesReceived = 0
	m.MessagesSent = 0
	m.TotalGasUsed = 0
	m.AverageGasUsed = 0
	m.StartTime = time.Now()
}


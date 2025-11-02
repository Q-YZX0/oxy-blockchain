package consensus

import (
	"sync"
	"time"
)

// RateLimiter maneja rate limiting de transacciones
type RateLimiter struct {
	mu sync.RWMutex

	// Límites por dirección
	perAddressLimit int           // Transacciones permitidas por ventana
	timeWindow      time.Duration // Ventana de tiempo
	transactions    map[string][]time.Time

	// Límite global del mempool
	mempoolSizeLimit int
}

// NewRateLimiter crea un nuevo rate limiter
func NewRateLimiter(perAddressLimit int, timeWindow time.Duration, mempoolSizeLimit int) *RateLimiter {
	return &RateLimiter{
		perAddressLimit:  perAddressLimit,
		timeWindow:        timeWindow,
		transactions:      make(map[string][]time.Time),
		mempoolSizeLimit:  mempoolSizeLimit,
	}
}

// Allow verifica si una transacción desde una dirección está permitida
func (rl *RateLimiter) Allow(address string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.timeWindow)

	// Limpiar transacciones antiguas
	if times, exists := rl.transactions[address]; exists {
		// Filtrar transacciones dentro de la ventana de tiempo
		validTimes := make([]time.Time, 0)
		for _, t := range times {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}
		rl.transactions[address] = validTimes
	}

	// Verificar límite
	if times, exists := rl.transactions[address]; exists {
		if len(times) >= rl.perAddressLimit {
			return false
		}
	}

	// Registrar nueva transacción
	if rl.transactions[address] == nil {
		rl.transactions[address] = make([]time.Time, 0)
	}
	rl.transactions[address] = append(rl.transactions[address], now)

	return true
}

// GetCount retorna el número de transacciones en la ventana para una dirección
func (rl *RateLimiter) GetCount(address string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-rl.timeWindow)

	if times, exists := rl.transactions[address]; exists {
		count := 0
		for _, t := range times {
			if t.After(cutoff) {
				count++
			}
		}
		return count
	}

	return 0
}

// Cleanup limpia transacciones antiguas de todas las direcciones
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.timeWindow)

	for address, times := range rl.transactions {
		validTimes := make([]time.Time, 0)
		for _, t := range times {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}

		if len(validTimes) == 0 {
			delete(rl.transactions, address)
		} else {
			rl.transactions[address] = validTimes
		}
	}
}

// StartCleanup inicia una goroutine que limpia periódicamente
func (rl *RateLimiter) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			rl.Cleanup()
		}
	}()
}

// CheckMempoolSize verifica si el mempool puede aceptar más transacciones
func (rl *RateLimiter) CheckMempoolSize(currentSize int) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return currentSize < rl.mempoolSizeLimit
}

// GetMempoolSizeLimit retorna el límite de tamaño del mempool
func (rl *RateLimiter) GetMempoolSizeLimit() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.mempoolSizeLimit
}


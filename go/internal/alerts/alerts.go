package alerts

import (
	"fmt"
	"sync"
	"time"
)

// AlertLevel representa el nivel de una alerta
type AlertLevel string

const (
	AlertLevelInfo    AlertLevel = "info"
	AlertLevelWarning AlertLevel = "warning"
	AlertLevelError   AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert representa una alerta
type Alert struct {
	Level     AlertLevel `json:"level"`
	Message   string     `json:"message"`
	Component string     `json:"component"`
	Timestamp time.Time  `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// AlertHandler maneja alertas del sistema
type AlertHandler struct {
	mu          sync.RWMutex
	alerts      []Alert
	maxAlerts   int
	callbacks   []AlertCallback
}

// AlertCallback es una función que se llama cuando se emite una alerta
type AlertCallback func(Alert)

// NewAlertHandler crea un nuevo manejador de alertas
func NewAlertHandler(maxAlerts int) *AlertHandler {
	return &AlertHandler{
		alerts:    make([]Alert, 0),
		maxAlerts: maxAlerts,
		callbacks: make([]AlertCallback, 0),
	}
}

// RegisterCallback registra un callback para alertas
func (ah *AlertHandler) RegisterCallback(callback AlertCallback) {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	ah.callbacks = append(ah.callbacks, callback)
}

// Emit emite una nueva alerta
func (ah *AlertHandler) Emit(level AlertLevel, component string, message string, details map[string]interface{}) {
	alert := Alert{
		Level:     level,
		Message:   message,
		Component: component,
		Timestamp: time.Now(),
		Details:   details,
	}

	ah.mu.Lock()
	defer ah.mu.Unlock()

	// Agregar alerta a la lista
	ah.alerts = append(ah.alerts, alert)

	// Limitar número de alertas guardadas
	if len(ah.alerts) > ah.maxAlerts {
		// Remover las más antiguas
		ah.alerts = ah.alerts[len(ah.alerts)-ah.maxAlerts:]
	}

	// Llamar callbacks
	for _, callback := range ah.callbacks {
		go callback(alert)
	}
}

// EmitInfo emite una alerta de información
func (ah *AlertHandler) EmitInfo(component string, message string) {
	ah.Emit(AlertLevelInfo, component, message, nil)
}

// EmitWarning emite una alerta de advertencia
func (ah *AlertHandler) EmitWarning(component string, message string, details map[string]interface{}) {
	ah.Emit(AlertLevelWarning, component, message, details)
}

// EmitError emite una alerta de error
func (ah *AlertHandler) EmitError(component string, message string, details map[string]interface{}) {
	ah.Emit(AlertLevelError, component, message, details)
}

// EmitCritical emite una alerta crítica
func (ah *AlertHandler) EmitCritical(component string, message string, details map[string]interface{}) {
	ah.Emit(AlertLevelCritical, component, message, details)
}

// GetAlerts retorna las alertas recientes
func (ah *AlertHandler) GetAlerts(limit int) []Alert {
	ah.mu.RLock()
	defer ah.mu.RUnlock()

	if limit <= 0 || limit > len(ah.alerts) {
		limit = len(ah.alerts)
	}

	// Retornar las últimas N alertas
	start := len(ah.alerts) - limit
	if start < 0 {
		start = 0
	}

	alerts := make([]Alert, limit)
	copy(alerts, ah.alerts[start:])
	return alerts
}

// GetRecentAlerts retorna alertas recientes filtradas por nivel
func (ah *AlertHandler) GetRecentAlerts(level AlertLevel, since time.Time) []Alert {
	ah.mu.RLock()
	defer ah.mu.RUnlock()

	var filtered []Alert
	for _, alert := range ah.alerts {
		if alert.Level == level && alert.Timestamp.After(since) {
			filtered = append(filtered, alert)
		}
	}

	return filtered
}

// ClearAlerts limpia todas las alertas
func (ah *AlertHandler) ClearAlerts() {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	ah.alerts = make([]Alert, 0)
}

// DefaultLogCallback es un callback que loguea alertas
func DefaultLogCallback(alert Alert) {
	// Este callback puede ser implementado para loguear alertas
	// Por ejemplo, usando el logger estructurado
	fmt.Printf("[%s] %s - %s: %s\n", 
		alert.Level, 
		alert.Component, 
		alert.Timestamp.Format(time.RFC3339), 
		alert.Message)
}


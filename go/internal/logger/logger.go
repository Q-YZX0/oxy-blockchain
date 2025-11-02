package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var globalLogger zerolog.Logger

func init() {
	// Configurar logger global
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Usar output de consola con formato humano-legible en desarrollo
	// En producción, usar JSON formatter
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	globalLogger = log.Output(output)
}

// Init inicializa el logger con configuración
func Init(logLevel string, useJSON bool) {
	// Configurar nivel de log
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configurar output
	if useJSON {
		// JSON format para producción
		globalLogger = zerolog.New(os.Stderr).With().
			Timestamp().
			Logger()
	} else {
		// Console format para desarrollo
		output := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
		globalLogger = log.Output(output)
	}
}

// Logger retorna el logger global configurado
func Logger() zerolog.Logger {
	return globalLogger
}

// WithContext añade contexto al logger
func WithContext(fields map[string]interface{}) zerolog.Logger {
	logger := globalLogger
	for k, v := range fields {
		logger = logger.With().Interface(k, v).Logger()
	}
	return logger
}

// Debug logs a debug message
func Debug(msg string) {
	globalLogger.Debug().Msg(msg)
}

// Info logs an info message
func Info(msg string) {
	globalLogger.Info().Msg(msg)
}

// Warn logs a warning message
func Warn(msg string) {
	globalLogger.Warn().Msg(msg)
}

// Error logs an error message
func Error(msg string) {
	globalLogger.Error().Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	globalLogger.Fatal().Msg(msg)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	globalLogger.Debug().Msgf(format, args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	globalLogger.Info().Msgf(format, args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	globalLogger.Warn().Msgf(format, args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	globalLogger.Error().Msgf(format, args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatal().Msgf(format, args...)
}

// WithField añade un campo al logger
func WithField(key string, value interface{}) zerolog.Logger {
	return globalLogger.With().Interface(key, value).Logger()
}

// WithFields añade múltiples campos al logger
func WithFields(fields map[string]interface{}) zerolog.Logger {
	return WithContext(fields)
}

// WithError añade un error al logger
func WithError(err error) zerolog.Logger {
	return globalLogger.With().Err(err).Logger()
}


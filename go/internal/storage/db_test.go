package storage

import (
	"os"
	"testing"
)

// TestBlockchainDBBasic verifica operaciones básicas de la base de datos
func TestBlockchainDBBasic(t *testing.T) {
	// Crear directorio temporal para tests
	tmpDir := "./test_data"
	defer os.RemoveAll(tmpDir)
	
	db, err := NewBlockchainDB(tmpDir)
	if err != nil {
		t.Fatalf("Error creando base de datos: %v", err)
	}
	defer db.Close()
	
	// Test SaveBlock y GetBlock
	blockData := []byte("test block data")
	height := uint64(1)
	
	if err := db.SaveBlock(height, blockData); err != nil {
		t.Fatalf("Error guardando bloque: %v", err)
	}
	
	retrieved, err := db.GetBlock(height)
	if err != nil {
		t.Fatalf("Error obteniendo bloque: %v", err)
	}
	
	if string(retrieved) != string(blockData) {
		t.Errorf("Datos del bloque no coinciden: esperado %s, obtenido %s", blockData, retrieved)
	}
	
	// Test SaveTransaction y GetTransaction
	txHash := "0x1234567890abcdef"
	txData := []byte("test transaction")
	
	if err := db.SaveTransaction(txHash, txData); err != nil {
		t.Fatalf("Error guardando transacción: %v", err)
	}
	
	retrievedTx, err := db.GetTransaction(txHash)
	if err != nil {
		t.Fatalf("Error obteniendo transacción: %v", err)
	}
	
	if string(retrievedTx) != string(txData) {
		t.Errorf("Datos de transacción no coinciden")
	}
	
	// Test SaveLatestHeight y GetLatestHeight
	testHeight := uint64(42)
	if err := db.SaveLatestHeight(testHeight); err != nil {
		t.Fatalf("Error guardando altura: %v", err)
	}
	
	retrievedHeight, err := db.GetLatestHeight()
	if err != nil {
		t.Fatalf("Error obteniendo altura: %v", err)
	}
	
	if retrievedHeight != testHeight {
		t.Errorf("Altura no coincide: esperado %d, obtenido %d", testHeight, retrievedHeight)
	}
}


package consensus

import (
	"encoding/json"
	"testing"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	execution "github.com/Q-YZX0/oxy-blockchain/internal/execution"
	"github.com/Q-YZX0/oxy-blockchain/internal/storage"
)

// TestABCIApp_BasicFlow prueba el flujo básico de ABCI
func TestABCIApp_BasicFlow(t *testing.T) {
	// Crear storage temporal
	db, err := storage.NewBlockchainDB("/tmp/test-blockchain")
	if err != nil {
		t.Fatalf("Error creando storage: %v", err)
	}
	defer db.Close()

	// Crear ejecutor EVM
	evm := execution.NewEVMExecutor(db, "/tmp/test-blockchain")
	if err := evm.Start(); err != nil {
		t.Fatalf("Error iniciando EVM: %v", err)
	}
	defer evm.Stop()

	// Crear ABCI app
	app := NewABCIApp(db, evm, nil, "test-chain")

	// Probar InitChain
	initChainReq := abcitypes.RequestInitChain{
		ChainId: "test-chain",
		Validators: []abcitypes.ValidatorUpdate{},
	}
	resp := app.InitChain(initChainReq)
	if resp.Validators == nil {
		t.Error("InitChain debe retornar validadores")
	}

	// Probar Info
	infoReq := abcitypes.RequestInfo{}
	infoResp := app.Info(infoReq)
	if infoResp.Data == "" {
		t.Error("Info debe retornar datos")
	}

	// Probar BeginBlock
	beginBlockReq := abcitypes.RequestBeginBlock{
		Header: abcitypes.Header{
			Height: 1,
		},
	}
	beginBlockResp := app.BeginBlock(beginBlockReq)
	if beginBlockResp.Events == nil {
		t.Error("BeginBlock debe retornar eventos")
	}

	// Probar EndBlock
	endBlockReq := abcitypes.RequestEndBlock{
		Height: 1,
	}
	endBlockResp := app.EndBlock(endBlockReq)
	if endBlockResp.Events == nil {
		t.Error("EndBlock debe retornar eventos")
	}

	// Probar Commit
	commitResp := app.Commit()
	if commitResp.Data == nil {
		t.Error("Commit debe retornar AppHash")
	}
}

// TestABCIApp_CheckTx prueba validación de transacciones
func TestABCIApp_CheckTx(t *testing.T) {
	// Crear storage temporal
	db, err := storage.NewBlockchainDB("/tmp/test-blockchain-checktx")
	if err != nil {
		t.Fatalf("Error creando storage: %v", err)
	}
	defer db.Close()

	// Crear ejecutor EVM
	evm := execution.NewEVMExecutor(db, "/tmp/test-blockchain-checktx")
	if err := evm.Start(); err != nil {
		t.Fatalf("Error iniciando EVM: %v", err)
	}
	defer evm.Stop()

	// Crear ABCI app
	app := NewABCIApp(db, evm, nil, "test-chain")

	// Generar clave privada para testing
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Error generando clave: %v", err)
	}

	fromAddr := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Crear transacción de prueba
	tx := Transaction{
		From:     fromAddr.Hex(),
		To:       common.HexToAddress("0x0000000000000000000000000000000000000000").Hex(),
		Value:    "1000000000000000000", // 1 OXG
		GasLimit: 21000,
		GasPrice: "1000000000", // 1 gwei
		Nonce:    0,
		Hash:     "0x1234567890abcdef",
		Signature: []byte{},
	}

	// Serializar transacción
	txData, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("Error serializando transacción: %v", err)
	}

	// Probar CheckTx
	checkTxReq := abcitypes.RequestCheckTx{
		Tx: txData,
	}

	checkTxResp := app.CheckTx(checkTxReq)

	// La transacción debería ser rechazada por falta de firma válida
	if checkTxResp.Code == 0 {
		t.Error("CheckTx debería rechazar transacción sin firma válida")
	}
}

// TestABCIApp_Query prueba el sistema de queries
func TestABCIApp_Query(t *testing.T) {
	// Crear storage temporal
	db, err := storage.NewBlockchainDB("/tmp/test-blockchain-query")
	if err != nil {
		t.Fatalf("Error creando storage: %v", err)
	}
	defer db.Close()

	// Crear ejecutor EVM
	evm := execution.NewEVMExecutor(db, "/tmp/test-blockchain-query")
	if err := evm.Start(); err != nil {
		t.Fatalf("Error iniciando EVM: %v", err)
	}
	defer evm.Stop()

	// Crear ABCI app
	app := NewABCIApp(db, evm, nil, "test-chain")

	// Probar query de altura
	queryReq := abcitypes.RequestQuery{
		Path: "height",
	}

	queryResp := app.Query(queryReq)
	if queryResp.Code != 0 {
		t.Errorf("Query de altura falló: %s", queryResp.Log)
	}

	// Probar query de balance
	queryReq.Path = "balance/0x0000000000000000000000000000000000000000"
	queryResp = app.Query(queryReq)
	if queryResp.Code != 0 {
		t.Errorf("Query de balance falló: %s", queryResp.Log)
	}
}


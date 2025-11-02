package consensus

import (
	"encoding/json"
	"time"
)

// BlockHeader representa el header de un bloque
type BlockHeader struct {
	Height     uint64
	Hash       string
	ParentHash string
	Timestamp  time.Time
	Validator  string
	ChainID    string
}

// Block representa un bloque completo en la blockchain
type Block struct {
	Header       BlockHeader
	Transactions []*Transaction
	Receipts     []*TransactionReceipt
}

// Transaction representa una transacción
type Transaction struct {
	Hash        string
	From        string
	To          string
	Value       string
	Data        []byte
	GasLimit    uint64
	GasPrice    string
	Nonce       uint64
	Signature   []byte // Firma de la transacción
	Timestamp   int64
}

// TransactionReceipt representa el recibo de una transacción
type TransactionReceipt struct {
	TransactionHash string
	BlockHash       string
	BlockNumber     uint64
	GasUsed         uint64
	Status          string // "success" o "failed"
	Logs            []Log
	Error           string
}

// Log representa un evento emitido por un contrato
type Log struct {
	Address     string
	Topics      []string
	Data        []byte
	BlockNumber uint64
	TxHash      string
}

// MarshalJSON implementa json.Marshaler para Block
func (b *Block) MarshalJSON() ([]byte, error) {
	type Alias Block
	return json.Marshal(&struct {
		*Alias
		Header BlockHeader `json:"header"`
	}{
		Alias:  (*Alias)(b),
		Header: b.Header,
	})
}

// UnmarshalJSON implementa json.Unmarshaler para Block
func (b *Block) UnmarshalJSON(data []byte) error {
	type Alias Block
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	return json.Unmarshal(data, aux)
}


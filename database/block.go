package database

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Hash is type for hashed db
type Hash [32]byte

// MarshalText will convert hash to byte
// ref: https://golang.org/src/encoding/json/encode.go?s=6458:6501#L148
// TODO: replace h[:] to h
func (h Hash) MarshalText() ([]byte, error) {
	return []byte(h.Hex()), nil
}

// UnmarshalText will convert byte to hash
// ref: https://golang.org/src/encoding/json/decode.go?s=4081:4129#L86
// TODO: replace h[:] to h
func (h *Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)
	return err
}

// Hex encode hash to hex string
// TODO: replace h[:] to h
func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

// BlockHeader is block metadata
type BlockHeader struct {
	Parent Hash   `json:"parent"` // parent block reference
	Number uint64 `json:"number"`
	Time   uint64 `json:"time"`
}

// BlockFS store unique hash from a block
type BlockFS struct {
	Key   Hash  `json:"hash"`
	Value Block `json:"block"`
}

// Block contains batches of transactions and hashed
type Block struct {
	Header BlockHeader `json:"header"`
	TXs    []Tx        `json:"payload"` // new transactions only (payload)
}

// NewBlock will return new block
func NewBlock(parent Hash, number uint64, time uint64, txs []Tx) Block {
	return Block{
		Header: BlockHeader{
			Parent: parent,
			Number: number,
			Time:   time,
		},
		TXs: txs,
	}
}

// Hash will return hash from a block
func (b Block) Hash() (Hash, error) {
	blockJSON, err := json.Marshal(b)
	if err != nil {
		return Hash{}, err
	}

	return sha256.Sum256(blockJSON), nil
}

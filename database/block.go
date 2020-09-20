package database

import (
	"crypto/sha256"
	"encoding/json"
)

// Hash is type for hashed db
type Hash [32]byte

// BlockHeader is block metadata
type BlockHeader struct {
	Parent Hash // parent block reference
	Time   uint64
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
func NewBlock(parent Hash, time uint64, txs []Tx) Block {
	return Block{
		Header: BlockHeader{
			Parent: parent,
			Time:   time,
		},
		TXs: txs,
	}
}

// Hash will return hash from a block
func (b Block) Hash() (Hash, error) {
	blockJSON, err := json.Marshal(b)
	return sha256.Sum256(blockJSON), err
}

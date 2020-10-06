package database

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

//BlockReward is reward for miner
const BlockReward = 100

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

// IsEmpty verify if a hash is empty
func (h Hash) IsEmpty() bool {
	emptyHash := Hash{}

	return bytes.Equal(emptyHash[:], h[:])
}

// BlockHeader is block metadata
type BlockHeader struct {
	Parent Hash    `json:"parent"` // parent block reference
	Number uint64  `json:"number"`
	Nonce  uint32  `json:"nonce"`
	Time   uint64  `json:"time"`
	Miner  Account `json:"miner"`
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
func NewBlock(parent Hash, number uint64, nonce uint32, time uint64, miner Account, txs []Tx) Block {
	return Block{
		Header: BlockHeader{
			Parent: parent,
			Number: number,
			Nonce:  nonce,
			Miner:  miner,
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

// IsBlockHashValid check is block hash valid
func IsBlockHashValid(hash Hash) bool {
	return fmt.Sprintf("%x", hash[0]) == "0" &&
		fmt.Sprintf("%x", hash[1]) == "0" &&
		fmt.Sprintf("%x", hash[2]) == "0" &&
		fmt.Sprintf("%x", hash[3]) != "0"
}

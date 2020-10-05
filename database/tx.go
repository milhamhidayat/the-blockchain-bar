package database

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

// Account is an alias for customer represented in DB
type Account string

// NewAccount return new customer account
func NewAccount(value string) Account {
	return Account(value)
}

// Tx represent each transaction in database
type Tx struct {
	From  Account `json:"from"`
	To    Account `json:"to"`
	Value uint    `json:"value"`
	Data  string  `json:"data"`
	Time  uint64  `json:"time"`
}

// NewTx return new transaction
func NewTx(from Account, to Account, value uint, data string) Tx {
	return Tx{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
		Time:  uint64(time.Now().Unix()),
	}
}

// IsReward check if transaction is eligible for a reward
func (t Tx) IsReward() bool {
	return t.Data == "reward"
}

// Hash return hash of txJSON
func (t Tx) Hash() (Hash, error) {
	txJSON, err := json.Marshal(t)
	if err != nil {
		return Hash{}, err
	}

	return sha256.Sum256(txJSON), nil
}

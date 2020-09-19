package database

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// State represent business logic for db component
// Know all user balances
// and who transferred tbb tokens to whom,
// and how many were transferred
type State struct {
	Balances  map[Account]uint
	txMempool []Tx
	dbFile    *os.File
}

// NewStateFromDisk update transaction data
func NewStateFromDisk() (*State, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	genFilePath := filepath.Join(cwd, "database", "genesis.json")
	gen, err := loadGenesis(genFilePath)
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	txDbFilePath := filepath.Join(cwd, "database", "tx.db")
	f, err := os.OpenFile(txDbFilePath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	state := &State{
		balances,
		make([]Tx, 0),
		f,
	}

	// Iterate over each the tx.db file's line by line
	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			return nil, err
		}

		// Convert JSON encoded TX into an object (struct)
		var tx Tx
		err = json.Unmarshal(scanner.Bytes(), &tx)
		if err != nil {
			return nil, err
		}

		// rebuild the state (user balances),
		// as a series of events
		err = state.apply(tx)
		if err != nil {
			return nil, err
		}
	}
	return state, nil
}

// Add will add new transactions to mempool
// Mempool is a collection of all token transactions
// awaiting verifications and confirmation which will be inclused
// in the next block
// https://medium.com/ecoinomic/what-is-the-bitcoin-mempool-and-why-does-it-matter-c7a9ed2859ff
func (s *State) Add(tx Tx) error {
	err := s.apply(tx)
	if err != nil {
		return err
	}

	s.txMempool = append(s.txMempool, tx)
	return nil
}

// Persist will write transactions to disk
func (s *State) Persist() error {
	// Make a copy of mempool because the s.txMempool will be modified
	// in the loop below
	mempool := make([]Tx, len(s.txMempool))
	copy(mempool, s.txMempool)

	fmt.Println("======== mempool ========")
	fmt.Printf("%+v\n", mempool)
	fmt.Println("=================")

	for i := 0; i < len(mempool); i++ {
		txJSON, err := json.Marshal(s.txMempool[i])
		if err != nil {
			return err
		}

		_, err = s.dbFile.Write(append(txJSON, '\n'))
		if err != nil {
			return err
		}

		// remove the tx written to a file from the mempool
		s.txMempool = append(s.txMempool[:i], s.txMempool[i+1:]...)

		fmt.Println("++++++++ mempool loop++++++++")
		fmt.Printf("%+v\n", s.txMempool)
		fmt.Println("+++++++++++++++++")
	}

	fmt.Println("++++++++ mempool final++++++++")
	fmt.Printf("%+v\n", s.txMempool)
	fmt.Println("+++++++++++++++++")

	return nil
}

// Close will close tx db file
func (s *State) Close() {
	s.dbFile.Close()
}

// apply will change and validate the state
func (s *State) apply(tx Tx) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if s.Balances[tx.From] < tx.Value {
		return errors.New("insufficient balances")
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}

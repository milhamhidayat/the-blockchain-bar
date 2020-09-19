package database

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Snapshot is a transaction db contents which has been hash-ed
type Snapshot [32]byte

// State represent business logic for db component
// Know all user balances
// and who transferred tbb tokens to whom,
// and how many were transferred
type State struct {
	Balances  map[Account]uint
	txMempool []Tx
	dbFile    *os.File
	snapshot  Snapshot
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
		Snapshot{},
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

	err = state.doSnapShot()
	if err != nil {
		return nil, err
	}

	return state, nil
}

// LatestSnapshot return latest hashed state snapshot
func (s *State) LatestSnapshot() Snapshot {
	return s.snapshot
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
func (s *State) Persist() (Snapshot, error) {
	// Make a copy of mempool because the s.txMempool will be modified
	// in the loop below
	mempool := make([]Tx, len(s.txMempool))
	copy(mempool, s.txMempool)

	for i := 0; i < len(mempool); i++ {
		txJSON, err := json.Marshal(s.txMempool[i])
		if err != nil {
			return Snapshot{}, err
		}

		fmt.Println("Persisting new TX to disk:")
		fmt.Printf("\t%s\n", txJSON)
		_, err = s.dbFile.Write(append(txJSON, '\n'))
		if err != nil {
			return Snapshot{}, err
		}

		err = s.doSnapShot()
		if err != nil {
			return Snapshot{}, err
		}
		fmt.Printf("New DB Snapshot: %x\n", s.snapshot)

		// remove the tx written to a file from the mempool
		s.txMempool = append(s.txMempool[:i], s.txMempool[i+1:]...)
	}

	return s.snapshot, nil
}

// Close will close tx db file
func (s *State) Close() error {
	return s.dbFile.Close()
}

// apply will change and validate the state
func (s *State) apply(tx Tx) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if s.Balances[tx.From]-tx.Value < 0 {
		return errors.New("insufficient balances")
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}

// doSnapShot will take hash snapshot from tx.db
func (s *State) doSnapShot() error {
	_, err := s.dbFile.Seek(0, 0)
	if err != nil {
		return err
	}

	txsData, err := ioutil.ReadAll(s.dbFile)
	if err != nil {
		return err
	}

	s.snapshot = sha256.Sum256(txsData)
	return nil
}

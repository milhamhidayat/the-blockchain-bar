package database

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// State represent business logic for db component
// Know all user balances
// and who transferred tbb tokens to whom,
// and how many were transferred
type State struct {
	Balances        map[Account]uint
	txMempool       []Tx
	dbFile          *os.File
	latestBlock     Block
	latestBlockHash Hash
}

// NewStateFromDisk update transaction data
func NewStateFromDisk(dataDir string) (*State, error) {
	dataDir = ExpandPath(dataDir)

	err := initDataDirIfNotExists(dataDir)
	if err != nil {
		return nil, err
	}

	gen, err := loadGenesis(getGenesisJSONFilePath(dataDir))
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	f, err := os.OpenFile(getBlocksDbFilePath(dataDir), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	state := &State{
		balances,
		make([]Tx, 0),
		f,
		Block{},
		Hash{},
	}

	// Iterate over each the tx.db file's line by line
	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			return nil, err
		}

		var blockFs BlockFS
		blockFsJSON := scanner.Bytes()

		if len(blockFsJSON) == 0 {
			break
		}

		err = json.Unmarshal(blockFsJSON, &blockFs)
		if err != nil {
			return nil, err
		}

		err = state.applyBlock(blockFs.Value)
		if err != nil {
			return nil, err
		}

		state.latestBlock = blockFs.Value
		state.latestBlockHash = blockFs.Key
	}
	return state, nil
}

// LatestBlock return latest block
func (s *State) LatestBlock() Block {
	return s.latestBlock
}

// LatestBlockHash return latest block hash
func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

// AddBlock adds new block to blockchain
func (s *State) AddBlock(b Block) error {
	for _, tx := range b.TXs {
		err := s.AddTx(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddTx will add new transactions to mempool
// Mempool is a collection of all token transactions
// awaiting verifications and confirmation which will be inclused
// in the next block
// https://medium.com/ecoinomic/what-is-the-bitcoin-mempool-and-why-does-it-matter-c7a9ed2859ff
func (s *State) AddTx(tx Tx) error {
	err := s.apply(tx)
	if err != nil {
		return err
	}

	s.txMempool = append(s.txMempool, tx)
	return nil
}

// Persist will write transactions to disk
func (s *State) Persist() (Hash, error) {
	latestBlockHash, err := s.latestBlock.Hash()
	if err != nil {
		return Hash{}, err
	}

	block := NewBlock(
		latestBlockHash,
		s.latestBlock.Header.Number+1,
		uint64(time.Now().Unix()),
		s.txMempool)
	blockHash, err := block.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFs := BlockFS{blockHash, block}
	blockFsJSON, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, err
	}

	fmt.Println("Persisting new block to disk")
	fmt.Printf("\t%s\n", blockFsJSON)

	_, err = s.dbFile.Write(append(blockFsJSON, '\n'))
	if err != nil {
		return Hash{}, err
	}
	s.latestBlockHash = latestBlockHash
	s.latestBlock = block
	s.txMempool = []Tx{}

	return latestBlockHash, nil
}

// Close will close tx db file
func (s *State) Close() error {
	return s.dbFile.Close()
}

func (s *State) applyBlock(b Block) error {
	for _, tx := range b.TXs {
		err := s.apply(tx)
		if err != nil {
			return err
		}
	}
	return nil
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

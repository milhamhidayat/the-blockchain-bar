package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
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
	hasGenesisBlock bool
}

// NewStateFromDisk update transaction data
func NewStateFromDisk(dataDir string) (*State, error) {
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

	dbFilePath := getBlocksDbFilePath(dataDir)
	f, err := os.OpenFile(dbFilePath, os.O_APPEND|os.O_RDWR, 0600)
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
		false,
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

		err = applyTXs(blockFs.Value.TXs, state)
		if err != nil {
			return nil, err
		}

		state.latestBlock = blockFs.Value
		state.latestBlockHash = blockFs.Key
		state.hasGenesisBlock = true
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

// AddBlocks will add multiple blokcs
func (s *State) AddBlocks(blocks []Block) error {
	for _, b := range blocks {
		_, err := s.AddBlock(b)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddBlock adds new block to blockchain
// Bug: for the header
func (s *State) AddBlock(b Block) (Hash, error) {
	pendingState := s.copy()

	// validate block meta + payload
	err := applyBlock(b, pendingState)
	if err != nil {
		return Hash{}, err
	}

	blockHash, err := b.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFs := BlockFS{blockHash, b}

	blockFsJSON, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, err
	}

	fmt.Println("Persisting new Block to disk")
	fmt.Printf("\t%s\n", blockFsJSON)

	_, err = s.dbFile.Write(append(blockFsJSON, '\n'))
	if err != nil {
		return Hash{}, err
	}

	s.Balances = pendingState.Balances
	s.latestBlockHash = blockHash
	s.latestBlock = b
	s.hasGenesisBlock = true

	return blockHash, nil
}

// NextBlockNumber will return next block header number
func (s *State) NextBlockNumber() uint64 {
	if !s.hasGenesisBlock {
		return uint64(0)
	}

	return s.LatestBlock().Header.Number + 1
}

// Close will close tx db file
func (s *State) Close() error {
	return s.dbFile.Close()
}

func (s *State) copy() State {
	c := State{}
	c.hasGenesisBlock = s.hasGenesisBlock
	c.latestBlock = s.latestBlock
	c.latestBlockHash = s.latestBlockHash
	c.txMempool = make([]Tx, len(s.txMempool))
	c.Balances = make(map[Account]uint)

	for acc, balance := range s.Balances {
		c.Balances[acc] = balance
	}

	for _, tx := range s.txMempool {
		c.txMempool = append(c.txMempool, tx)
	}

	return c
}

// applyBlock verifies if block can be added to the blockchain
// Block meatadata are verified as well as transactions within (sufficient balances, etc).
func applyBlock(b Block, s State) error {
	nextExpectedBlockNumber := s.latestBlock.Header.Number + 1

	if b.Header.Number != nextExpectedBlockNumber {
		return fmt.Errorf("next expected block must '%d' be '%d'", nextExpectedBlockNumber, b.Header.Number)
	}

	if s.hasGenesisBlock && s.latestBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.Parent, s.latestBlockHash) {
		return fmt.Errorf("next block parent hash must be '%x' not '%x'", s.latestBlockHash, b.Header.Parent)
	}

	return applyTXs(b.TXs, &s)
}

// applyTXs will validate list of transaction
func applyTXs(txs []Tx, s *State) error {
	for _, tx := range txs {
		err := applyTx(tx, s)
		if err != nil {
			return err
		}
	}

	return nil
}

// apply will change and validate the transaction
func applyTx(tx Tx, s *State) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if tx.Value > s.Balances[tx.From] {
		return fmt.Errorf("wrong TX. Sender '%s' balance is %d TBB. TX cost is %d TBB", tx.From, s.Balances[tx.From], tx.Value)
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}

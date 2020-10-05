package node

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"the-blockchain-bar/database"
	"the-blockchain-bar/fs"
	"time"
)

// PendingBlock is a block where waiting to be validate
type PendingBlock struct {
	parent database.Hash
	number uint64
	time   uint64
	txs    []database.Tx
}

// NewPendingBlock will return new pending block
func NewPendingBlock(
	parent database.Hash,
	number uint64,
	txs []database.Tx) PendingBlock {
	return PendingBlock{
		parent: parent,
		number: number,
		time:   uint64(time.Now().Unix()),
		txs:    txs,
	}
}

// Mine will mine token by validating transaction (consensus)
func Mine(ctx context.Context, pb PendingBlock) (database.Block, error) {
	if len(pb.txs) == 0 {
		return database.Block{}, errors.New("mining empty blocks is not allowed")
	}

	start := time.Now()
	attempt := 0
	var (
		block database.Block
		hash  database.Hash
		nonce uint32
	)

	for !database.IsBlockHashValid(hash) {
		select {
		case <-ctx.Done():
			fmt.Println("mining cancelled")
			return database.Block{}, fmt.Errorf("mining cancelled. %s", ctx.Err())
		default:
		}

		attempt++
		nonce = generateNonce()

		if attempt%1000000 == 0 || attempt == 1 {
			fmt.Printf("mining %d pending TXs. Attempt: %d\n", len(pb.txs), attempt)
		}

		block = database.NewBlock(
			pb.parent,
			pb.number,
			nonce,
			pb.time,
			pb.txs,
		)

		blockHash, err := block.Hash()
		if err != nil {
			return database.Block{}, fmt.Errorf("couldn't mine block. %s", err.Error())
		}

		hash = blockHash
	}

	fmt.Printf("Mined new block '%x' using PoW 🥳 %s:\n", hash, fs.Unicode("\\UIF389"))
	fmt.Printf("Height: '%v'\n", pb.number)
	fmt.Printf("Nonce: '%v'", nonce)
	fmt.Printf("Created: '%v'\n", pb.time)
	fmt.Printf("Parent: '%v'\n\n", pb.parent.Hex())
	fmt.Printf("Attempt: '%v'\n", attempt)
	fmt.Printf("Time: %s\n\n", time.Since(start))

	return block, nil
}

func generateNonce() uint32 {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Uint32()
}

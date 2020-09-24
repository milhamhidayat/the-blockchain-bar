package database

import (
	"bufio"
	"encoding/json"
	"os"
	"reflect"
)

// GetBlocksAfter will get next block node
func GetBlocksAfter(blockHash Hash, dataDir string) ([]Block, error) {
	f, err := os.OpenFile(
		getBlocksDbFilePath(dataDir),
		os.O_RDONLY,
		0600)
	if err != nil {
		return nil, err
	}

	blocks := make([]Block, 0)
	shouldStartCollecting := false

	if reflect.DeepEqual(blockHash, Hash{}) {
		shouldStartCollecting = true
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			return nil, err
		}

		var blockFS BlockFS
		err = json.Unmarshal(scanner.Bytes(), &blockFS)
		if err != nil {
			return nil, err
		}

		if shouldStartCollecting {
			blocks = append(blocks, blockFS.Value)
			continue
		}

		if blockHash == blockFS.Key {
			shouldStartCollecting = true
		}
	}

	return blocks, nil
}

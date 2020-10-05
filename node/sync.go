package node

import (
	"context"
	"fmt"
	"net/http"
	"the-blockchain-bar/database"
	"time"
)

func (n *Node) sync(ctx context.Context) error {
	ticker := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-ticker.C:
			n.doSync()
		case <-ctx.Done():
			ticker.Stop()
		}
	}
}

func (n *Node) doSync() {
	for _, peer := range n.knownPeers {
		if n.info.IP == peer.IP && n.info.Port == peer.Port {
			continue
		}

		fmt.Printf("Searching for new Peers and their Blocks and Peers: '%s'\n", peer.TCPAddress())

		status, err := queryPeerStatus(peer)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			fmt.Printf("Peer '%s' was removed from KnownPeers\n", peer.TCPAddress())
			n.RemovePeer(peer)
			continue
		}

		err = n.joinKnownPeers(peer)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}

		err = n.syncBlocks(peer, status)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}

		err = n.syncBlocks(peer, status)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}

		err = n.syncKnownPeers(status)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}

		err = n.syncPendingTXs(peer, status.PendingTXs)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			continue
		}

		return
	}
}

func (n *Node) syncBlocks(peer PeerNode, status StatusRes) error {
	localBlockNumber := n.state.LatestBlock().Header.Number

	//if the peer has no blocks, ignore it
	if status.Hash.IsEmpty() {
		return nil
	}

	// if the peer has less blocks than us, ignore it
	if status.Number < localBlockNumber {
		return nil
	}

	// if it's the genesis block and we already synced it, ignore it
	if status.Number == 0 && !n.state.LatestBlockHash().IsEmpty() {
		return nil
	}

	newBlocksCount := status.Number - localBlockNumber
	if localBlockNumber == 0 && status.Number == 0 {
		newBlocksCount = 1
	}
	fmt.Printf("found %d new blocks from peer %s\n", newBlocksCount, peer.TCPAddress())

	blocks, err := fetchBlocksFromPeer(peer, n.state.LatestBlockHash())
	if err != nil {
		return err
	}

	for _, block := range blocks {
		_, err = n.state.AddBlock(block)
		if err != nil {
			return err
		}
		n.newSyncedBlocks <- block
	}

	return nil
}

func (n *Node) syncKnownPeers(status StatusRes) error {
	for _, statusPeer := range status.KnownPeers {
		if !n.IsKnownPeer(statusPeer) {
			fmt.Printf("found new peer %s\n", statusPeer.TCPAddress())
			n.AddPeer(statusPeer)
		}
	}

	return nil
}

func (n *Node) syncPendingTXs(peer PeerNode, txs []database.Tx) error {
	for _, tx := range txs {
		err := n.AddPendingTX(tx, peer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) joinKnownPeers(peer PeerNode) error {
	if peer.connected {
		return nil
	}

	url := fmt.Sprintf(
		"http://%s%s?%s=%s&%s=%d",
		peer.TCPAddress(),
		endPointAddPeer,
		endPointAddPeerQueryKeyIP,
		n.info.IP,
		endpointAddPeerQueryKeyPort,
		n.info.Port,
	)

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	addPeerRes := AddPeerRes{}
	err = readRes(res, &addPeerRes)
	if err != nil {
		return err
	}
	if addPeerRes.Error != "" {
		return fmt.Errorf(addPeerRes.Error)
	}

	knownPeer := n.knownPeers[peer.TCPAddress()]
	knownPeer.connected = addPeerRes.Success

	n.AddPeer(knownPeer)

	if !addPeerRes.Success {
		return fmt.Errorf("unable to join KnownPeers of '%s'", peer.TCPAddress())
	}

	return nil
}

func queryPeerStatus(peer PeerNode) (StatusRes, error) {
	url := fmt.Sprintf("http://%s%s", peer.TCPAddress(), endPointStatus)
	res, err := http.Get(url)
	if err != nil {
		return StatusRes{}, err
	}

	statusRes := StatusRes{}
	err = readRes(res, &statusRes)
	if err != nil {
		return StatusRes{}, err
	}

	return statusRes, nil
}

func fetchBlocksFromPeer(peer PeerNode, fromBlock database.Hash) ([]database.Block, error) {
	fmt.Printf("imporint blocks from Peer %s... \n", peer.TCPAddress())

	url := fmt.Sprintf(
		"http://%s%s?%s=%s",
		peer.TCPAddress(),
		endPointSync,
		endPointSyncQueryKeyFromBlock,
		fromBlock.Hex(),
	)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	syncRes := SyncRes{}
	err = readRes(res, &syncRes)
	if err != nil {
		return nil, err
	}

	return syncRes.Blocks, nil
}

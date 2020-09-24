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
		if n.ip == peer.IP && n.port == peer.Port {
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

		err = n.syncKnownPeers(peer, status)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}
	}
}

func (n *Node) syncBlocks(peer PeerNode, status StatusRes) error {
	localBlockNumber := n.state.LatestBlock().Header.Number
	if localBlockNumber < status.Number {
		newBlocksCount := status.Number - localBlockNumber

		fmt.Printf("Found %d new blocks from Peer %s\n", newBlocksCount, peer.TCPAddress())

		blocks, err := fetchBlocksFromPeer(peer, n.state.LatestBlockHash())
		if err != nil {
			return err
		}

		err = n.state.AddBlocks(blocks)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) syncKnownPeers(peer PeerNode, status StatusRes) error {
	for _, statusPeer := range status.KnownPeers {
		if !n.IsKnownPeer(statusPeer) {
			fmt.Printf("found new peer %s\n", statusPeer.TCPAddress())
			n.AddPeer(statusPeer)
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
		n.ip,
		endpointAddPeerQueryKeyPort,
		n.port,
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

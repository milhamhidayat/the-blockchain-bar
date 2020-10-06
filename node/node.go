package node

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"the-blockchain-bar/database"
)

const (
	// DefaultMiner is default miner value
	DefaultMiner = ""
	// DefaultIP is default ip for api
	DefaultIP = "127.0.0.1"
	// DefaultHTTPPort is default http port for api
	DefaultHTTPPort = 8080

	endPointStatus                = "/node/status"
	endPointSync                  = "/node/sync"
	endPointSyncQueryKeyFromBlock = "fromBlock"

	endPointAddPeer              = "/node/peer"
	endPointAddPeerQueryKeyIP    = "ip"
	endpointAddPeerQueryKeyPort  = "port"
	endpointAddPeerQueryKeyMiner = "miner"

	miningIntervalSeconds = 10
)

// PeerNode is node owned by other user
// connected to blockchain that can be peered
type PeerNode struct {
	IP          string           `json:"ip"`
	Port        uint64           `json:"port"`
	IsBootstrap bool             `json:"is_bootstrap"`
	Account     database.Account `json:"account"`
	// Check whet node already connected, sync with this Peer
	connected bool
}

// TCPAddress return ip address with port
func (pn PeerNode) TCPAddress() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}

// Node is consist of host ledger and smart contract
type Node struct {
	dataDir         string
	info            PeerNode
	state           *database.State
	knownPeers      map[string]PeerNode
	pendingTXs      map[string]database.Tx
	archivedTXs     map[string]database.Tx
	newSyncedBlocks chan database.Block
	newPendingTXs   chan database.Tx
	isMining        bool
}

// New will return new node
func New(dataDir string, ip string, port uint64, acc database.Account, bootstrap PeerNode) *Node {
	knownPeers := make(map[string]PeerNode)
	knownPeers[bootstrap.TCPAddress()] = bootstrap
	return &Node{
		dataDir: dataDir,
		info: NewPeerNode(
			ip,
			port,
			false,
			acc,
			true,
		),
		knownPeers:      knownPeers,
		pendingTXs:      make(map[string]database.Tx),
		archivedTXs:     make(map[string]database.Tx),
		newSyncedBlocks: make(chan database.Block),
		newPendingTXs:   make(chan database.Tx, 10000),
		isMining:        false,
	}
}

// NewPeerNode will return new peer node
func NewPeerNode(ip string, port uint64, isBootstrap bool, acc database.Account, connected bool) PeerNode {
	return PeerNode{
		IP:          ip,
		Port:        port,
		IsBootstrap: isBootstrap,
		Account:     acc,
		connected:   connected,
	}
}

// Run will run rest API
func (n *Node) Run(ctx context.Context) error {
	fmt.Println(fmt.Sprintf("Listening on %s:%d", n.info.IP, n.info.Port))

	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	n.state = state

	fmt.Println("blockchain state:")
	fmt.Printf("- height: %d\n", n.state.LatestBlock().Header.Number)
	fmt.Printf("- hash: %s\n", n.state.LatestBlockHash().Hex())

	go n.sync(ctx)
	go n.mine(ctx)

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalancesHandler(w, r, state)
	})

	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, n)
	})

	http.HandleFunc(endPointStatus, func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, n)
	})

	http.HandleFunc(endPointSync, func(w http.ResponseWriter, r *http.Request) {
		syncHandler(w, r, n)
	})

	http.HandleFunc(endPointAddPeer, func(w http.ResponseWriter, r *http.Request) {
		addPeerHandler(w, r, n)
	})

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", n.info.Port),
	}
	go func() {
		<-ctx.Done()
		_ = server.Close()
	}()

	return server.ListenAndServe()
}

// AddPeer will add new peer to known peers
func (n *Node) AddPeer(peer PeerNode) {
	n.knownPeers[peer.TCPAddress()] = peer
}

// RemovePeer remove known peers
func (n *Node) RemovePeer(peer PeerNode) {
	delete(n.knownPeers, peer.TCPAddress())
}

// IsKnownPeer check if a peer is known by bootstrap node
func (n *Node) IsKnownPeer(peer PeerNode) bool {
	if peer.IP == n.info.IP && peer.Port == n.info.Port {
		return true
	}

	_, isKnownPeer := n.knownPeers[peer.TCPAddress()]
	return isKnownPeer
}

func (n *Node) mine(ctx context.Context) error {
	var (
		miningCtx         context.Context
		stopCurrentMining context.CancelFunc
	)

	ticker := time.NewTicker(time.Second * miningIntervalSeconds)
	for {
		select {
		case <-ticker.C:
			go func() {
				if len(n.pendingTXs) > 0 && !n.isMining {
					n.isMining = true
				}

				miningCtx, stopCurrentMining = context.WithCancel(ctx)
				err := n.minePendingTXs(miningCtx)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err)
				}
				n.isMining = false
			}()
		case block, _ := <-n.newSyncedBlocks:
			if n.isMining {
				blockHash, _ := block.Hash()
				fmt.Printf("\nPeer mined next block '%s' faster :\n", blockHash.Hex())
				n.removeMinedPendingTXs(block)
				stopCurrentMining()
			}
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

func (n *Node) minePendingTXs(ctx context.Context) error {
	blockToMine := NewPendingBlock(
		n.state.LatestBlockHash(),
		n.state.LatestBlock().Header.Number+1,
		n.info.Account,
		n.getPendingTXsAsArray(),
	)

	minedBlock, err := Mine(ctx, blockToMine)
	if err != nil {
		return err
	}

	n.removeMinedPendingTXs(minedBlock)

	_, err = n.state.AddBlock(minedBlock)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) removeMinedPendingTXs(block database.Block) {
	if len(block.TXs) > 0 && len(n.pendingTXs) > 0 {
		fmt.Println("updateing in memory pending TXS Pool:")
	}

	for _, tx := range block.TXs {
		txHash, _ := tx.Hash()
		if _, exists := n.pendingTXs[txHash.Hex()]; exists {
			fmt.Printf("\t-archiving mined TX: %s\n", txHash.Hex())
			n.archivedTXs[txHash.Hex()] = tx
			delete(n.pendingTXs, txHash.Hex())
		}
	}
}

// AddPendingTX will add pending tx
func (n *Node) AddPendingTX(tx database.Tx, fromPeer PeerNode) error {
	txHash, err := tx.Hash()
	if err != nil {
		return err
	}

	txJSON, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	_, isAlredyPending := n.pendingTXs[txHash.Hex()]
	_, isArchived := n.archivedTXs[txHash.Hex()]

	if !isAlredyPending && !isArchived {
		fmt.Printf("added pending tx %s from Peer %s\n", txJSON, fromPeer.TCPAddress())
		n.pendingTXs[txHash.Hex()] = tx
		n.newPendingTXs <- tx
	}

	return nil
}

func (n *Node) getPendingTXsAsArray() []database.Tx {
	txs := make([]database.Tx, 0)
	for _, tx := range n.pendingTXs {
		txs = append(txs, tx)
	}

	return txs
}

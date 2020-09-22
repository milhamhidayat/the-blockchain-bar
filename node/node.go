package node

import (
	"context"
	"fmt"
	"net/http"

	"the-blockchain-bar/database"
)

const (
	// DefaultHTTPPort is default http port for api
	DefaultHTTPPort               = 8080
	endPointStatus                = "/node/status"
	endPointSync                  = "/node/sync"
	endPointSyncQueryKeyFromBlock = "fromBlock"
)

// PeerNode is node owned by other user
// connected to blockchain that can be peered
type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`
	IsActive    bool   `json:"is_active"`
}

// TCPAddress return ip address with port
func (pn PeerNode) TCPAddress() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}

// Node is consist of host ledger and smart contract
type Node struct {
	dataDir    string
	port       uint64
	state      *database.State
	knownPeers map[string]PeerNode
}

// New will return new node
func New(dataDir string, port uint64, bootstrap PeerNode) *Node {
	knownPeers := make(map[string]PeerNode)
	knownPeers[bootstrap.TCPAddress()] = bootstrap
	return &Node{
		dataDir:    dataDir,
		port:       port,
		knownPeers: knownPeers,
	}
}

// NewPeerNode will return new peer node
func NewPeerNode(ip string, port uint64, isBootstrap bool, isActive bool) PeerNode {
	return PeerNode{
		IP:          ip,
		Port:        port,
		IsBootstrap: isBootstrap,
		IsActive:    isActive,
	}
}

// Run will run rest API
func (n *Node) Run() error {
	ctx := context.Background()
	fmt.Println(fmt.Sprintf("Listening on HTTP port: %d", n.port))

	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	n.state = state

	go n.sync(ctx)

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalancesHandler(w, r, state)
	})

	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, state)
	})

	http.HandleFunc(endPointStatus, func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, n)
	})

	http.HandleFunc(endPointSync, func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, n)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)
}

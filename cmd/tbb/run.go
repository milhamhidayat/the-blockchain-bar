package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"the-blockchain-bar/database"
	"the-blockchain-bar/node"
)

func runCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launches the TBB node and its HTTP API",
		Run: func(cmd *cobra.Command, args []string) {
			miner, _ := cmd.Flags().GetString(flagMiner)
			ip, _ := cmd.Flags().GetString(flagIP)
			port, _ := cmd.Flags().GetUint64(flagPort)
			fmt.Println("Launching TBB Node and its HTTP API...")

			bootstrap := node.NewPeerNode(
				"127.0.0.1",
				8080,
				true,
				database.NewAccount("andrej"),
				false,
			)

			n := node.New(
				getDataDirFromCmd(cmd),
				ip,
				port,
				database.NewAccount(miner),
				bootstrap,
			)

			err := n.Run(context.Background())
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(runCmd)
	runCmd.Flags().String(
		flagMiner,
		node.DefaultMiner,
		"miner account of this node to receive block rewards")
	runCmd.Flags().String(
		flagIP,
		"",
		"exposedIP for communication with peers",
	)
	runCmd.Flags().Uint64(
		flagPort,
		node.DefaultHTTPPort,
		"exposed HTTP port for communication with peers",
	)

	return runCmd
}

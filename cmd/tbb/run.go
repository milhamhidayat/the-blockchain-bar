package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"the-blockchain-bar/node"
)

func runCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launches the TBB node and its HTTP API",
		Run: func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetUint64(flagPort)
			fmt.Println("Launching TBB Node and its HTTP API...")

			bootstrap := node.NewPeerNode(
				"localhost",
				8081,
				true,
				true)

			n := node.New(
				getDataDirFromCmd(cmd),
				port,
				bootstrap)

			err := n.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(runCmd)
	runCmd.Flags().Uint64(flagPort,
		node.DefaultHTTPPort,
		"exposed HTTP port for communication with peers")

	return runCmd
}

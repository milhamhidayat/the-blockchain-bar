package main

import (
	"fmt"
	"os"

	"the-blockchain-bar/database"
	"the-blockchain-bar/node"

	"github.com/spf13/cobra"
)

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain database according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			pendingBlock := node.NewPendingBlock(
				database.Hash{},
				state.NextBlockNumber(),
				database.NewAccount("andrej"),
				[]database.Tx{
					database.NewTx("andrej", "andrej", 3, ""),
					database.NewTx("andrej", "babayaga", 2000, ""),
					database.NewTx("babayaga", "andrej", 1, ""),
					database.NewTx("babayaga", "caesar", 1000, ""),
					database.NewTx("babayaga", "andrej", 50, ""),
				},
			)
			_, err = node.Mine(cmd.Context(), pendingBlock)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)

	return migrateCmd
}

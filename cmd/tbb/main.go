package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"the-blockchain-bar/fs"
)

const (
	flagDataDir = "datadir"
	flagIP      = "ip"
	flagPort    = "port"
	flagMiner   = "miner"
)

func main() {
	var tbbCmd = &cobra.Command{
		Use:   "tbb",
		Short: "The Blockchain Bar CLI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ok")
		},
	}

	tbbCmd.AddCommand(versionCmd)
	tbbCmd.AddCommand(balancesCmd())
	tbbCmd.AddCommand(runCmd())
	tbbCmd.AddCommand(migrateCmd())

	err := tbbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagDataDir, "", "Absolute path to the node data dir where the DB will/is stored")
	cmd.MarkFlagRequired(flagDataDir)
}

func getDataDirFromCmd(cmd *cobra.Command) string {
	dataDir, _ := cmd.Flags().GetString(flagDataDir)
	return fs.ExpandPath(dataDir)
}

func incorrectUsageErr() error {
	return errors.New("incorrect usage")
}

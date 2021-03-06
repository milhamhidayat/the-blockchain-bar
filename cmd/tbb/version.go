package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	// Major is app major versioning
	Major = "0"
	// Minor is app minor versioning
	Minor = "9"
	// Fix is app fix versioning
	Fix = "0"
	// Verbal is app information
	Verbal = "PoW Reward"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Describes version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s.%s.%s-beta %s", Major, Minor, Fix, Verbal)
	},
}

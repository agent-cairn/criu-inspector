package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "criu-inspector",
	Short: "CRIU checkpoint inspector tool",
	Long: `A CLI tool to inspect CRIU (Checkpoint/Restore In Userspace) checkpoint
directories and extract structured information about processes, memory maps,
file descriptors, and network state. Designed for live migration debugging
and conference demonstrations.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("criu-inspector v0.1.0")
	},
}

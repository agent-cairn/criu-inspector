package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/agent-cairn/criu-inspector/pkg/models"
	"github.com/agent-cairn/criu-inspector/pkg/output"
	"github.com/agent-cairn/criu-inspector/pkg/parser"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	jsonOut bool
	pretty  bool
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <checkpoint-dir>",
	Short: "Parse and display CRIU checkpoint information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		checkpointDir := args[0]

		// Validate directory exists
		if _, err := os.Stat(checkpointDir); os.IsNotExist(err) {
			return fmt.Errorf("checkpoint directory not found: %s", checkpointDir)
		}

		// Parse checkpoint
		data, err := parser.ParseCheckpoint(checkpointDir, verbose)
		if err != nil {
			return fmt.Errorf("failed to parse checkpoint: %w", err)
		}

		// Output results
		if jsonOut {
			if err := printJSON(data, pretty); err != nil {
				return err
			}
		} else {
			output.PrintHumanReadable(data, verbose)
		}

		return nil
	},
}

func init() {
	inspectCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Include inventory and file descriptors")
	inspectCmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	inspectCmd.Flags().BoolVar(&pretty, "pretty", false, "Pretty-print JSON (requires --json)")
}

func printJSON(data *models.CheckpointData, pretty bool) error {
	var bytes []byte
	var err error

	if pretty {
		bytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		bytes, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(bytes))
	return nil
}

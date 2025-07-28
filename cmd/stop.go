package cmd

import (
	"fmt"
	"mindsdb-go-cli/internal/mindsdb"

	"github.com/spf13/cobra"
)

var removeContainer bool

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop embedded MindsDB instance",
	Long: `Stop the embedded MindsDB Docker container.
    
This command will gracefully stop the MindsDB container while preserving
any models and data for the next time you start it.

Examples:
  mindsdb-cli stop                    # Stop the container
  mindsdb-cli stop --remove           # Stop and remove the container`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🛑 Stopping embedded MindsDB instance...")

		// Check if Docker is available
		if !mindsdb.IsDockerAvailable() {
			fmt.Println("❌ Docker is not available.")
			fmt.Println("   Please install Docker Desktop or ensure Docker daemon is running.")
			return
		}

		// Create a MindsDB client for container management
		mindsdbClient := &mindsdb.MindsDBClient{
			EmbeddedMode: true,
		}

		// Get container status
		isRunning, startedAt, err := mindsdbClient.GetContainerStatus()
		if err != nil {
			fmt.Printf("❌ Failed to get container status: %v\n", err)
			return
		}

		if startedAt == "" {
			fmt.Println("ℹ️  No MindsDB container found")
			return
		}

		if !isRunning && !removeContainer {
			fmt.Println("ℹ️  MindsDB container is not running")
			return
		}

		// Stop the container (and optionally remove it)
		if err := mindsdbClient.StopEmbeddedMindsDB(removeContainer); err != nil {
			fmt.Printf("❌ Failed to stop container: %v\n", err)
			return
		}

		if !removeContainer {
			fmt.Println("💡 Use 'mindsdb-cli start' to start it again")
			fmt.Println("💡 Use 'mindsdb-cli stop --remove' to remove the container completely")
		}
	},
}

func init() {
	stopCmd.Flags().BoolVar(&removeContainer, "remove", false, "Remove the container after stopping")
}

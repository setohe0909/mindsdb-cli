package cmd

import (
	"fmt"
	"mindsdb-go-cli/internal/mindsdb"

	"github.com/spf13/cobra"
)

var startUser, startPass string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start embedded MindsDB instance",
	Long: `Start an embedded MindsDB instance using Docker.
    
This command will:
1. Check if Docker is available
2. Pull the MindsDB Docker image if needed
3. Start the MindsDB container
4. Wait for MindsDB to be ready

MindsDB by default doesn't require authentication unless specifically configured.

Examples:
  mindsdb-cli start                              # No authentication (MindsDB default)
  mindsdb-cli start --user admin --pass mypass  # Use custom credentials if auth is enabled`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Starting embedded MindsDB instance...")

		// Check if Docker is available
		if !mindsdb.IsDockerAvailable() {
			fmt.Println("❌ Docker is not available.")
			fmt.Println("   Please install Docker Desktop or ensure Docker daemon is running.")
			return
		}

		// MindsDB by default doesn't require authentication unless configured
		// Only show credentials if they are provided
		if startUser != "" || startPass != "" {
			fmt.Printf("📋 Using credentials: %s / %s\n", startUser, startPass)
		} else {
			fmt.Println("📋 Using MindsDB default (no authentication required)")
		}

		// Create embedded client (this will start the container)
		client, err := mindsdb.NewEmbeddedClient(startUser, startPass)
		if err != nil {
			fmt.Printf("❌ Failed to start embedded MindsDB: %v\n", err)
			return
		}
		defer client.Close()

		fmt.Println("✅ Embedded MindsDB started successfully!")
		fmt.Println("   - Web UI: http://localhost:47334")
		fmt.Println("   - Database: localhost:47335")
		fmt.Println("   - Container: mindsdb-cli-embedded")
		fmt.Println()
		fmt.Println("💡 Use 'mindsdb-cli status' to check the status")
		fmt.Println("💡 Use 'mindsdb-cli stop' to stop the instance")
		fmt.Println("💡 Use 'mindsdb-cli connect --embedded' to connect and run queries")
	},
}

func init() {
	startCmd.Flags().StringVarP(&startUser, "user", "u", "", "MindsDB username (optional, only needed if auth is enabled)")
	startCmd.Flags().StringVarP(&startPass, "pass", "p", "", "MindsDB password (optional, only needed if auth is enabled)")
}

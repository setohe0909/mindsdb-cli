package cmd

import (
	"fmt"
	"mindsdb-go-cli/internal/mindsdb"
	"os/exec"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check MindsDB instance status",
	Long: `Check the status of your embedded MindsDB instance.
    
This command shows:
- Whether Docker is available
- MindsDB container status
- Connection information if running

Example:
  mindsdb-cli status`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📊 MindsDB Status Check")
		fmt.Println("======================")

		// Check Docker availability
		fmt.Print("🐳 Docker: ")
		if !mindsdb.IsDockerAvailable() {
			fmt.Println("❌ Not available")
			fmt.Println("   Please install Docker Desktop or ensure Docker daemon is running.")
			fmt.Println("   You can still connect to external MindsDB instances using:")
			fmt.Println("   mindsdb-cli connect --host <host> --user <user> --pass <pass>")
			return
		}
		fmt.Println("✅ Available")

		// Get Docker version info
		if cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}"); cmd != nil {
			if output, err := cmd.Output(); err == nil {
				fmt.Printf("   Version: %s", string(output))
			}
		}

		// Check MindsDB container status
		fmt.Print("\n🧠 MindsDB Container: ")

		mindsdbClient := &mindsdb.MindsDBClient{
			EmbeddedMode: true,
		}

		isRunning, startedAt, err := mindsdbClient.GetContainerStatus()
		if err != nil {
			fmt.Printf("❌ Error checking status: %v\n", err)
			return
		}

		if startedAt == "" {
			fmt.Println("⚪ Not created")
			fmt.Println("   Use 'mindsdb-cli start' to create and start a container")
		} else if isRunning {
			fmt.Println("✅ Running")
			fmt.Println("   - Web UI: http://localhost:47334")
			fmt.Println("   - Database: localhost:47335")
			fmt.Println("   - Container: mindsdb-cli-embedded")
			fmt.Printf("   - Started: %s\n", startedAt)
		} else {
			fmt.Println("🛑 Stopped")
			fmt.Println("   Use 'mindsdb-cli start' to start the container")
		}

		// Show available commands
		fmt.Println("\n📋 Available Commands:")
		fmt.Println("   mindsdb-cli start --user <user> --pass <pass>  # Start embedded MindsDB")
		fmt.Println("   mindsdb-cli stop                               # Stop embedded MindsDB")
		fmt.Println("   mindsdb-cli connect --embedded                 # Connect to embedded instance")
		fmt.Println("   mindsdb-cli list-models                        # List ML models")
		fmt.Println("   mindsdb-cli query \"SELECT * FROM models\"       # Run SQL queries")
	},
}

func init() {
	// No flags needed for status command
}

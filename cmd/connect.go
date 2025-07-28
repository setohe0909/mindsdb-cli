package cmd

import (
	"fmt"
	"mindsdb-go-cli/internal/mindsdb"

	"github.com/spf13/cobra"
)

var host, user, pass string
var embedded bool

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a MindsDB instance",
	Long: `Connect to either an external MindsDB instance or an embedded one.
    
Examples:
  # Connect to external MindsDB
  mindsdb-cli connect --host localhost:47335 --user admin --pass admin
  
  # Connect to embedded MindsDB (requires Docker)
  mindsdb-cli connect --embedded --user admin --pass admin
  
  # Connect to MindsDB Cloud
  mindsdb-cli connect --host cloud.mindsdb.com --user your-email --pass your-password`,
	Run: func(cmd *cobra.Command, args []string) {
		var client *mindsdb.MindsDBClient
		var err error

		if embedded {
			fmt.Println("🔗 Connecting to embedded MindsDB instance...")

			// MindsDB by default doesn't require authentication unless configured
			// Only use provided credentials if they are explicitly given

			// Create embedded client
			client, err = mindsdb.NewEmbeddedClient(user, pass)
			if err != nil {
				fmt.Printf("❌ Failed to connect to embedded MindsDB: %v\n", err)
				fmt.Println("💡 Try 'mindsdb-cli start' first to ensure the container is running.")
				return
			}

			fmt.Printf("✅ Connected to embedded MindsDB!\n")
		} else {
			// Validate external connection parameters
			if host == "" || user == "" || pass == "" {
				fmt.Println("❌ Username, password, and host are required for external connections.")
				fmt.Println("   Use: mindsdb-cli connect --host <host> --user <username> --pass <password>")
				return
			}

			// Create external client
			client, err = mindsdb.NewClient(host, user, pass)
			if err != nil {
				fmt.Printf("❌ Failed to connect to MindsDB: %v\n", err)
				return
			}

			fmt.Printf("✅ Connected to MindsDB at %s!\n", host)
		}

		defer client.Close()

		// Test the connection with a simple query
		fmt.Println("\n🧪 Testing connection...")
		rows, err := client.Query("SELECT 1 as test")
		if err != nil {
			fmt.Printf("⚠️  Connection established but query failed: %v\n", err)
		} else {
			rows.Close() // Close the result set
			fmt.Println("✅ Connection test successful!")
		}

		fmt.Println("\n🚀 Ready to use MindsDB!")
		fmt.Println("💡 Try these commands:")
		fmt.Println("   mindsdb-cli list-models")
		fmt.Println("   mindsdb-cli query \"SELECT * FROM models\"")
		fmt.Println("   mindsdb-cli query \"SHOW DATABASES\"")
	},
}

func init() {
	connectCmd.Flags().StringVarP(&host, "host", "H", "", "MindsDB host (e.g. localhost:47335)")
	connectCmd.Flags().StringVarP(&user, "user", "u", "", "MindsDB username")
	connectCmd.Flags().StringVarP(&pass, "pass", "p", "", "MindsDB password")
	connectCmd.Flags().BoolVar(&embedded, "embedded", false, "Connect to embedded MindsDB instance")
}

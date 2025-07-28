package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "mindsdb-cli",
	Short: "MindsDB CLI",
	Long:  "Command line interface for MindsDB in Go with embedded MindsDB support.",
	Run: func(cmd *cobra.Command, args []string) {
		printBanner()
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(listModelsCmd)
	rootCmd.AddCommand(createModelCmd)
	rootCmd.AddCommand(queryCmd)
}

func printBanner() {
	logo := `
    ███╗   ███╗██╗███╗   ██╗██████╗ ███████╗██████╗ ██████╗
    ████╗ ████║██║████╗  ██║██╔══██╗██╔════╝██╔══██╗██╔══██╗
    ██╔████╔██║██║██╔██╗ ██║██║  ██║███████╗██║  ██║██████╔╝
    ██║╚██╔╝██║██║██║╚██╗██║██║  ██║╚════██║██║  ██║██╔══██╗
    ██║ ╚═╝ ██║██║██║ ╚████║██████╔╝███████║██████╔╝██████╔╝
    ╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═════╝ ╚══════╝╚═════╝ ╚═════╝
    `
	fmt.Println(logo)
	fmt.Printf("🧠  MindsDB CLI v%s\n", version)
	fmt.Println("-----------------------")
	fmt.Println("\nWelcome to the MindsDB Command Line Interface!")
	fmt.Println("Interact with your AI models directly from your terminal.\n")

	fmt.Println("📦 Embedded MindsDB Commands:")
	fmt.Println("  start          Start embedded MindsDB instance (Docker)")
	fmt.Println("  stop           Stop embedded MindsDB instance")
	fmt.Println("  status         Check MindsDB instance status")
	fmt.Println("")
	fmt.Println("🔗 Connection Commands:")
	fmt.Println("  connect        Connect to a MindsDB instance")
	fmt.Println("")
	fmt.Println("🤖 Model Management:")
	fmt.Println("  list-models    List available ML models")
	fmt.Println("  create-model   Create and train a new ML model")
	fmt.Println("  query          Execute SQL queries and predictions")
	fmt.Println("")
	fmt.Println("💡 Quick Start:")
	fmt.Println("  # Start embedded MindsDB (no separate installation needed!)")
	fmt.Println("  mindsdb-cli start --user admin --pass admin")
	fmt.Println("")
	fmt.Println("  # Connect to embedded instance")
	fmt.Println("  mindsdb-cli connect --embedded --user admin --pass admin")
	fmt.Println("")
	fmt.Println("  # Or connect to external MindsDB")
	fmt.Println("  mindsdb-cli connect --host localhost:47335 --user admin --pass admin")
	fmt.Println("")
	fmt.Println("Use 'mindsdb-cli <command> --help' for more information about a command.")
}

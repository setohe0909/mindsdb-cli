package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
    Use:   "mindsdb-cli",
    Short: "MindsDB CLI",
    Long:  "Command line interface for MindsDB in Go.",
    Run: func(cmd *cobra.Command, args []string) {
        printBanner()
    },
}

func Execute() {
    cobra.CheckErr(rootCmd.Execute())
}

func init() {
    rootCmd.AddCommand(connectCmd)
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
    fmt.Println("[✓] Connected to: MindsDB 25.7.3 (localhost)")
    fmt.Println("\nAvailable Commands:")
    fmt.Println("  connect        Connect to a MindsDB instance")
    fmt.Println("  list-models    List available models")
    fmt.Println("  create-model   Train a new model")
    fmt.Println("  query          Run prediction queries")
    fmt.Println("\nType \"mindsdb-cli --help\" for more info.")

    os.Exit(0)
}
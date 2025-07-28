package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

var listModelsCmd = &cobra.Command{
    Use:   "list-models",
    Short: "List all available models",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Listing all models...")
        // Lógica para listar modelos
    },
}
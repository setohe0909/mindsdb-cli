package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

var querySQL string

var queryCmd = &cobra.Command{
    Use:   "query",
    Short: "Execute a SQL query",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Executing query: %s\n", querySQL)
        // Lógica para ejecutar query
    },
}

func init() {
    queryCmd.Flags().StringVar(&querySQL, "sql", "", "SQL query to execute")
}
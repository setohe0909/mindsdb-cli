package cmd

import (
    "fmt"
    "mindsdb-go-cli/internal/mindsdb"

    "github.com/spf13/cobra"
)

var host, user, pass string

var connectCmd = &cobra.Command{
    Use:   "connect",
    Short: "Connect to a MindsDB instance",
    Run: func(cmd *cobra.Command, args []string) {
        client, err := mindsdb.NewClient(host, user, pass)
        if err != nil {
            fmt.Println("❌ Connection failed:", err)
            return
        }
        defer client.Close()

        version, err := client.QueryVersion()
        if err != nil {
            fmt.Println("⚠️ Failed to get MindsDB version:", err)
            return
        }

        fmt.Println("✅ Connected to MindsDB Version:", version)
    },
}

func init() {
    connectCmd.Flags().StringVar(&host, "host", "localhost", "MindsDB host")
    connectCmd.Flags().StringVar(&user, "user", "", "Username")
    connectCmd.Flags().StringVar(&pass, "pass", "", "Password")
}
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var modelName, fromTable, predictColumn string

var createModelCmd = &cobra.Command{
	Use:   "create-model",
	Short: "Create a new model",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Creating model '%s' from '%s' to predict '%s'\n", modelName, fromTable, predictColumn)
		// Lógica para crear modelo
	},
}

func init() {
	createModelCmd.Flags().StringVar(&modelName, "name", "", "Model name")
	createModelCmd.Flags().StringVar(&fromTable, "from", "", "Source table")
	createModelCmd.Flags().StringVar(&predictColumn, "predict", "", "Target column")
}

package cmd

import (
	"fmt"
	"mindsdb-go-cli/internal/mindsdb"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var querySQL string
var queryHost, queryUser, queryPass string
var queryEmbedded bool

var queryCmd = &cobra.Command{
	Use:   "query [SQL]",
	Short: "Execute a SQL query on MindsDB",
	Long: `Execute SQL queries on MindsDB instance.
    
You can provide the query as an argument or use the --sql flag.
The command will automatically connect to MindsDB and execute your query.

Examples:
  mindsdb-cli query "SHOW DATABASES"
  mindsdb-cli query "SELECT * FROM models"
  mindsdb-cli query --sql "DESCRIBE information_schema.models"
  mindsdb-cli query --embedded "SELECT name FROM models"
  mindsdb-cli query --host localhost:47335 --user admin --pass admin "SHOW TABLES"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get SQL query from args or flag
		var sql string
		if len(args) > 0 {
			sql = strings.Join(args, " ")
		} else if querySQL != "" {
			sql = querySQL
		} else {
			color.Red("❌ Please provide a SQL query")
			fmt.Println("   Use: mindsdb-cli query \"SHOW DATABASES\"")
			fmt.Println("   Or:  mindsdb-cli query --sql \"SELECT * FROM models\"")
			return
		}

		color.Cyan("🔍 Executing query: %s", sql)
		fmt.Println()

		// Connect to MindsDB
		var client *mindsdb.MindsDBClient
		var err error

		if queryEmbedded {
			color.Blue("🔗 Connecting to embedded MindsDB...")
			client, err = mindsdb.NewEmbeddedClient(queryUser, queryPass)
			if err != nil {
				color.Red("❌ Failed to connect to embedded MindsDB: %v", err)
				color.Yellow("💡 Try 'mindsdb-cli start' first to ensure the container is running.")
				return
			}
		} else if queryHost != "" {
			color.Blue("🔗 Connecting to MindsDB at %s...", queryHost)
			if queryUser == "" || queryPass == "" {
				color.Red("❌ Username and password are required for external connections.")
				fmt.Println("   Use: mindsdb-cli query --host <host> --user <user> --pass <pass> \"<SQL>\"")
				return
			}
			client, err = mindsdb.NewClient(queryHost, queryUser, queryPass)
			if err != nil {
				color.Red("❌ Failed to connect to MindsDB: %v", err)
				return
			}
		} else {
			// Default to embedded mode
			color.Blue("🔗 Connecting to embedded MindsDB (default)...")
			client, err = mindsdb.NewEmbeddedClient("", "")
			if err != nil {
				color.Red("❌ Failed to connect to embedded MindsDB: %v", err)
				color.Yellow("💡 Try 'mindsdb-cli start' first or use --host for external connections.")
				return
			}
		}
		defer client.Close()

		// Execute query
		if err := executeAndDisplayQuery(client, sql); err != nil {
			color.Red("❌ Query execution failed: %v", err)
			return
		}
	},
}

func executeAndDisplayQuery(client *mindsdb.MindsDBClient, sql string) error {
	rows, err := client.Query(sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	if len(columns) == 0 {
		color.Green("✅ Query executed successfully (no results returned)")
		return nil
	}

	// Collect all data first to calculate column widths
	var allRows [][]string

	for rows.Next() {
		valuePtrs := make([]interface{}, len(columns))
		for i := range valuePtrs {
			valuePtrs[i] = new(interface{})
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		row := make([]string, len(columns))
		for i, val := range valuePtrs {
			if val == nil || *(val.(*interface{})) == nil {
				row[i] = "NULL"
			} else {
				v := *(val.(*interface{}))
				switch v := v.(type) {
				case []byte:
					row[i] = string(v)
				case string:
					row[i] = v
				case nil:
					row[i] = "NULL"
				default:
					row[i] = fmt.Sprintf("%v", v)
				}
			}
		}
		allRows = append(allRows, row)
	}

	// Calculate column widths
	colWidths := make([]int, len(columns))
	for i, col := range columns {
		colWidths[i] = len(col)
	}

	for _, row := range allRows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print results header
	color.New(color.FgHiMagenta, color.Bold).Println("📊 Results:")
	fmt.Println()

	// Print top border
	printTableBorder(colWidths, "┌", "┬", "┐", "─")

	// Print headers
	fmt.Print("│ ")
	for i, col := range columns {
		headerColor := color.New(color.FgHiCyan, color.Bold)
		fmt.Print(headerColor.Sprintf("%-*s", colWidths[i], col))
		if i < len(columns)-1 {
			fmt.Print(" │ ")
		}
	}
	fmt.Println(" │")

	// Print header separator
	printTableBorder(colWidths, "├", "┼", "┤", "─")

	// Print data rows
	for _, row := range allRows {
		fmt.Print("│ ")
		for i, cell := range row {
			var cellColor *color.Color
			if cell == "NULL" {
				cellColor = color.New(color.FgHiBlack)
			} else {
				cellColor = color.New(color.FgWhite)
			}
			fmt.Print(cellColor.Sprintf("%-*s", colWidths[i], cell))
			if i < len(columns)-1 {
				fmt.Print(" │ ")
			}
		}
		fmt.Println(" │")
	}

	// Print bottom border
	printTableBorder(colWidths, "└", "┴", "┘", "─")

	// Print summary
	fmt.Println()
	if len(allRows) == 0 {
		color.Yellow("📝 No rows returned")
	} else if len(allRows) == 1 {
		color.Green("✅ Query completed successfully (%d row)", len(allRows))
	} else {
		color.Green("✅ Query completed successfully (%d rows)", len(allRows))
	}

	return rows.Err()
}

func printTableBorder(colWidths []int, left, middle, right, fill string) {
	fmt.Print(left)
	for i, width := range colWidths {
		fmt.Print(strings.Repeat(fill, width+2))
		if i < len(colWidths)-1 {
			fmt.Print(middle)
		}
	}
	fmt.Println(right)
}

func init() {
	queryCmd.Flags().StringVar(&querySQL, "sql", "", "SQL query to execute")
	queryCmd.Flags().StringVar(&queryHost, "host", "", "MindsDB host (e.g., localhost:47335)")
	queryCmd.Flags().StringVar(&queryUser, "user", "", "MindsDB username")
	queryCmd.Flags().StringVar(&queryPass, "pass", "", "MindsDB password")
	queryCmd.Flags().BoolVar(&queryEmbedded, "embedded", false, "Use embedded MindsDB instance")
}

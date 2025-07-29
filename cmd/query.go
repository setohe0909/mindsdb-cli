package cmd

import (
	"bufio"
	"fmt"
	"mindsdb-go-cli/internal/mindsdb"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var querySQL string
var queryHost, queryUser, queryPass string
var queryEmbedded bool
var queryFormat string
var queryMaxWidth int
var queryCompact bool
var queryVertical bool
var queryLimit int
var queryForceTable bool

var queryCmd = &cobra.Command{
	Use:   "query [SQL]",
	Short: "Execute a SQL query on MindsDB or start interactive mode",
	Long: `Execute SQL queries on MindsDB instance.
    
You can provide the query as an argument, use the --sql flag, or start interactive mode.
When no query is provided, an interactive SQL prompt will start.

Examples:
  mindsdb-cli query                                                    # Start interactive mode
  mindsdb-cli query "SHOW DATABASES"
  mindsdb-cli query "SELECT * FROM models"
  mindsdb-cli query --sql "DESCRIBE information_schema.models"
  mindsdb-cli query --embedded "SELECT name FROM models"
  mindsdb-cli query --host localhost:47335 --user admin --pass admin "SHOW TABLES"
  mindsdb-cli query --format json "SELECT * FROM models"
  mindsdb-cli query --max-width 40 "SELECT * FROM training_data"
  mindsdb-cli query --compact "SELECT * FROM large_table"
  mindsdb-cli query --vertical "SELECT * FROM models"     # Force vertical layout
  mindsdb-cli query --limit 3 "SELECT * FROM big_table"   # Limit rows displayed
  mindsdb-cli query --force-table "SHOW HANDLERS"         # Force table format even if wide`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get SQL query from args or flag
		var sql string
		if len(args) > 0 {
			sql = strings.Join(args, " ")
		} else if querySQL != "" {
			sql = querySQL
		} else {
			// Start interactive mode
			startInteractiveMode()
			return
		}

		color.Cyan("🔍 Executing query: %s", sql)
		fmt.Println()

		// Connect to MindsDB
		client, err := connectToMindsDB()
		if err != nil {
			color.Red("❌ Connection failed: %v", err)
			return
		}
		defer client.Close()

		// Execute single query
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

	// Collect all data first
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

	// Display results based on format
	switch queryFormat {
	case "json":
		return displayAsJSON(columns, allRows)
	case "csv":
		return displayAsCSV(columns, allRows)
	default:
		return displayAsTable(columns, allRows)
	}
}

func displayAsJSON(columns []string, rows [][]string) error {
	color.New(color.FgHiMagenta, color.Bold).Println("📊 Results (JSON):")
	fmt.Println()

	fmt.Println("[")
	for i, row := range rows {
		fmt.Print("  {")
		for j, col := range columns {
			fmt.Printf(`"%s": "%s"`, col, strings.ReplaceAll(row[j], `"`, `\"`))
			if j < len(columns)-1 {
				fmt.Print(", ")
			}
		}
		fmt.Print("}")
		if i < len(rows)-1 {
			fmt.Print(",")
		}
		fmt.Println()
	}
	fmt.Println("]")

	color.Green("✅ Query completed successfully (%d rows)", len(rows))
	return nil
}

func displayAsCSV(columns []string, rows [][]string) error {
	color.New(color.FgHiMagenta, color.Bold).Println("📊 Results (CSV):")
	fmt.Println()

	// Header
	for i, col := range columns {
		fmt.Printf(`"%s"`, strings.ReplaceAll(col, `"`, `""`))
		if i < len(columns)-1 {
			fmt.Print(",")
		}
	}
	fmt.Println()

	// Rows
	for _, row := range rows {
		for i, cell := range row {
			fmt.Printf(`"%s"`, strings.ReplaceAll(cell, `"`, `""`))
			if i < len(row)-1 {
				fmt.Print(",")
			}
		}
		fmt.Println()
	}

	color.Green("✅ Query completed successfully (%d rows)", len(rows))
	return nil
}

func displayAsTable(columns []string, rows [][]string) error {
	// Apply row limit if specified
	if queryLimit > 0 && len(rows) > queryLimit {
		rows = rows[:queryLimit]
		color.Yellow("💡 Showing first %d rows (use --limit 0 to show all)", queryLimit)
		fmt.Println()
	}

	// Get terminal width for adaptive sizing
	termWidth := getTerminalWidth()

	// Force vertical layout if requested
	if queryVertical {
		return displayWideTable(columns, rows, termWidth)
	}

	// Skip all smart detection if force-table is enabled
	if !queryForceTable {
		// Smart detection for wide tables (much less aggressive now)
		shouldUseVertical := shouldUseVerticalLayout(columns, rows, termWidth)
		if shouldUseVertical {
			return displayWideTable(columns, rows, termWidth)
		}
	}

	// Continue with regular table formatting
	// Calculate available width for content (subtract borders and padding)
	availableWidth := termWidth - (len(columns) * 3) - 1
	if availableWidth < 20 {
		availableWidth = 80 // fallback for very narrow terminals
	}

	// Calculate optimal column widths
	colWidths := calculateColumnWidths(columns, rows, availableWidth)

	// Print results header
	color.New(color.FgHiMagenta, color.Bold).Println("📊 Results:")
	fmt.Println()

	// Print table (handles text wrapping internally)
	printTable(columns, rows, colWidths)

	// Print summary
	fmt.Println()
	if len(rows) == 0 {
		color.Yellow("📝 No rows returned")
	} else if len(rows) == 1 {
		color.Green("✅ Query completed successfully (%d row)", len(rows))
	} else {
		color.Green("✅ Query completed successfully (%d rows)", len(rows))
	}

	return nil
}

func shouldUseVerticalLayout(columns []string, rows [][]string, termWidth int) bool {
	// Only use vertical for extremely wide tables (15+ columns)
	if len(columns) >= 15 {
		return true
	}

	// Check if any single cell has extremely long content (over 200 chars)
	for _, row := range rows {
		for _, cell := range row {
			if len(cell) > 200 {
				return true
			}
		}
	}

	// Check total estimated width - only if it's REALLY too wide
	estimatedWidth := 0
	for i, col := range columns {
		colWidth := len(col)
		// Check content in this column
		for _, row := range rows {
			if i < len(row) && len(row[i]) > colWidth {
				colWidth = len(row[i])
			}
		}
		// Cap at reasonable width for estimation
		if colWidth > 40 {
			colWidth = 40
		}
		estimatedWidth += colWidth + 3 // +3 for borders and padding
	}

	// Only use vertical if estimated width is MUCH larger than terminal
	if estimatedWidth > termWidth+50 {
		return true
	}

	return false
}

func displayWideTable(columns []string, rows [][]string, termWidth int) error {
	// Print results header
	color.New(color.FgHiMagenta, color.Bold).Println("📊 Results:")
	color.Yellow("💡 Wide table detected (%d columns) - using vertical layout for better readability", len(columns))
	fmt.Println()

	if len(rows) == 0 {
		color.Yellow("📝 No rows returned")
		return nil
	}

	// Display each row vertically
	for rowIndex, row := range rows {
		// Row header
		color.New(color.FgHiCyan, color.Bold).Printf("📋 Row %d:\n", rowIndex+1)
		fmt.Println(strings.Repeat("─", min(termWidth-1, 60)))

		// Display each column-value pair
		maxColNameLen := 0
		for _, col := range columns {
			if len(col) > maxColNameLen {
				maxColNameLen = len(col)
			}
		}

		for i, col := range columns {
			value := ""
			if i < len(row) {
				value = row[i]
			}

			// Truncate very long values for vertical display
			if len(value) > termWidth-maxColNameLen-10 {
				value = value[:termWidth-maxColNameLen-13] + "..."
			}

			// Color the column name
			colColor := color.New(color.FgHiBlue, color.Bold)
			valueColor := color.New(color.FgWhite)
			if value == "NULL" || value == "" {
				valueColor = color.New(color.FgHiBlack)
				if value == "" {
					value = "NULL"
				}
			}

			fmt.Printf("  %s: %s\n",
				colColor.Sprintf("%-*s", maxColNameLen, col),
				valueColor.Sprint(value))
		}

		if rowIndex < len(rows)-1 {
			fmt.Println()
		}
	}

	// Print summary
	fmt.Println()
	fmt.Println(strings.Repeat("─", min(termWidth-1, 60)))
	if len(rows) == 1 {
		color.Green("✅ Query completed successfully (%d row)", len(rows))
	} else {
		color.Green("✅ Query completed successfully (%d rows)", len(rows))
	}

	// Suggest alternatives for large datasets
	if len(rows) > 5 {
		fmt.Println()
		color.Yellow("💡 For large datasets, try:")
		color.White("   --format json    # JSON format for full data")
		color.White("   --format csv     # CSV format for exports")
		color.White("   LIMIT 5          # Add LIMIT to your query")
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getTerminalWidth() int {
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		return width
	}
	return 120 // default fallback
}

func calculateColumnWidths(columns []string, rows [][]string, availableWidth int) []int {
	colWidths := make([]int, len(columns))

	// Start with header widths
	for i, col := range columns {
		colWidths[i] = len(col)
	}

	// Consider content widths
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				cellLen := len(cell)
				if cellLen > colWidths[i] {
					colWidths[i] = cellLen
				}
			}
		}
	}

	// Apply maximum width constraints - much more aggressive
	maxWidth := queryMaxWidth
	if maxWidth <= 0 {
		maxWidth = 20 // reduced for better alignment
	}

	// Apply compact mode if requested
	if queryCompact {
		maxWidth = 15 // very compact mode
	}

	// For very wide content, be even more aggressive
	totalContentWidth := 0
	for _, row := range rows {
		for _, cell := range row {
			totalContentWidth += len(cell)
		}
	}
	avgContentLength := 0
	if len(rows) > 0 && len(columns) > 0 {
		avgContentLength = totalContentWidth / (len(rows) * len(columns))
	}

	// If content is very long on average, use compact mode automatically
	if avgContentLength > 50 {
		maxWidth = 15 // very compact for long content
	}

	// Calculate total required width
	totalWidth := 0
	for _, width := range colWidths {
		if width > maxWidth {
			totalWidth += maxWidth
		} else {
			totalWidth += width
		}
	}
	totalWidth += (len(columns)-1)*3 + 4 // borders and separators

	// If still too wide, distribute space proportionally
	if totalWidth > availableWidth {
		factor := float64(availableWidth-(len(columns)-1)*3-4) / float64(totalWidth-(len(columns)-1)*3-4)
		for i := range colWidths {
			newWidth := int(float64(colWidths[i]) * factor)
			if newWidth < 4 { // minimum readable width
				newWidth = 4
			}
			if newWidth > maxWidth {
				newWidth = maxWidth
			}
			colWidths[i] = newWidth
		}
	} else {
		// Apply max width constraints
		for i := range colWidths {
			if colWidths[i] > maxWidth {
				colWidths[i] = maxWidth
			}
		}
	}

	return colWidths
}

func wrapRowsContent(rows [][]string, colWidths []int) [][]string {
	wrappedRows := make([][]string, 0)

	for _, row := range rows {
		wrappedRow := make([]string, len(row))
		for i, cell := range row {
			if i < len(colWidths) {
				wrappedRow[i] = truncateOrWrapText(cell, colWidths[i])
			} else {
				wrappedRow[i] = cell
			}
		}
		wrappedRows = append(wrappedRows, wrappedRow)
	}

	return wrappedRows
}

func truncateOrWrapText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	// For very short widths, just truncate with ellipsis
	if maxWidth <= 8 {
		if len(text) > maxWidth-3 {
			return text[:maxWidth-3] + "..."
		}
		return text
	}

	// For long content, be more aggressive about truncation
	if len(text) > 200 {
		// Show first meaningful part + ellipsis
		words := strings.Fields(text)
		if len(words) > 0 {
			result := ""
			for _, word := range words {
				if len(result)+len(word)+4 <= maxWidth { // +4 for " ..."
					if result == "" {
						result = word
					} else {
						result += " " + word
					}
				} else {
					break
				}
			}
			if result != "" && len(result) < len(text) {
				return result + "..."
			}
		}
	}

	// For medium content, try to wrap at word boundaries
	if strings.Contains(text, " ") && maxWidth > 15 {
		wrapped := wrapText(text, maxWidth)
		lines := strings.Split(wrapped, "\n")
		if len(lines) > 1 {
			return lines[0] + "..." // Show first line with ellipsis
		}
	}

	// Default truncation with ellipsis
	if maxWidth > 3 {
		return text[:maxWidth-3] + "..."
	}
	return text[:maxWidth]
}

func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= width {
			currentLine.WriteString(" " + word)
		} else {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

func printTable(columns []string, rows [][]string, colWidths []int) {
	// Print top border
	printTableBorder(colWidths, "┌", "┬", "┐", "─")

	// Print headers
	fmt.Print("│ ")
	for i, col := range columns {
		headerColor := color.New(color.FgHiCyan, color.Bold)
		// Ensure exact width by truncating or padding
		displayText := col
		if len(displayText) > colWidths[i] {
			displayText = displayText[:colWidths[i]]
		}
		fmt.Print(headerColor.Sprintf("%-*s", colWidths[i], displayText))
		if i < len(columns)-1 {
			fmt.Print(" │ ")
		}
	}
	fmt.Println(" │")

	// Print header separator
	printTableBorder(colWidths, "├", "┼", "┤", "─")

	// Print data rows
	for _, row := range rows {
		fmt.Print("│ ")
		for i := range columns {
			var cell string
			if i < len(row) {
				cell = row[i]
			}

			// Ensure exact width by truncating or padding
			if len(cell) > colWidths[i] {
				if colWidths[i] > 3 {
					cell = cell[:colWidths[i]-3] + "..."
				} else {
					cell = cell[:colWidths[i]]
				}
			}

			var cellColor *color.Color
			if cell == "NULL" || cell == "" {
				cellColor = color.New(color.FgHiBlack)
				if cell == "" {
					cell = "NULL"
				}
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

func connectToMindsDB() (*mindsdb.MindsDBClient, error) {
	var client *mindsdb.MindsDBClient
	var err error

	if queryEmbedded {
		color.Blue("🔗 Connecting to embedded MindsDB...")
		client, err = mindsdb.NewEmbeddedClient(queryUser, queryPass)
		if err != nil {
			color.Red("❌ Failed to connect to embedded MindsDB: %v", err)
			color.Yellow("💡 Try 'mindsdb-cli start' first to ensure the container is running.")
			return nil, err
		}
	} else if queryHost != "" {
		color.Blue("🔗 Connecting to MindsDB at %s...", queryHost)
		if queryUser == "" || queryPass == "" {
			color.Red("❌ Username and password are required for external connections.")
			fmt.Println("   Use: mindsdb-cli query --host <host> --user <user> --pass <pass>")
			return nil, fmt.Errorf("missing credentials")
		}
		client, err = mindsdb.NewClient(queryHost, queryUser, queryPass)
		if err != nil {
			return nil, err
		}
	} else {
		// Default to embedded mode
		color.Blue("🔗 Connecting to embedded MindsDB (default)...")
		client, err = mindsdb.NewEmbeddedClient("", "")
		if err != nil {
			color.Red("❌ Failed to connect to embedded MindsDB: %v", err)
			color.Yellow("💡 Try 'mindsdb-cli start' first or use --host for external connections.")
			return nil, err
		}
	}

	return client, nil
}

func startInteractiveMode() {
	// Print welcome message
	color.New(color.FgHiCyan, color.Bold).Println("🧠 MindsDB Interactive SQL Mode")
	fmt.Println("================================")
	fmt.Println()
	color.Yellow("💡 Type SQL queries and press Enter to execute")
	color.Yellow("💡 Use semicolon (;) for multi-line queries")
	color.Yellow("💡 Commands: .help, .exit, .format, .compact, .vertical, .limit")
	fmt.Println()

	// Connect to MindsDB
	client, err := connectToMindsDB()
	if err != nil {
		return
	}
	defer client.Close()

	color.Green("✅ Connected! Ready for queries.")
	fmt.Println()

	// Start interactive loop
	scanner := bufio.NewScanner(os.Stdin)
	var multiLineQuery strings.Builder
	inMultiLine := false

	for {
		// Show prompt
		var prompt string
		if inMultiLine {
			prompt = color.New(color.FgHiBlack).Sprint("  ... ")
		} else {
			prompt = color.New(color.FgHiMagenta, color.Bold).Sprint("mindsdb> ")
		}
		fmt.Print(prompt)

		// Read input
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())

		// Handle empty lines
		if line == "" {
			if inMultiLine && multiLineQuery.Len() > 0 {
				// Execute multi-line query
				sql := strings.TrimSpace(multiLineQuery.String())
				executeInteractiveQuery(client, sql)
				multiLineQuery.Reset()
				inMultiLine = false
			}
			continue
		}

		// Handle special commands
		if strings.HasPrefix(line, ".") {
			if handleSpecialCommand(line) {
				break // Exit requested
			}
			continue
		}

		// Handle SQL queries
		if inMultiLine {
			multiLineQuery.WriteString(" ")
			multiLineQuery.WriteString(line)
		} else {
			multiLineQuery.WriteString(line)
		}

		// Check if query ends with semicolon
		if strings.HasSuffix(line, ";") {
			// Execute complete query
			sql := strings.TrimSpace(strings.TrimSuffix(multiLineQuery.String(), ";"))
			if sql != "" {
				executeInteractiveQuery(client, sql)
			}
			multiLineQuery.Reset()
			inMultiLine = false
		} else {
			// Continue multi-line
			inMultiLine = true
		}
	}

	color.New(color.FgHiGreen).Println("\n👋 Goodbye!")
}

func handleSpecialCommand(command string) bool {
	switch {
	case command == ".exit" || command == ".quit":
		return true

	case command == ".help":
		fmt.Println()
		color.New(color.FgHiCyan, color.Bold).Println("📚 MindsDB Interactive Commands:")
		fmt.Println()
		color.White("  .help                    Show this help message")
		color.White("  .exit, .quit             Exit interactive mode")
		color.White("  .format <table|json|csv> Change output format")
		color.White("  .compact                 Toggle compact table mode")
		color.White("  .vertical                Toggle vertical layout for wide tables")
		color.White("  .limit <number>          Set row limit (0 for no limit)")
		color.White("  .clear                   Clear screen")
		fmt.Println()
		color.Yellow("💡 SQL Tips:")
		color.White("  - End queries with semicolon (;) for execution")
		color.White("  - Press Enter on empty line to execute multi-line query")
		color.White("  - Use SHOW DATABASES, SHOW TABLES for exploration")
		color.White("  - Wide tables (8+ columns) automatically use vertical layout")
		fmt.Println()

	case strings.HasPrefix(command, ".format "):
		newFormat := strings.TrimSpace(strings.TrimPrefix(command, ".format "))
		if newFormat == "table" || newFormat == "json" || newFormat == "csv" {
			queryFormat = newFormat
			color.Green("✅ Output format changed to: %s", newFormat)
		} else {
			color.Red("❌ Invalid format. Use: table, json, or csv")
		}

	case command == ".compact":
		queryCompact = !queryCompact
		if queryCompact {
			color.Green("✅ Compact mode enabled")
		} else {
			color.Green("✅ Compact mode disabled")
		}

	case command == ".vertical":
		queryVertical = !queryVertical
		if queryVertical {
			color.Green("✅ Vertical layout enabled")
		} else {
			color.Green("✅ Vertical layout disabled")
		}

	case strings.HasPrefix(command, ".limit "):
		limitStr := strings.TrimSpace(strings.TrimPrefix(command, ".limit "))
		if limitStr == "" {
			color.Yellow("Current row limit: %d (0 = no limit)", queryLimit)
		} else {
			if newLimit, err := strconv.Atoi(limitStr); err == nil {
				queryLimit = newLimit
				if queryLimit == 0 {
					color.Green("✅ Row limit disabled (showing all rows)")
				} else {
					color.Green("✅ Row limit set to: %d", queryLimit)
				}
			} else {
				color.Red("❌ Invalid number. Use: .limit <number> or .limit 0 for no limit")
			}
		}

	case command == ".clear":
		// Clear screen
		fmt.Print("\033[2J\033[H")

	default:
		color.Red("❌ Unknown command: %s", command)
		color.Yellow("💡 Type .help for available commands")
	}

	return false
}

func executeInteractiveQuery(client *mindsdb.MindsDBClient, sql string) {
	fmt.Println()
	color.Cyan("🔍 Executing: %s", sql)
	fmt.Println()

	if err := executeAndDisplayQuery(client, sql); err != nil {
		color.Red("❌ Error: %v", err)
	}
	fmt.Println()
}

func init() {
	queryCmd.Flags().StringVar(&querySQL, "sql", "", "SQL query to execute")
	queryCmd.Flags().StringVar(&queryHost, "host", "", "MindsDB host (e.g., localhost:47335)")
	queryCmd.Flags().StringVar(&queryUser, "user", "", "MindsDB username")
	queryCmd.Flags().StringVar(&queryPass, "pass", "", "MindsDB password")
	queryCmd.Flags().BoolVar(&queryEmbedded, "embedded", false, "Use embedded MindsDB instance")
	queryCmd.Flags().StringVar(&queryFormat, "format", "table", "Output format: table, json, csv")
	queryCmd.Flags().IntVar(&queryMaxWidth, "max-width", 20, "Maximum column width for table display")
	queryCmd.Flags().BoolVar(&queryCompact, "compact", false, "Use compact mode for table display")
	queryCmd.Flags().BoolVar(&queryVertical, "vertical", false, "Force vertical layout for wide tables")
	queryCmd.Flags().IntVar(&queryLimit, "limit", 0, "Limit the number of rows displayed")
	queryCmd.Flags().BoolVar(&queryForceTable, "force-table", false, "Force traditional table format even for wide tables")
}

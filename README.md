# MindsDB CLI

A command-line interface for MindsDB written in Go. This tool allows you to interact with MindsDB instances directly from your terminal, enabling you to connect, manage models, and run predictions with ease.

## 🚀 Features

- **Easy Connection**: Connect to MindsDB instances using PostgreSQL protocol
- **Model Management**: Create and list machine learning models
- **Query Execution**: Run SQL queries and predictions
- **Beautiful CLI**: Clean interface with helpful banners and status messages
- **Cross-platform**: Works on macOS, Linux, and Windows

## 📦 Installation

### Prerequisites

- Go 1.20 or higher
- Access to a MindsDB instance

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd mindsdb-cli

# Build the application
go build -o mindsdb-cli

# Make it executable (on Unix-like systems)
chmod +x mindsdb-cli

# Optionally, move to PATH
sudo mv mindsdb-cli /usr/local/bin/
```

## 🎯 Usage

### Getting Started

Run the CLI without any arguments to see the welcome banner:

```bash
mindsdb-cli
```

This will display:
- MindsDB ASCII logo
- Version information
- Available commands
- Connection status

### Available Commands

#### 1. Connect to MindsDB

Connect to a MindsDB instance:

```bash
# Connect to localhost (default)
mindsdb-cli connect --user your_username --pass your_password

# Connect to a remote instance
mindsdb-cli connect --host remote-mindsdb.com --user username --pass password

# Connect with default host (localhost)
mindsdb-cli connect --user admin --pass admin
```

**Flags:**
- `--host`: MindsDB host (default: "localhost")
- `--user`: Username for authentication
- `--pass`: Password for authentication

#### 2. List Models

View all available models in your MindsDB instance:

```bash
mindsdb-cli list-models
```

#### 3. Create a Model

Train a new machine learning model:

```bash
mindsdb-cli create-model --name my_model --from source_table --predict target_column
```

**Flags:**
- `--name`: Name for the new model
- `--from`: Source table containing training data
- `--predict`: Target column to predict

**Example:**
```bash
mindsdb-cli create-model --name house_price_predictor --from real_estate_data --predict price
```

#### 4. Execute Queries

Run SQL queries and predictions:

```bash
mindsdb-cli query --sql "SELECT * FROM mindsdb.models"

# Make predictions
mindsdb-cli query --sql "SELECT price FROM house_price_predictor WHERE bedrooms=3 AND bathrooms=2"
```

**Flags:**
- `--sql`: SQL query to execute

### Help and Documentation

Get help for any command:

```bash
# General help
mindsdb-cli --help

# Command-specific help
mindsdb-cli connect --help
mindsdb-cli create-model --help
```

## 🏗️ Project Structure

```
mindsdb-cli/
├── main.go                 # Application entry point
├── go.mod                  # Go module dependencies
├── cmd/                    # CLI commands (Cobra-based)
│   ├── root.go            # Root command and CLI setup
│   ├── connect.go         # Connection command
│   ├── create_model.go    # Model creation command
│   ├── list_models.go     # Model listing command
│   └── query.go           # Query execution command
├── internal/              # Internal packages
│   └── mindsdb/
│       └── client.go      # MindsDB client implementation
├── LICENSE                # Project license
└── README.md             # This file
```

### Architecture Overview

#### Core Components

1. **Main Entry Point** (`main.go`)
   - Simple entry point that delegates to the Cobra command system

2. **CLI Commands** (`cmd/`)
   - **Root Command** (`root.go`): Main CLI setup, banner display, and command registration
   - **Connect** (`connect.go`): Handles connection to MindsDB instances
   - **Create Model** (`create_model.go`): Manages model creation workflow
   - **List Models** (`list_models.go`): Lists available models
   - **Query** (`query.go`): Executes SQL queries and predictions

3. **MindsDB Client** (`internal/mindsdb/client.go`)
   - PostgreSQL-based client for communicating with MindsDB
   - Handles connection management and query execution
   - Provides version checking capabilities

#### Dependencies

- **[Cobra](https://github.com/spf13/cobra)**: Modern CLI framework for Go
- **[pgx](https://github.com/jackc/pgx)**: PostgreSQL driver for Go (MindsDB uses PostgreSQL wire protocol)

#### Design Patterns

- **Command Pattern**: Each CLI command is implemented as a separate Cobra command
- **Client Pattern**: MindsDB client abstracts connection and communication logic
- **Flag-based Configuration**: Commands use flags for parameter input

## 🛠️ Development

### Prerequisites for Development

- Go 1.20 or higher
- Git
- A running MindsDB instance for testing

### Setting Up Development Environment

```bash
# Clone the repository
git clone <repository-url>
cd mindsdb-cli

# Install dependencies
go mod tidy

# Run the application
go run main.go

# Run tests (when available)
go test ./...

# Build for development
go build -o mindsdb-cli-dev
```

### Code Organization

- **Commands**: Add new commands in the `cmd/` directory following the existing pattern
- **Client Logic**: Extend the MindsDB client in `internal/mindsdb/client.go`
- **Utilities**: Add shared utilities in appropriate internal packages

### Adding New Commands

1. Create a new file in `cmd/` (e.g., `cmd/new_command.go`)
2. Define the command using Cobra conventions
3. Register the command in `cmd/root.go` init function
4. Add any necessary flags and validation

Example:
```go
var newCmd = &cobra.Command{
    Use:   "new-command",
    Short: "Description of the new command",
    Run: func(cmd *cobra.Command, args []string) {
        // Command implementation
    },
}

func init() {
    // Add flags if needed
    newCmd.Flags().StringVar(&flagVar, "flag", "default", "Flag description")
}
```
 
## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests if applicable
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## 📝 License

This project is licensed under the terms specified in the LICENSE file.

## 🔗 Related Resources

- [MindsDB Documentation](https://docs.mindsdb.com/)
- [MindsDB GitHub Repository](https://github.com/mindsdb/mindsdb)
- [Cobra CLI Framework](https://cobra.dev/)


---

**Version**: 0.1.0  
**Go Version**: 1.20+
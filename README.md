# MindsDB CLI

A command-line interface for MindsDB written in Go. This tool allows you to interact with MindsDB instances directly from your terminal, enabling you to connect, manage models, and run predictions with ease.

## 🚀 Features

- **Easy Connection**: Connect to MindsDB instances using PostgreSQL or MySQL protocols
- **Model Management**: Create and list machine learning models
- **Query Execution**: Run SQL queries and predictions
- **Beautiful CLI**: Clean interface with helpful banners and status messages
- **Cross-platform**: Works on macOS, Linux, and Windows
- **✅ Embedded MindsDB**: No separate installation required! Run MindsDB directly from the CLI using Docker

## 📦 Installation

### Prerequisites

- Go 1.23 or higher
- Docker (for embedded MindsDB support)
- *Optional*: Access to an external MindsDB instance (local or cloud)

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
- Getting started instructions

## 🔌 Connecting to MindsDB

### Option 1: Embedded MindsDB (Recommended) ✅

Use MindsDB without any separate installation - everything runs in Docker:

```bash
# Start embedded MindsDB (automatically downloads and starts MindsDB)
mindsdb-cli start --user admin --pass admin

# Connect to embedded instance
mindsdb-cli connect --embedded --user admin --pass admin

# Check status
mindsdb-cli status

# Stop when done
mindsdb-cli stop
```

### Option 2: Connect to Existing MindsDB Instance

If you have MindsDB already running (locally or remotely):

```bash
# Connect to localhost (default MindsDB installation)
mindsdb-cli connect --host localhost:47335 --user mindsdb --pass ""

# Connect to MindsDB Cloud
mindsdb-cli connect --host cloud.mindsdb.com --user your_email --pass your_password

# Connect to a custom MindsDB instance
mindsdb-cli connect --host your-host:port --user username --pass password
```

### Option 3: Install MindsDB Locally (Traditional Way)

If you prefer to install MindsDB separately:

```bash
# Using pip
pip install mindsdb

# Start MindsDB
python -m mindsdb

# Then connect from another terminal
mindsdb-cli connect --host localhost:47335 --user mindsdb --pass ""
```

## 📋 Available Commands

### Embedded MindsDB Commands ✅

#### 1. Start Embedded MindsDB

Start MindsDB in a Docker container:

```bash
mindsdb-cli start --user admin --pass admin
```

**Flags:**
- `--user`: Username for MindsDB (optional, only needed if auth is enabled)
- `--pass`: Password for MindsDB (optional, only needed if auth is enabled)

**What it does:**
1. Checks if Docker is available
2. Pulls the MindsDB Docker image if needed
3. Starts the MindsDB container
4. Waits for MindsDB to be ready

#### 2. Stop Embedded MindsDB

Stop the MindsDB container:

```bash
mindsdb-cli stop                    # Stop the container
mindsdb-cli stop --remove           # Stop and remove the container
```

**Flags:**
- `--remove`: Remove the container after stopping

#### 3. Check Status

Check the status of your embedded MindsDB instance:

```bash
mindsdb-cli status
```

**Shows:**
- Docker availability
- MindsDB container status
- Connection information if running
- Available commands

### Connection Commands

#### 4. Connect to MindsDB

Connect to a MindsDB instance (embedded or external):

```bash
# Connect to embedded instance
mindsdb-cli connect --embedded --user admin --pass admin

# Connect to external instance
mindsdb-cli connect --host localhost:47335 --user mindsdb --pass ""
```

**Flags:**
- `--host`: MindsDB host and port (e.g., "localhost:47335")
- `--user`: Username for authentication
- `--pass`: Password for authentication
- `--embedded`: Connect to embedded MindsDB instance

### Model Management Commands

#### 5. List Models

View all available models in your MindsDB instance:

```bash
mindsdb-cli list-models
```

#### 6. Create a Model

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

#### 7. Execute Queries

Run SQL queries and predictions:

```bash
mindsdb-cli query --sql "SELECT * FROM mindsdb.models"

# Make predictions
mindsdb-cli query --sql "SELECT price FROM house_price_predictor WHERE bedrooms=3 AND bathrooms=2"

# Use with embedded instance
mindsdb-cli query --embedded "SELECT name FROM models"

# Use with external instance
mindsdb-cli query --host localhost:47335 --user admin --pass admin "SHOW TABLES"
```

**Flags:**
- `--sql`: SQL query to execute
- `--embedded`: Use embedded MindsDB instance
- `--host`, `--user`, `--pass`: External MindsDB connection details

### Help and Documentation

Get help for any command:

```bash
# General help
mindsdb-cli --help

# Command-specific help
mindsdb-cli connect --help
mindsdb-cli start --help
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
│   ├── start.go           # Start embedded MindsDB
│   ├── stop.go            # Stop embedded MindsDB
│   ├── status.go          # Check MindsDB status
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
   - **Start** (`start.go`): Starts embedded MindsDB in Docker
   - **Stop** (`stop.go`): Stops embedded MindsDB container
   - **Status** (`status.go`): Checks Docker and container status
   - **Connect** (`connect.go`): Handles connection to MindsDB instances (embedded/external)
   - **Create Model** (`create_model.go`): Manages model creation workflow
   - **List Models** (`list_models.go`): Lists available models
   - **Query** (`query.go`): Executes SQL queries and predictions

3. **MindsDB Client** (`internal/mindsdb/client.go`)
   - **PostgreSQL Client**: For communicating with external MindsDB instances
   - **MySQL Client**: For communicating with embedded MindsDB instances
   - **Docker Management**: Complete container lifecycle management (start, stop, status)
   - **Connection Management**: Automatic protocol detection and connection handling
   - **Health Checking**: Ensures MindsDB is ready before connecting

#### Dependencies

- **[Cobra](https://github.com/spf13/cobra)**: Modern CLI framework for Go
- **[pgx](https://github.com/jackc/pgx)**: PostgreSQL driver for Go (external MindsDB connections)
- **[go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)**: MySQL driver for Go (embedded MindsDB connections)
- **[fatih/color](https://github.com/fatih/color)**: Colored terminal output

#### Design Patterns

- **Command Pattern**: Each CLI command is implemented as a separate Cobra command
- **Client Pattern**: MindsDB client abstracts connection and communication logic
- **Flag-based Configuration**: Commands use flags for parameter input
- **Protocol Abstraction**: Supports both PostgreSQL and MySQL protocols transparently

## ✅ Embedded MindsDB Implementation

### Vision: Self-Contained MindsDB CLI

The CLI now provides a completely self-contained MindsDB experience, eliminating the need for users to install MindsDB separately.

### Implementation Details

#### Docker-based Embedding ✅ (Fully Implemented)
- ✅ Docker integration for embedded MindsDB
- ✅ Container lifecycle management (start, stop, status)
- ✅ Automatic MindsDB image download and setup
- ✅ Health checking and connection management
- ✅ Port management and networking
- ✅ Data persistence across container restarts

### Technical Architecture

```
mindsdb-cli
├── Docker Management Layer ✅
│   ├── Container lifecycle (start/stop/status)
│   ├── Image management (pull/update)
│   └── Port management and networking
├── Connection Abstraction ✅
│   ├── Embedded mode (localhost Docker + MySQL)
│   └── External mode (remote MindsDB + PostgreSQL)
└── CLI Interface ✅
    ├── Unified commands work with both modes
    └── Automatic mode detection
```

### Benefits of Embedded Approach

1. **✅ Zero Installation**: No need to install MindsDB separately
2. **✅ Version Consistency**: CLI and MindsDB versions are matched
3. **✅ Isolated Environment**: No conflicts with system Python/packages
4. **✅ Easy Updates**: Single binary update includes everything
5. **✅ Portability**: Works anywhere Docker is available
6. **✅ Quick Setup**: Get started in minutes

## 🛠️ Development

### Prerequisites for Development

- Go 1.23 or higher
- Git
- Docker (for embedded features)
- *Optional*: A running MindsDB instance for testing external connections

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

## 📞 Support

For questions and support:
- Check the [MindsDB Documentation](https://docs.mindsdb.com/)
- Open an issue in this repository
- Join the [MindsDB Community](https://mindsdb.com/community)

## 🗺️ Roadmap

### Current Version (v0.2.0) ✅
- ✅ Basic CLI structure with Cobra
- ✅ PostgreSQL connection to external MindsDB instances
- ✅ MySQL connection to embedded MindsDB instances
- ✅ Core commands: connect, list-models, create-model, query
- ✅ Docker integration for embedded MindsDB
- ✅ Container lifecycle management (start/stop/status)
- ✅ Automatic MindsDB image download
- ✅ Health checking and auto-connection
- ✅ Cross-platform builds

### Next Version (v0.3.0) - Enhanced Features
- 📋 Enhanced model management features
- 📋 Configuration file support
- 📋 Interactive mode with auto-completion
- 📋 Export/import functionality for models
- 📋 Integration with MindsDB Cloud features
- 📋 Comprehensive test suite

### Future Versions
- 📋 Binary embedding (explore MindsDB as Go library)
- 📋 Plugin system for custom extensions
- 📋 Advanced monitoring and logging
- 📋 Multi-container orchestration
- 📋 Performance optimization and caching

---

**Version**: 0.2.0  
**Go Version**: 1.23+
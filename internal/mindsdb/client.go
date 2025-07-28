package mindsdb

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
)

const (
	MindsDBImage  = "mindsdb/mindsdb:latest"
	ContainerName = "mindsdb-cli-embedded"
	MindsDBPort   = "47334"
	MySQLPort     = "47335" // MindsDB uses MySQL protocol
)

type MindsDBClient struct {
	PgConn       *pgx.Conn // For PostgreSQL connections (external)
	MySQLConn    *sql.DB   // For MySQL connections (embedded)
	ContainerID  string
	EmbeddedMode bool
	IsMySQL      bool
}

// NewClient creates a client for external MindsDB connection (PostgreSQL)
func NewClient(host, user, pass string) (*MindsDBClient, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/mindsdb", user, pass, host)
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MindsDB: %w", err)
	}
	return &MindsDBClient{PgConn: conn, EmbeddedMode: false, IsMySQL: false}, nil
}

// NewEmbeddedClient creates a client with embedded MindsDB using Docker CLI
func NewEmbeddedClient(user, pass string) (*MindsDBClient, error) {
	// Check if Docker is available
	if !IsDockerAvailable() {
		return nil, fmt.Errorf("Docker is not available - required for embedded mode")
	}

	client := &MindsDBClient{EmbeddedMode: true, IsMySQL: true}

	// Start the container if not running
	containerID, err := client.StartEmbeddedMindsDB(user, pass)
	if err != nil {
		return nil, fmt.Errorf("failed to start embedded MindsDB: %w", err)
	}
	client.ContainerID = containerID

	// Try connecting without credentials first (MindsDB default behavior)
	fmt.Println("🔐 Trying connection with MindsDB defaults (user: mindsdb, no password)...")
	mysqlDSN := fmt.Sprintf("mindsdb:@tcp(localhost:%s)/mindsdb", MySQLPort)

	mysqlConn, err := sql.Open("mysql", mysqlDSN)
	if err == nil {
		if err = mysqlConn.Ping(); err == nil {
			fmt.Println("✅ Connected successfully with MindsDB defaults")
			client.MySQLConn = mysqlConn
			return client, nil
		}
		mysqlConn.Close()
	}
	fmt.Println("❌ Failed with MindsDB defaults, trying with provided credentials...")

	// If no-auth fails, try with provided credentials
	if user != "" && pass != "" {
		fmt.Printf("🔐 Trying provided credentials (%s)...\n", user)
		mysqlDSN = fmt.Sprintf("%s:%s@tcp(localhost:%s)/mindsdb", user, pass, MySQLPort)

		mysqlConn, err = sql.Open("mysql", mysqlDSN)
		if err == nil {
			if err = mysqlConn.Ping(); err == nil {
				fmt.Printf("✅ Connected successfully with provided credentials\n")
				client.MySQLConn = mysqlConn
				return client, nil
			}
			mysqlConn.Close()
		}
		fmt.Printf("❌ Failed with provided credentials\n")
	}

	return nil, fmt.Errorf("failed to connect to MindsDB. Default credentials are user 'mindsdb' with empty password. Last error: %w", err)
}

// Query executes a SQL query on the appropriate connection
func (c *MindsDBClient) Query(query string) (*sql.Rows, error) {
	if c.IsMySQL && c.MySQLConn != nil {
		return c.MySQLConn.Query(query)
	} else if c.PgConn != nil {
		// For PostgreSQL, we need to handle this differently
		return nil, fmt.Errorf("PostgreSQL query execution needs implementation")
	}
	return nil, fmt.Errorf("no valid connection available")
}

// QueryPg executes a PostgreSQL query (for external connections)
func (c *MindsDBClient) QueryPg(query string) (pgx.Rows, error) {
	if c.PgConn == nil {
		return nil, fmt.Errorf("no PostgreSQL connection available")
	}
	return c.PgConn.Query(context.Background(), query)
}

// IsDockerAvailable checks if Docker is installed and running
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	err := cmd.Run()
	return err == nil
}

// StartEmbeddedMindsDB starts a MindsDB container using Docker CLI
func (c *MindsDBClient) StartEmbeddedMindsDB(user, pass string) (string, error) {
	// Check if container already exists and is running
	if containerID := c.findExistingContainer(); containerID != "" {
		if c.isContainerRunning(containerID) {
			fmt.Println("✅ MindsDB container is already running")
			return containerID, nil
		}

		// Container exists but not running, start it
		fmt.Println("▶️  Starting existing MindsDB container...")
		if err := c.startContainer(containerID); err == nil {
			if err := c.waitForMindsDB(user, pass); err != nil {
				return "", err
			}
			return containerID, nil
		}
	}

	// Pull the image
	fmt.Println("📥 Pulling MindsDB Docker image...")
	cmd := exec.Command("docker", "pull", MindsDBImage)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to pull MindsDB image: %w", err)
	}

	// Create and start container
	fmt.Println("🚀 Creating MindsDB container...")
	cmd = exec.Command("docker", "run", "-d",
		"--name", ContainerName,
		"-p", MindsDBPort+":"+MindsDBPort,
		"-p", MySQLPort+":"+MySQLPort,
		"-e", "MINDSDB_DB_SERVICE_HOST=0.0.0.0",
		"-e", "MINDSDB_DB_SERVICE_PORT="+MySQLPort,
		MindsDBImage)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	fmt.Println("✅ MindsDB container started successfully")

	// Wait for MindsDB to be ready
	if err := c.waitForMindsDB(user, pass); err != nil {
		return "", err
	}

	return containerID, nil
}

// findExistingContainer looks for an existing MindsDB container
func (c *MindsDBClient) findExistingContainer() string {
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name="+ContainerName, "--format", "{{.ID}}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	containerID := strings.TrimSpace(string(output))
	return containerID
}

// isContainerRunning checks if a container is currently running
func (c *MindsDBClient) isContainerRunning(containerID string) bool {
	cmd := exec.Command("docker", "ps", "--filter", "id="+containerID, "--format", "{{.ID}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

// startContainer starts an existing container
func (c *MindsDBClient) startContainer(containerID string) error {
	cmd := exec.Command("docker", "start", containerID)
	return cmd.Run()
}

// waitForMindsDB waits for MindsDB to be ready to accept connections
func (c *MindsDBClient) waitForMindsDB(user, pass string) error {
	fmt.Print("⏳ Waiting for MindsDB to be ready")

	mysqlDSN := fmt.Sprintf("%s:%s@tcp(localhost:%s)/mindsdb", user, pass, MySQLPort)

	maxAttempts := 30
	for i := 1; i <= maxAttempts; i++ {
		db, err := sql.Open("mysql", mysqlDSN)
		if err == nil {
			if err := db.Ping(); err == nil {
				db.Close()
				fmt.Println(" ✅")
				fmt.Printf("🎉 MindsDB is ready! Web UI: http://localhost:%s\n", MindsDBPort)
				return nil
			}
			db.Close()
		}

		fmt.Print(".")
		time.Sleep(2 * time.Second)
	}

	fmt.Println(" ❌")
	return fmt.Errorf("MindsDB did not become ready after %d seconds", maxAttempts*2)
}

// StopEmbeddedMindsDB stops the MindsDB container
func (c *MindsDBClient) StopEmbeddedMindsDB(remove bool) error {
	containerID := c.findExistingContainer()
	if containerID == "" {
		return fmt.Errorf("MindsDB container not found")
	}

	// Stop the container
	fmt.Println("🛑 Stopping MindsDB container...")
	cmd := exec.Command("docker", "stop", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	if remove {
		// Remove the container
		cmd = exec.Command("docker", "rm", containerID)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to remove container: %w", err)
		}
		fmt.Println("🗑️  MindsDB container stopped and removed")
	} else {
		fmt.Println("✅ MindsDB container stopped successfully")
	}

	return nil
}

// GetContainerStatus returns the status of the MindsDB container
func (c *MindsDBClient) GetContainerStatus() (bool, string, error) {
	containerID := c.findExistingContainer()
	if containerID == "" {
		return false, "", nil // Container doesn't exist
	}

	// Check if running
	isRunning := c.isContainerRunning(containerID)

	// Get start time
	cmd := exec.Command("docker", "inspect", containerID, "--format", "{{.State.StartedAt}}")
	output, err := cmd.Output()
	startedAt := ""
	if err == nil {
		startedAt = strings.TrimSpace(string(output))
	}

	return isRunning, startedAt, nil
}

// Close closes the client connections
func (c *MindsDBClient) Close() {
	if c.PgConn != nil {
		c.PgConn.Close(context.Background())
	}
	if c.MySQLConn != nil {
		c.MySQLConn.Close()
	}
}

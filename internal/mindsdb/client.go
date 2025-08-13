package mindsdb

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strconv"
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
		return nil, fmt.Errorf("docker is not available - required for embedded mode")
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
	mysqlDSN := fmt.Sprintf("mindsdb:@tcp(localhost:%s)/mindsdb?timeout=10s&readTimeout=10s&writeTimeout=10s&parseTime=true", MySQLPort)

	mysqlConn, err := sql.Open("mysql", mysqlDSN)
	if err == nil {
		if err = mysqlConn.Ping(); err == nil {
			fmt.Println("✅ Connected successfully with MindsDB defaults")
			client.MySQLConn = mysqlConn
			return client, nil
		}
		mysqlConn.Close()
	}
	fmt.Printf("❌ Failed with MindsDB defaults: %v\n", err)
	fmt.Println("🔄 Trying with provided credentials...")

	// If no-auth fails, try with provided credentials
	if user != "" && pass != "" {
		fmt.Printf("🔐 Trying provided credentials (%s)...\n", user)
		mysqlDSN = fmt.Sprintf("%s:%s@tcp(localhost:%s)/mindsdb?timeout=10s&readTimeout=10s&writeTimeout=10s&parseTime=true", user, pass, MySQLPort)

		mysqlConn, err = sql.Open("mysql", mysqlDSN)
		if err == nil {
			if err = mysqlConn.Ping(); err == nil {
				fmt.Printf("✅ Connected successfully with provided credentials\n")
				client.MySQLConn = mysqlConn
				return client, nil
			}
			mysqlConn.Close()
		}
		fmt.Printf("❌ Failed with provided credentials: %v\n", err)
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

// checkSystemResources provides warnings about system resource constraints
func (c *MindsDBClient) checkSystemResources() {
	fmt.Println("🔍 Checking system resources...")

	// Check available memory
	if cmd := exec.Command("sh", "-c", "free -m 2>/dev/null | awk 'NR==2{printf \"%.0f\", $7}'"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			if availableMemMB := strings.TrimSpace(string(output)); availableMemMB != "" {
				if mem, err := strconv.Atoi(availableMemMB); err == nil {
					if mem < 2048 {
						fmt.Printf("⚠️  Available memory: %dMB (MindsDB recommends 2GB+)\n", mem)
						fmt.Println("   Consider using a larger EC2 instance if MindsDB fails to start")
					} else {
						fmt.Printf("✅ Available memory: %dMB\n", mem)
					}
				}
			}
		}
	}

	// Check Docker memory limit
	if cmd := exec.Command("docker", "system", "info", "--format", "{{.MemTotal}}"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			if dockerMemStr := strings.TrimSpace(string(output)); dockerMemStr != "" {
				if dockerMem, err := strconv.ParseInt(dockerMemStr, 10, 64); err == nil {
					dockerMemGB := dockerMem / (1024 * 1024 * 1024)
					if dockerMemGB < 2 {
						fmt.Printf("⚠️  Docker memory limit: %dGB (MindsDB needs 2GB+)\n", dockerMemGB)
						fmt.Println("   You may need to increase Docker's memory allocation")
					} else {
						fmt.Printf("✅ Docker memory limit: %dGB\n", dockerMemGB)
					}
				}
			}
		}
	}
}

// StartEmbeddedMindsDB starts a MindsDB container using Docker CLI
func (c *MindsDBClient) StartEmbeddedMindsDB(user, pass string) (string, error) {
	// Check system resources first (helpful for EC2 instances)
	c.checkSystemResources()

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

	// Create and start container with EC2-optimized settings
	fmt.Println("🚀 Creating MindsDB container...")
	cmd = exec.Command("docker", "run", "-d",
		"--name", ContainerName,
		"-p", MindsDBPort+":"+MindsDBPort,
		"-p", MySQLPort+":"+MySQLPort,
		// Resource limits for EC2 compatibility
		"--memory=2g",
		"--memory-swap=4g",
		"--cpus=1.5",
		"--oom-kill-disable=false",
		// Health check
		"--health-cmd=mysqladmin ping -h localhost -P "+MySQLPort+" -u mindsdb --silent",
		"--health-interval=30s",
		"--health-timeout=10s",
		"--health-retries=3",
		"--health-start-period=60s",
		// Environment variables
		"-e", "MINDSDB_DB_SERVICE_HOST=0.0.0.0",
		"-e", "MINDSDB_DB_SERVICE_PORT="+MySQLPort,
		// Performance tuning for cloud environments
		"-e", "MINDSDB_STORAGE_ENGINE=sqlite",
		"-e", "PYTHONUNBUFFERED=1",
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

	// Build a list of DSNs to try each attempt. Always try MindsDB defaults first,
	// then try provided credentials if they were supplied.
	// Use more generous timeouts to avoid "unexpected EOF" errors during startup
	defaultDSN := fmt.Sprintf("mindsdb:@tcp(localhost:%s)/mindsdb?timeout=10s&readTimeout=10s&writeTimeout=10s&parseTime=true", MySQLPort)
	var providedDSN string
	if user != "" { // only construct provided DSN if a username was supplied
		providedDSN = fmt.Sprintf("%s:%s@tcp(localhost:%s)/mindsdb?timeout=10s&readTimeout=10s&writeTimeout=10s&parseTime=true", user, pass, MySQLPort)
	}

	// Extend timeout for cloud/EC2 environments where startup can be slower
	maxAttempts := 60 // Doubled from 30 to 60 for EC2
	var lastErr error

	// Wait longer initially for MindsDB to start up on EC2
	fmt.Println("\n💡 EC2 tip: MindsDB startup can take 2-3 minutes on cloud instances")
	time.Sleep(10 * time.Second) // Increased from 5 to 10 seconds

	for i := 1; i <= maxAttempts; i++ {
		// Attempt with default credentials first
		if db, err := sql.Open("mysql", defaultDSN); err == nil {
			if err := db.Ping(); err == nil {
				db.Close()
				fmt.Println(" ✅")
				fmt.Printf("🎉 MindsDB is ready! Web UI: http://localhost:%s\n", MindsDBPort)
				return nil
			} else {
				lastErr = err
			}
			db.Close()
		}

		// Then attempt with provided credentials (if any)
		if providedDSN != "" {
			if db, err := sql.Open("mysql", providedDSN); err == nil {
				if err := db.Ping(); err == nil {
					db.Close()
					fmt.Println(" ✅")
					fmt.Printf("🎉 MindsDB is ready! Web UI: http://localhost:%s\n", MindsDBPort)
					return nil
				} else {
					lastErr = err
				}
				db.Close()
			}
		}

		// Use progressive backoff - wait longer between attempts as time goes on
		// More patient for EC2/cloud environments
		waitTime := 4 * time.Second
		if i > 15 {
			waitTime = 6 * time.Second
		}
		if i > 30 {
			waitTime = 8 * time.Second
		}
		if i > 45 {
			waitTime = 10 * time.Second
		}

		fmt.Print(".")
		time.Sleep(waitTime)
	}

	fmt.Println(" ❌")
	if lastErr != nil {
		return fmt.Errorf("MindsDB did not become ready after %d attempts (up to 10 minutes). Last error: %v", maxAttempts, lastErr)
	}
	return fmt.Errorf("MindsDB did not become ready after %d attempts (up to 10 minutes)", maxAttempts)
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

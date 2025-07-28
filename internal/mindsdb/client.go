package mindsdb

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
)

type MindsDBClient struct {
    Conn *pgx.Conn
}

func NewClient(host, user, pass string) (*MindsDBClient, error) {
    dsn := fmt.Sprintf("postgres://%s:%s@%s/mindsdb", user, pass, host)
    conn, err := pgx.Connect(context.Background(), dsn)
    if err != nil {
        return nil, err
    }
    return &MindsDBClient{Conn: conn}, nil
}

func (c *MindsDBClient) Close() {
    if c.Conn != nil {
        c.Conn.Close(context.Background())
    }
}

func (c *MindsDBClient) QueryVersion() (string, error) {
    var version string
    err := c.Conn.QueryRow(context.Background(), "SELECT * FROM mindsdb.mindsdb_version").Scan(&version)
    if err != nil {
        return "", err
    }
    return version, nil
}
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/nathanthorell/dataspy/config"
)

func ExecuteRule(server config.DbServer, rule config.Rule) error {
	fmt.Printf("\nExecuting Rule [%s] on DB Server [%s]\n", rule.Name, server.Name)
	connStr, err := server.GetConnString()
	if err != nil {
		return fmt.Errorf("failed to get connection string for server %s: %w", server.Name, err)
	}

	db, err := sql.Open(server.Type, connStr)
	if err != nil {
		return fmt.Errorf("failed to open db connection: %w", err)
	}
	defer db.Close()

	// Check if the connection is successful
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Execute a simple SQL query
	var result string
	err = db.QueryRow(rule.Query).Scan(&result)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	fmt.Println(result)
	return nil
}

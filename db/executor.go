package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/nathanthorell/dataspy/config"
)

type ExecutionResult struct {
	RowCount int64
	Results  string
}

func ExecuteRule(server config.DbServer, rule config.Rule) (ExecutionResult, error) {
	fmt.Printf("\nExecuting Rule [%s] on DB Server [%s]\n", rule.Name, server.Name)
	connStr, err := server.GetConnString()
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("failed to get connection string for server %s: %w", server.Name, err)
	}

	db, err := sql.Open(server.Type, connStr)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("failed to open db connection: %w", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("failed to ping database: %w", err)
	}

	rows, err := db.Query(rule.Query)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("failed to get columns: %w", err)
	}

	var results [][]string
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return ExecutionResult{}, fmt.Errorf("failed to scan row: %w", err)
		}

		rowStrings := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				rowStrings[i] = "NULL"
			} else {
				// Handle []byte type specifically for MySQL
				switch v := val.(type) {
				case []byte:
					rowStrings[i] = string(v)
				default:
					rowStrings[i] = fmt.Sprintf("%v", v)
				}
			}
		}
		results = append(results, rowStrings)
	}

	// Format results into a string
	var resultString string
	if len(results) > 0 {
		resultString = fmt.Sprintf("Found %d rows:\n", len(results))
		for i, row := range results {
			// Use strings.Join for cleaner output
			resultString += fmt.Sprintf("Row %d: %s\n", i+1, strings.Join(row, " "))
		}
	} else {
		resultString = "Query completed successfully (0 rows)"
	}

	return ExecutionResult{
		RowCount: int64(len(results)),
		Results:  resultString,
	}, nil
}

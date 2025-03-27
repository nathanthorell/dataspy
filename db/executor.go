package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/nathanthorell/dataspy/config"
)

// dbOpener is a function type that matches sql.Open's signature
type dbOpener func(driverName, dataSource string) (*sql.DB, error)

type ExecutionResult struct {
	RowCount  int64
	Results   string
	LogEvents []LogEvent
}

type LogEvent struct {
	Level   string // "info", "error", "success", "warn", "task", "rule", "db"
	Message string
	Fields  map[string]interface{}
	Error   error // Only for error events
}

func executeRuleWithOpener(
	server config.DbServer,
	rule config.Rule,
	opener dbOpener,
) (ExecutionResult, error) {
	result := ExecutionResult{
		LogEvents: make([]LogEvent, 0),
	}

	result.LogEvents = append(result.LogEvents, LogEvent{
		Level:   "task",
		Message: fmt.Sprintf("Executing on %s", server.Name),
		Fields:  map[string]interface{}{"rule": rule.Name},
	})

	connStr, err := server.GetConnString()
	if err != nil {
		result.LogEvents = append(result.LogEvents, LogEvent{
			Level:   "error",
			Message: "Failed to get connection string",
			Fields:  map[string]interface{}{"server": server.Name},
			Error:   err,
		})
		return result, fmt.Errorf("failed to get connection string for server %s: %w", server.Name, err)
	}

	db, err := opener(server.Type, connStr)
	if err != nil {
		result.LogEvents = append(result.LogEvents, LogEvent{
			Level:   "error",
			Message: "Failed to open DB connection",
			Fields:  map[string]interface{}{"server": server.Name},
			Error:   err,
		})
		return result, fmt.Errorf("failed to open db connection: %w", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		result.LogEvents = append(result.LogEvents, LogEvent{
			Level:   "error",
			Message: "Failed to ping database",
			Fields:  map[string]interface{}{"server": server.Name},
			Error:   err,
		})
		return result, fmt.Errorf("failed to ping database: %w", err)
	}

	result.LogEvents = append(result.LogEvents, LogEvent{
		Level:   "db",
		Message: "Connection established successfully",
		Fields:  map[string]interface{}{"server": server.Name},
	})
	
	result.LogEvents = append(result.LogEvents, LogEvent{
		Level:   "rule",
		Message: "Executing query",
		Fields:  map[string]interface{}{"rule": rule.Name},
	})

	rows, err := db.Query(rule.Query)
	if err != nil {
		result.LogEvents = append(result.LogEvents, LogEvent{
			Level:   "error",
			Message: "Failed to execute query",
			Fields:  map[string]interface{}{"rule": rule.Name},
			Error:   err,
		})
		return result, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		result.LogEvents = append(result.LogEvents, LogEvent{
			Level:   "error",
			Message: "Failed to get columns",
			Fields:  map[string]interface{}{"rule": rule.Name},
			Error:   err,
		})
		return result, fmt.Errorf("failed to get columns: %w", err)
	}

	var results [][]string
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			result.LogEvents = append(result.LogEvents, LogEvent{
				Level:   "error",
				Message: "Failed to scan row",
				Fields:  map[string]interface{}{"rule": rule.Name},
				Error:   err,
			})
			return result, fmt.Errorf("failed to scan row: %w", err)
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

	var resultString string
	if len(results) > 0 {
		resultString = fmt.Sprintf("Found %d rows:\n", len(results))
		for i, row := range results {
			resultString += fmt.Sprintf("Row %d: %s\n", i+1, strings.Join(row, " "))
		}
	} else {
		resultString = "Query completed successfully (0 rows)"
	}

	result.LogEvents = append(result.LogEvents, LogEvent{
		Level:   "success",
		Message: "Query executed successfully",
		Fields: map[string]interface{}{
			"rule":   rule.Name,
			"server": server.Name,
			"rows":   len(results),
		},
	})

	result.RowCount = int64(len(results))
	result.Results = resultString
	
	return result, nil
}

func ExecuteRule(server config.DbServer, rule config.Rule) (ExecutionResult, error) {
	return executeRuleWithOpener(server, rule, sql.Open)
}

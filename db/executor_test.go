package db

import (
	"database/sql"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nathanthorell/dataspy/config"
	"github.com/stretchr/testify/assert"
)

// openTestDB returns a mock database for testing
func openTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	return mockDB, mock
}

func TestExecuteRule(t *testing.T) {
	tests := []struct {
		name        string
		server      config.DbServer
		rule        config.Rule
		setupMock   func(mock sqlmock.Sqlmock)
		expectErr   bool
		expectRows  int64
		expectMatch string
	}{
		{
			name: "successful query with results",
			server: config.DbServer{
				Name:          "test-server",
				Type:          "postgres",
				ConnStringVar: "PG_DBCONN",
			},
			rule: config.Rule{
				Name:        "test-rule",
				Query:       "SELECT version()",
				DbType:      "postgres",
				Description: "Test getting postgres version",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
				rows := sqlmock.NewRows([]string{"version"}).
					AddRow("PostgreSQL 14.0")
				mock.ExpectQuery("SELECT version()").
					WillReturnRows(rows)
			},
			expectErr:   false,
			expectRows:  1,
			expectMatch: "PostgreSQL 14.0",
		},
		{
			name: "empty results",
			server: config.DbServer{
				Name:          "test-server",
				Type:          "postgres",
				ConnStringVar: "PG_DBCONN",
			},
			rule: config.Rule{
				Name:        "empty-rule",
				Query:       "SELECT * FROM non_existent WHERE false",
				DbType:      "postgres",
				Description: "Test empty results",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
				rows := sqlmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery("SELECT \\* FROM non_existent WHERE false").
					WillReturnRows(rows)
			},
			expectErr:   false,
			expectRows:  0,
			expectMatch: "Query completed successfully (0 rows)",
		},
		{
			name: "query execution error",
			server: config.DbServer{
				Name:          "test-server",
				Type:          "postgres",
				ConnStringVar: "PG_DBCONN",
			},
			rule: config.Rule{
				Name:        "error-rule",
				Query:       "SELECT * FROM non_existent",
				DbType:      "postgres",
				Description: "Test query error",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
				mock.ExpectQuery("SELECT \\* FROM non_existent").
					WillReturnError(assert.AnError)
			},
			expectErr: true,
		},
		{
			name: "missing environment variable",
			server: config.DbServer{
				Name:          "test-server",
				Type:          "postgres",
				ConnStringVar: "NONEXISTENT_CONN",
			},
			rule: config.Rule{
				Name:        "env-error-rule",
				Query:       "SELECT 1",
				DbType:      "postgres",
				Description: "Test missing connection string",
			},
			setupMock: func(mock sqlmock.Sqlmock) {},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock database for each test case
			mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			if err != nil {
				t.Fatalf("Failed to create mock: %v", err)
			}
			defer mockDB.Close()

			// Create a database opener that returns our mock
			dbOpen := func(driverName, dataSource string) (*sql.DB, error) {
				return mockDB, nil
			}

			if tt.server.ConnStringVar == "PG_DBCONN" {
				os.Setenv(tt.server.ConnStringVar, "mock_conn_string")
				defer os.Unsetenv(tt.server.ConnStringVar)

				tt.setupMock(mock)
			}

			result, err := executeRuleWithOpener(tt.server, tt.rule, dbOpen)

			// Verify expectations
			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectRows, result.RowCount)
			if tt.expectMatch != "" {
				assert.Contains(t, result.Results, tt.expectMatch)
			}

			if !tt.expectErr {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

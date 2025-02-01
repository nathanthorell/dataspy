package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nathanthorell/dataspy/config"
	"github.com/nathanthorell/dataspy/db"
	"github.com/nathanthorell/dataspy/storage"
	"github.com/stretchr/testify/assert"
)

type testFixtures struct {
	store     *storage.Store
	dbPath    string
	cleanup   func()
	config    config.Config
	scheduler *Scheduler
}

// --------- HELPERS ---------

func createTestStore(t *testing.T) (*storage.Store, string, func()) {
	tmpDir, err := os.MkdirTemp("", "dataspy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := storage.NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return store, dbPath, cleanup
}

func setupTest(t *testing.T) *testFixtures {
	store, dbPath, cleanup := createTestStore(t)

	cfg := config.Config{
		DBServers: []config.DbServer{
			{
				Name:          "test-postgres",
				Type:          "postgres",
				ConnStringVar: "PG_TEST_CONN",
			},
			{
				Name:          "test-mysql",
				Type:          "mysql",
				ConnStringVar: "MYSQL_TEST_CONN",
			},
		},
		Rules: []config.Rule{
			{
				Name:        "test-rule",
				Description: "Test rule description",
				DbType:      "postgres",
				Query:       "SELECT 1",
			},
		},
	}

	scheduler := NewScheduler(cfg, store)

	return &testFixtures{
		store:     store,
		dbPath:    dbPath,
		cleanup:   cleanup,
		config:    cfg,
		scheduler: scheduler,
	}
}

// --------- TESTS ---------

func TestNewScheduler(t *testing.T) {
	f := setupTest(t)
	defer f.cleanup()

	assert.NotNil(t, f.scheduler)
	assert.NotNil(t, f.scheduler.scheduler)
	assert.Equal(t, f.config, f.scheduler.config)
	assert.Equal(t, f.store, f.scheduler.store)
}

func TestAddTask(t *testing.T) {
	f := setupTest(t)
	defer f.cleanup()

	tests := []struct {
		name        string
		schedule    config.Schedule
		shouldError bool
	}{
		{
			name: "valid schedule",
			schedule: config.Schedule{
				Server:  "test-server",
				Rule:    "test-rule",
				CronStr: "*/5 * * * * *",
			},
			shouldError: false,
		},
		{
			name: "invalid cron expression",
			schedule: config.Schedule{
				Server:  "test-server",
				Rule:    "test-rule",
				CronStr: "invalid cron",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.scheduler.addTask(tt.schedule)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "error scheduling task")
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, f.scheduler.scheduler.Entries())
			}
		})
	}
}

func TestFindRule(t *testing.T) {
	f := setupTest(t)
	defer f.cleanup()

	tests := []struct {
		name       string
		ruleName   string
		shouldFind bool
		wantRule   config.Rule
	}{
		{
			name:       "existing rule",
			ruleName:   "test-rule",
			shouldFind: true,
			wantRule: config.Rule{
				Name:        "test-rule",
				Description: "Test rule description",
				DbType:      "postgres",
				Query:       "SELECT 1",
			},
		},
		{
			name:       "non-existent rule",
			ruleName:   "missing-rule",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := f.scheduler.findRule(tt.ruleName)

			if tt.shouldFind {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRule, rule)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "rule not found")
			}
		})
	}
}

func TestFindServer(t *testing.T) {
	f := setupTest(t)
	defer f.cleanup()

	tests := []struct {
		name       string
		dbType     string
		shouldFind bool
		wantServer config.DbServer
	}{
		{
			name:       "existing postgres server",
			dbType:     "postgres",
			shouldFind: true,
			wantServer: config.DbServer{
				Name:          "test-postgres",
				Type:          "postgres",
				ConnStringVar: "PG_TEST_CONN",
			},
		},
		{
			name:       "existing mysql server",
			dbType:     "mysql",
			shouldFind: true,
			wantServer: config.DbServer{
				Name:          "test-mysql",
				Type:          "mysql",
				ConnStringVar: "MYSQL_TEST_CONN",
			},
		},
		{
			name:       "non-existent server",
			dbType:     "oracle",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := f.scheduler.findServer(tt.dbType)

			if tt.shouldFind {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantServer, server)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "server not found")
			}
		})
	}
}

func TestRecordExecution(t *testing.T) {
	f := setupTest(t)
	defer f.cleanup()

	testRule := config.Rule{
		Name:        "test-execution-rule",
		Description: "Rule for testing execution recording",
		DbType:      "postgres",
		Query:       "SELECT 1",
	}

	testServer := config.DbServer{
		Name:          "test-execution-server",
		Type:          "postgres",
		ConnStringVar: "PG_TEST_CONN",
	}

	tests := []struct {
		name       string
		rule       config.Rule
		server     config.DbServer
		result     db.ExecutionResult
		executeErr error
		wantStatus string
	}{
		{
			name:       "successful execution",
			rule:       testRule,
			server:     testServer,
			result:     db.ExecutionResult{RowCount: 1, Results: "test result"},
			executeErr: nil,
			wantStatus: "success",
		},
		{
			name:       "failed execution",
			rule:       testRule,
			server:     testServer,
			result:     db.ExecutionResult{},
			executeErr: fmt.Errorf("test error"),
			wantStatus: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now().Add(-time.Second) // Execution started 1 second ago
			f.scheduler.recordExecution(tt.rule, tt.server, startTime, tt.result, tt.executeErr)

			records, err := f.store.GetExecutionsByRule(tt.rule.Name)
			assert.NoError(t, err)
			assert.NotEmpty(t, records)

			lastRecord := records[len(records)-1]
			assert.Equal(t, tt.rule.Name, lastRecord.RuleName)
			assert.Equal(t, tt.server.Name, lastRecord.ServerName)
			assert.Equal(t, tt.wantStatus, lastRecord.Status)

			if tt.executeErr != nil {
				assert.Equal(t, tt.executeErr.Error(), lastRecord.Error)
				assert.Empty(t, lastRecord.Result)
			} else {
				assert.Empty(t, lastRecord.Error)
				assert.Equal(t, tt.result.Results, lastRecord.Result)
			}

			assert.True(t, lastRecord.Duration >= 1000.0)
		})
	}
}

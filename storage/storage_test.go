package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dataspy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	record := &ExecutionRecord{
		RuleName:     "test-rule",
		ServerName:   "test-server",
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(time.Second),
		Status:       "success",
		Result:       "Found 1 row:\nRow 1: [value1 value2]",
		Description:  "Test rule description",
		Duration:     1000.0, // 1 second in milliseconds
		RowsAffected: 1,
	}

	if err := store.SaveExecutionRecord(record); err != nil {
		t.Fatalf("Failed to save execution record: %v", err)
	}

	records, err := store.GetLatestExecutions(1)
	if err != nil {
		t.Fatalf("Failed to get latest executions: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	got := records[0]
	if got.RuleName != record.RuleName {
		t.Errorf("Expected rule name %s, got %s", record.RuleName, got.RuleName)
	}
	if got.ServerName != record.ServerName {
		t.Errorf("Expected server name %s, got %s", record.ServerName, got.ServerName)
	}
	if got.Status != record.Status {
		t.Errorf("Expected status %s, got %s", record.Status, got.Status)
	}
	if got.Result != record.Result {
		t.Errorf("Expected result %s, got %s", record.Result, got.Result)
	}
	if got.Description != record.Description {
		t.Errorf("Expected description %s, got %s", record.Description, got.Description)
	}
	if got.Duration != record.Duration {
		t.Errorf("Expected duration %f, got %f", record.Duration, got.Duration)
	}
	if got.RowsAffected != record.RowsAffected {
		t.Errorf("Expected rows affected %d, got %d", record.RowsAffected, got.RowsAffected)
	}

	ruleRecords, err := store.GetExecutionsByRule("test-rule")
	if err != nil {
		t.Fatalf("Failed to get executions by rule: %v", err)
	}

	if len(ruleRecords) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(ruleRecords))
	}

	got = ruleRecords[0]
	if got.RuleName != record.RuleName {
		t.Errorf("Expected rule name %s, got %s", record.RuleName, got.RuleName)
	}
	if got.Result != record.Result {
		t.Errorf("Expected result %s, got %s", record.Result, got.Result)
	}
}

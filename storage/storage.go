package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

const (
	ExecutionHistoryBucket = "execution_history"
	RuleMetadataBucket     = "rule_metadata"
)

type Store struct {
	db *bbolt.DB
}

type ExecutionRecord struct {
	RuleName     string    `json:"rule_name"`
	ServerName   string    `json:"server_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Status       string    `json:"status"`
	Result       string    `json:"result"`
	Error        string    `json:"error,omitempty"`
	Description  string    `json:"description"`
	Duration     float64   `json:"duration_ms"`
	RowsAffected int64     `json:"rows_affected"`
}

// Helper function to ensure directory exists
func ensureDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}

func NewStore(dbPath string) (*Store, error) {
	// Ensure the directory exists
	if err := ensureDir(dbPath); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open the db
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Initialize buckets
	err = db.Update(func(tx *bbolt.Tx) error {
		// Create buckets if they don't exist
		for _, bucket := range []string{ExecutionHistoryBucket, RuleMetadataBucket} {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize buckets: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SaveExecutionRecord(record *ExecutionRecord) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(ExecutionHistoryBucket))

		// Generate a key using timestamp and rule name for ordering
		key := fmt.Sprintf("%d-%s-%s", record.StartTime.UnixNano(), record.RuleName, record.ServerName)

		value, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal record: %w", err)
		}

		return b.Put([]byte(key), value)
	})
}

func (s *Store) GetLatestExecutions(n int) ([]ExecutionRecord, error) {
	var records []ExecutionRecord

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(ExecutionHistoryBucket))
		c := b.Cursor()

		// Iterate backwards through the bucket
		for k, v := c.Last(); k != nil && len(records) < n; k, v = c.Prev() {
			var record ExecutionRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return fmt.Errorf("failed to unmarshal record: %w", err)
			}
			records = append(records, record)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get latest executions: %w", err)
	}

	return records, nil
}

func (s *Store) GetExecutionsByRule(ruleName string) ([]ExecutionRecord, error) {
	var records []ExecutionRecord

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(ExecutionHistoryBucket))
		c := b.Cursor()

		searchTerm := "-" + ruleName + "-"
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if !bytes.Contains(k, []byte(searchTerm)) {
				continue
			}

			var record ExecutionRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return fmt.Errorf("failed to unmarshal record: %w", err)
			}
			records = append(records, record)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get executions by rule: %w", err)
	}

	return records, nil
}

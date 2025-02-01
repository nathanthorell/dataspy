package runner

import (
	"fmt"
	"log"
	"time"

	"github.com/nathanthorell/dataspy/config"
	"github.com/nathanthorell/dataspy/db"
	"github.com/nathanthorell/dataspy/storage"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	config    config.Config
	scheduler *cron.Cron
	store     *storage.Store
}

func NewScheduler(config config.Config, store *storage.Store) *Scheduler {
	return &Scheduler{
		config:    config,
		scheduler: cron.New(cron.WithSeconds()),
		store:     store,
	}
}

func (s *Scheduler) Start() error {
	for _, schedule := range s.config.Schedules {
		fmt.Printf("Adding task [%s] on schedule [%s] for DB Server [%s]\n", schedule.Rule, schedule.CronStr, schedule.Server)
		err := s.addTask(schedule)
		if err != nil {
			return fmt.Errorf("error adding task: %v", err)
		}
	}

	fmt.Println("\nStarting Scheduler...")
	s.scheduler.Start()

	// Add some debug info
	entries := s.scheduler.Entries()
	fmt.Printf("Number of scheduled tasks: %d\n", len(entries))
	for _, entry := range entries {
		fmt.Printf("Next run for task: %s\n", entry.Next)
	}
	return nil
}

func (s *Scheduler) addTask(schedule config.Schedule) error {
	fmt.Println("Adding task: ", schedule.Rule)
	entryID, err := s.scheduler.AddFunc(schedule.CronStr, func() {
		fmt.Printf("\nTriggering scheduled task at %s\n", time.Now().Format(time.RFC3339))
		s.runTask(schedule.Rule)
	})
	if err != nil {
		return fmt.Errorf("error scheduling task: %w", err)
	}
	fmt.Printf("Successfully scheduled task with ID: %d\n\n", entryID)
	return nil
}

func (s *Scheduler) runTask(scheduleName string) {
	startTime := time.Now()

	rule, err := s.findRule(scheduleName)
	if err != nil {
		s.recordExecution(config.Rule{Name: scheduleName}, config.DbServer{}, startTime, db.ExecutionResult{}, err)
		log.Printf("error finding rule: %v", err)
		return
	}

	server, err := s.findServer(rule.DbType)
	if err != nil {
		s.recordExecution(rule, config.DbServer{}, startTime, db.ExecutionResult{}, err)
		log.Printf("error finding server: %v", err)
		return
	}

	result, err := db.ExecuteRule(server, rule)
	s.recordExecution(rule, server, startTime, result, err)

	if err != nil {
		log.Printf("error executing rule %s: %v", rule.Name, err)
		return
	}

	fmt.Printf("Results for %s:\n%s\n", rule.Name, result.Results)
}

func (s *Scheduler) findRule(name string) (config.Rule, error) {
	for _, r := range s.config.Rules {
		if r.Name == name {
			return r, nil
		}
	}
	return config.Rule{}, fmt.Errorf("rule not found: %s", name)
}

func (s *Scheduler) findServer(dbType string) (config.DbServer, error) {
	for _, srv := range s.config.DBServers {
		if srv.Type == dbType {
			return srv, nil
		}
	}
	return config.DbServer{}, fmt.Errorf("server not found for db type: %s", dbType)
}

func (s *Scheduler) recordExecution(
	rule config.Rule,
	server config.DbServer,
	startTime time.Time,
	result db.ExecutionResult,
	err error,
) {
	endTime := time.Now()
	duration := float64(endTime.Sub(startTime).Milliseconds())

	record := &storage.ExecutionRecord{
		RuleName:     rule.Name,
		ServerName:   server.Name,
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       "success",
		Result:       result.Results,
		Description:  rule.Description,
		Duration:     duration,
		RowsAffected: result.RowCount,
	}

	if err != nil {
		record.Status = "error"
		record.Error = err.Error()
		record.Result = ""
	}

	if err := s.store.SaveExecutionRecord(record); err != nil {
		log.Printf("failed to save execution record: %v", err)
	}
}

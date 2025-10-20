package runner

import (
	"fmt"
	"time"

	"github.com/nathanthorell/dataspy/config"
	"github.com/nathanthorell/dataspy/db"
	"github.com/nathanthorell/dataspy/logger"
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
		logger.Info(fmt.Sprintf("Adding task [%s] on schedule [%s] for DB Server [%s]", schedule.Rule, schedule.CronStr, schedule.Server))
		err := s.addTask(schedule)
		if err != nil {
			return fmt.Errorf("error adding task: %v", err)
		}
	}

	logger.Info("Starting Scheduler...")
	s.scheduler.Start()

	// Add some debug info
	entries := s.scheduler.Entries()
	logger.Info(fmt.Sprintf("Number of scheduled tasks: %d", len(entries)))
	for _, entry := range entries {
		logger.Info(fmt.Sprintf("Next run for task: %s", entry.Next))
	}
	return nil
}

func (s *Scheduler) addTask(schedule config.Schedule) error {
	logger.Task(schedule.Rule, "Adding scheduled task")
	entryID, err := s.scheduler.AddFunc(schedule.CronStr, func() {
		logger.Info(fmt.Sprintf("Triggering scheduled task at %s\n", time.Now().Format(time.RFC3339)))
		s.runTask(schedule.Rule)
	})
	if err != nil {
		return fmt.Errorf("error scheduling task: %w", err)
	}
	logger.Success(fmt.Sprintf("Successfully scheduled task with ID: %d\n", entryID))
	return nil
}

func (s *Scheduler) runTask(scheduleName string) {
	if err := s.ExecuteRuleByName(scheduleName); err != nil {
		logger.Error(err, fmt.Sprintf("error executing scheduled task %s", scheduleName))
	}
}

// ExecuteRuleByName executes a rule by name and records the result
func (s *Scheduler) ExecuteRuleByName(ruleName string) error {
	startTime := time.Now()

	rule, err := s.findRule(ruleName)
	if err != nil {
		s.recordExecution(config.Rule{Name: ruleName}, config.DbServer{}, startTime, db.ExecutionResult{}, err)
		logger.Error(err, "error finding rule")
		return err
	}

	server, err := s.findServer(rule.DbType)
	if err != nil {
		s.recordExecution(rule, config.DbServer{}, startTime, db.ExecutionResult{}, err)
		logger.Error(err, "error finding server")
		return err
	}

	result, err := db.ExecuteRule(server, rule)
	s.processLogEvents(result.LogEvents)
	s.recordExecution(rule, server, startTime, result, err)

	if err != nil {
		logger.Error(err, fmt.Sprintf("error executing rule %s", rule.Name))
		return err
	}

	logger.Result(rule.Name, result.Results)
	return nil
}

// ExecuteAllRules executes all configured rules
func (s *Scheduler) ExecuteAllRules() (successCount int, errorCount int) {
	logger.Info(fmt.Sprintf("Running all %d rules", len(s.config.Rules)))

	for _, rule := range s.config.Rules {
		if err := s.ExecuteRuleByName(rule.Name); err != nil {
			errorCount++
		} else {
			successCount++
		}
		fmt.Println() // Add spacing between rule executions
	}

	logger.Info(fmt.Sprintf("Completed: %d successful, %d errors", successCount, errorCount))
	return successCount, errorCount
}

func (s *Scheduler) processLogEvents(events []db.LogEvent) {
	for _, event := range events {
		switch event.Level {
		case "info":
			args := convertFieldsToArgs(event.Fields)
			logger.Info(event.Message, args...)
		case "error":
			logger.Error(event.Error, event.Message, convertFieldsToArgs(event.Fields)...)
		case "success":
			args := convertFieldsToArgs(event.Fields)
			logger.Success(event.Message, args...)
		case "warn":
			args := convertFieldsToArgs(event.Fields)
			logger.Warn(event.Message, args...)
		case "task":
			if name, ok := event.Fields["rule"].(string); ok {
				logger.Task(name, event.Message)
			} else {
				logger.Task("Task", event.Message)
			}
		case "rule":
			if name, ok := event.Fields["rule"].(string); ok {
				logger.Rule(name, event.Message)
			} else {
				logger.Rule("Rule", event.Message)
			}
		case "db":
			if name, ok := event.Fields["server"].(string); ok {
				logger.DB(name, event.Message)
			} else {
				logger.DB("DB", event.Message)
			}
		}
	}
}

// Helper function for above processLogEvents
func convertFieldsToArgs(fields map[string]interface{}) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return args
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
		logger.Error(err, "failed to save execution record")
	}
}

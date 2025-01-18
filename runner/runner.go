package runner

import (
	"fmt"
	"log"
	"time"

	"github.com/nathanthorell/dataspy/config"
	"github.com/nathanthorell/dataspy/db"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	config    config.Config
	scheduler *cron.Cron
}

func NewScheduler(config config.Config) *Scheduler {
	return &Scheduler{
		config:    config,
		scheduler: cron.New(cron.WithSeconds()),
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
	rule, err := s.findRule(scheduleName)
	if err != nil {
		log.Printf("error finding rule: %v", err)
		return
	}

	server, err := s.findServer(rule.DbType)
	if err != nil {
		log.Printf("error finding server: %v", err)
		return
	}

	err = db.ExecuteRule(server, rule)
	if err != nil {
		log.Printf("error executing rule %s: %v", rule.Name, err)
		return
	}
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

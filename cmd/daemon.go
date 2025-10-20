package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nathanthorell/dataspy/runner"
	"github.com/nathanthorell/dataspy/storage"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run dataspy as a daemon with scheduled monitoring",
	Long:  `Start the scheduler to run rules on their configured cron schedules.`,
	Run:   runDaemon,
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}

func runDaemon(cmd *cobra.Command, args []string) {
	if err := loadEnv(); err != nil {
		log.Fatal(err)
	}

	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// bbolt storage
	store, err := storage.NewStore("data/dataspy.db")
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	sched := runner.NewScheduler(config, store)
	if err := sched.Start(); err != nil {
		log.Fatal(err)
	}

	// Keep the program running until terminated
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

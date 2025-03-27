package main

import (
	_ "embed"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"

	"github.com/joho/godotenv"
	"github.com/nathanthorell/dataspy/config"
	"github.com/nathanthorell/dataspy/runner"
	"github.com/nathanthorell/dataspy/storage"
)

//go:embed config/config.toml
var configTOML []byte

func main() {
	// Check if a file path argument is provided
	var envFile string
	if len(os.Args) > 1 {
		envFile = os.Args[1]
	}

	if envFile == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		envFile = filepath.Join(cwd, ".env")
	}

	if err := godotenv.Load(envFile); err != nil {
		log.Fatalf("Error loading environment variables from %s: %v", envFile, err)
	}

	// Load configuration
	config, err := config.LoadConfigBytes(configTOML)
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

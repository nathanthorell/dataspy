package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/nathanthorell/dataspy/config"
	"github.com/spf13/cobra"
)

var (
	envFile    string
	configData []byte
)

var rootCmd = &cobra.Command{
	Use:   "dataspy",
	Short: "A lightweight database monitoring tool",
	Long:  `dataspy checks your databases for business rule violations.`,
}

func Execute(embeddedConfig []byte) {
	configData = embeddedConfig
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&envFile, "env", "", "path to .env file (default: ./.env)")
}

// loadEnv loads environment variables from .env file
func loadEnv() error {
	if envFile == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		envFile = filepath.Join(cwd, ".env")
	}

	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("error loading environment variables from %s: %w", envFile, err)
	}
	return nil
}

// loadConfig loads the configuration from embedded TOML
func loadConfig() (config.Config, error) {
	cfg, err := config.LoadConfigBytes(configData)
	if err != nil {
		log.Fatal(err)
	}
	return cfg, nil
}

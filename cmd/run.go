package cmd

import (
	"log"

	"github.com/nathanthorell/dataspy/runner"
	"github.com/nathanthorell/dataspy/storage"
	"github.com/spf13/cobra"
)

var (
	ruleName string
	runAll   bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run rules on-demand",
	Long:  `Execute one or all rules immediately without waiting for scheduled execution.`,
	Run:   runRules,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&ruleName, "rule", "r", "", "name of the rule to run")
	runCmd.Flags().BoolVarP(&runAll, "all", "a", false, "run all rules")
	runCmd.MarkFlagsMutuallyExclusive("rule", "all")
}

func runRules(cmd *cobra.Command, args []string) {
	if !runAll && ruleName == "" {
		log.Fatal("must specify either --rule or --all")
	}

	if err := loadEnv(); err != nil {
		log.Fatal(err)
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// bbolt storage
	store, err := storage.NewStore("data/dataspy.db")
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	// Create scheduler (without starting it) to use execution methods
	sched := runner.NewScheduler(cfg, store)

	if runAll {
		sched.ExecuteAllRules()
	} else {
		if err := sched.ExecuteRuleByName(ruleName); err != nil {
			log.Fatal(err)
		}
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"judo-cli-module/internal/commands"
	"judo-cli-module/internal/config"
	"judo-cli-module/internal/help"
	"judo-cli-module/internal/session"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "judo",
		Short: "JUDO CLI",
		Long:  help.RootHelp(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config.LoadProperties()
			config.SetupEnvironment()
		},
	}

	// Common flags
	rootCmd.PersistentFlags().StringVarP(&config.Profile, "env", "e", "judo", "Use alternate environment")

	// Add commands
	rootCmd.AddCommand(
		commands.CreateDoctorCommand(),
		commands.CreateCleanCommand(),
		commands.CreatePruneCommand(),
		commands.CreateUpdateCommand(),
		commands.CreateGenerateCommand(),
		commands.CreateGenerateRootCommand(),
		commands.CreateDumpCommand(),
		commands.CreateImportCommand(),
		commands.CreateSchemaUpgradeCommand(),
		commands.CreateBuildCommand(),
		commands.CreateRecklessCommand(),
		commands.CreateStartCommand(),
		commands.CreateStopCommand(),
		commands.CreateStatusCommand(),
		commands.CreateInitCommand(),
		createSessionCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createSessionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "session",
		Short: "Start interactive JUDO CLI session",
		Long:  "Start an interactive session with command history, auto-completion, and persistent state",
		Run: func(cmd *cobra.Command, args []string) {
			session.StartInteractiveSession()
		},
	}
}

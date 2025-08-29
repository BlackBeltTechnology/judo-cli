package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"judo-cli-module/internal/commands"
	"judo-cli-module/internal/config"
	"judo-cli-module/internal/help"
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
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

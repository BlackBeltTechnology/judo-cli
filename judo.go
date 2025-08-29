package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"judo-cli-module/internal/commands"
	"judo-cli-module/internal/config"
)

var (
	version string
	rootCmd = &cobra.Command{
		Use:   "judo",
		Short: "judo-cli is a command line tool for managing JUDO applications.",
		Long:  `judo-cli is a command line tool for managing the lifecycle of JUDO applications.`,
		Version: version,
	}
)

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.LoadProperties, config.SetupEnvironment)

	rootCmd.AddCommand(commands.CreateGenerateCommand())
	rootCmd.AddCommand(commands.CreateBuildCommand())
	rootCmd.AddCommand(commands.CreateRecklessCommand())
	rootCmd.AddCommand(commands.CreateGenerateRootCommand())
	rootCmd.AddCommand(commands.CreateStatusCommand())
	rootCmd.AddCommand(commands.CreateDumpCommand())
	rootCmd.AddCommand(commands.CreateImportCommand())
	rootCmd.AddCommand(commands.CreateSchemaUpgradeCommand())
	rootCmd.AddCommand(commands.CreateCleanCommand())
	rootCmd.AddCommand(commands.CreatePruneCommand())
	rootCmd.AddCommand(commands.CreateUpdateCommand())
	rootCmd.AddCommand(commands.CreateStopCommand())
	rootCmd.AddCommand(commands.CreateStartCommand())
	rootCmd.AddCommand(commands.CreateInitCommand())
}

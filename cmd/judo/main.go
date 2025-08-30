package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"judo-cli-module/internal/commands"
	"judo-cli-module/internal/config"
	"judo-cli-module/internal/docker"
	"judo-cli-module/internal/help"
	"judo-cli-module/internal/session"
)

// Build information. Populated at build-time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Ensure Docker client is properly closed when the application exits
	defer docker.CloseDockerClient()

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
		commands.CreateLogCommand(),
		commands.CreateInitCommand(),
		createSessionCommand(),
		createVersionCommand(),
	)
	
	// Debug: print all commands
	fmt.Println("DEBUG: Registered commands:")
	for _, cmd := range rootCmd.Commands() {
		fmt.Printf("  %s - %s\n", cmd.Use, cmd.Short)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// createSessionCommand creates the session command to avoid circular import issues
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

// createVersionCommand creates the version command
func createVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print detailed version information including build details",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("JUDO CLI %s\n", version)
			fmt.Printf("  Version:    %s\n", version)
			fmt.Printf("  Git commit: %s\n", commit)
			fmt.Printf("  Built:      %s\n", date)
			fmt.Printf("  Built by:   %s\n", builtBy)
			fmt.Printf("  Go version: %s\n", "1.25.0")
		},
	}
}

/*
 * Copyright © 2026 BlackBelt Meta Zrt.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License 2.0 which is available at
 * https://www.eclipse.org/legal/epl-2.0/
 *
 * SPDX-License-Identifier: EPL-2.0
 */

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"judo-cli-module/internal/commands"
	"judo-cli-module/internal/config"
	"judo-cli-module/internal/docker"
	"judo-cli-module/internal/help"
	"judo-cli-module/internal/server"
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
		commands.CreateSelfUpdateCommand(version),
		createFrontendCommand(),
		createSessionCommand(),
		createServerCommand(),
		createVersionCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// createFrontendCommand creates the frontend build command
func createFrontendCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "frontend",
		Short: "Build frontend React application",
		Long:  "Build the React frontend application for the JUDO CLI web interface",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Building frontend React application...")

			// Change to frontend directory and run npm build
			frontendDir := "frontend"
			if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
				fmt.Printf("Frontend directory '%s' not found\n", frontendDir)
				os.Exit(1)
			}

			// Save current directory
			originalDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error getting current directory: %v\n", err)
				os.Exit(1)
			}

			// Change to frontend directory
			if err := os.Chdir(frontendDir); err != nil {
				fmt.Printf("Error changing to frontend directory: %v\n", err)
				os.Exit(1)
			}

			// Run npm build
			buildCmd := exec.Command("npm", "run", "build")
			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr

			if err := buildCmd.Run(); err != nil {
				fmt.Printf("Error building frontend: %v\n", err)
				// Change back to original directory before exiting
				os.Chdir(originalDir)
				os.Exit(1)
			}

			// Change back to original directory
			if err := os.Chdir(originalDir); err != nil {
				fmt.Printf("Error changing back to original directory: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Frontend build completed successfully!")
			fmt.Println("Built files are in: frontend/build/")
		},
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

// createServerCommand creates the server command
func createServerCommand() *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start JUDO CLI web server",
		Long:  "Start a web server with browser-based interface for JUDO CLI",
		Run: func(cmd *cobra.Command, args []string) {
			server := server.NewServer(port)
			if err := server.Start(); err != nil {
				fmt.Printf("Server error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 6969, "Port to run the server on")
	return cmd
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

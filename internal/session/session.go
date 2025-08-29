package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"judo-cli-module/internal/commands"
	"judo-cli-module/internal/config"
	"judo-cli-module/internal/help"
	"judo-cli-module/internal/utils"
)

type SessionState struct {
	CurrentDir     string
	ProjectInitialized bool
	CommandHistory []string
	StartTime      time.Time
	Verbose        bool
}

func StartInteractiveSession() {
	// Load previous session history
	previousHistory := loadSessionHistory()
	
	state := &SessionState{
		CurrentDir:     utils.GetCurrentDir(),
		ProjectInitialized: config.IsProjectInitialized(),
		StartTime:      time.Now(),
		CommandHistory: previousHistory,
	}

	// Display JUDO banner
	judobanner := "\u001B[0m\u001B[38;5;16m        \u001B[38;5;167m█\u001B[38;5;209m███\u001B[38;5;232m█\u001B[38;5;16m                                               \n\u001B[38;5;16m       \u001B[38;5;209m██████\u001B[38;5;16m                                               \n\u001B[38;5;16m       \u001B[38;5;238m█\u001B[38;5;209m████\u001B[38;5;131m█\u001B[38;5;16m                                               \n\u001B[38;5;16m       \u001B[38;5;232m█\u001B[38;5;242m█\u001B[38;5;239m█\u001B[38;5;238m█\u001B[38;5;241m█\u001B[38;5;238m█\u001B[38;5;240m█\u001B[38;5;242m████\u001B[38;5;16m     \u001B[38;5;238m█\u001B[38;5;242m████\u001B[38;5;233m█\u001B[38;5;242m███████\u001B[38;5;241m█\u001B[38;5;238m█\u001B[38;5;234m█\u001B[38;5;16m         \u001B[38;5;234m█\u001B[38;5;244m█\u001B[38;5;253m█\u001B[38;5;231m█\u001B[38;5;255m█\u001B[38;5;251m█\u001B[38;5;59m█\u001B[38;5;233m█\u001B[38;5;16m    \n\u001B[38;5;16m       \u001B[38;5;233m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;252m█\u001B[38;5;231m████\u001B[38;5;232m█\u001B[38;5;16m    \u001B[38;5;247m█\u001B[38;5;231m████\u001B[38;5;235m█\u001B[38;5;231m████████████\u001B[38;5;239m█\u001B[38;5;16m   \u001B[38;5;234m█\u001B[38;5;231m███████████\u001B[38;5;254m█\u001B[38;5;16m  \n\u001B[38;5;16m       \u001B[38;5;233m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;252m█\u001B[38;5;231m████\u001B[38;5;232m█\u001B[38;5;16m    \u001B[38;5;247m█\u001B[38;5;231m████\u001B[38;5;235m█\u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;234m█\u001B[38;5;235m█\u001B[38;5;244m█\u001B[38;5;231m█████\u001B[38;5;250m█\u001B[38;5;16m \u001B[38;5;242m█\u001B[38;5;231m█████\u001B[38;5;239m█\u001B[38;5;233m██\u001B[38;5;246m█\u001B[38;5;231m█████\u001B[38;5;16m \n\u001B[38;5;16m       \u001B[38;5;233m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;252m█\u001B[38;5;231m████\u001B[38;5;232m█\u001B[38;5;16m    \u001B[38;5;247m█\u001B[38;5;231m████\u001B[38;5;235m█\u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m    \u001B[38;5;255m█\u001B[38;5;231m████\u001B[38;5;16m \u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m      \u001B[38;5;231m████\u001B[38;5;255m█\n\u001B[38;5;16m       \u001B[38;5;233m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;252m█\u001B[38;5;231m████\u001B[38;5;232m█\u001B[38;5;16m    \u001B[38;5;247m█\u001B[38;5;231m████\u001B[38;5;235m█\u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m    \u001B[38;5;249m█\u001B[38;5;231m████\u001B[38;5;16m \u001B[38;5;231m████\u001B[38;5;251m█\u001B[38;5;16m      \u001B[38;5;255m█\u001B[38;5;231m████\n\u001B[38;5;16m       \u001B[38;5;233m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;252m█\u001B[38;5;231m████\u001B[38;5;232m█\u001B[38;5;16m    \u001B[38;5;247m█\u001B[38;5;231m████\u001B[38;5;235m█\u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m    \u001B[38;5;254m█\u001B[38;5;231m████\u001B[38;5;16m \u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m      \u001B[38;5;231m█████\n\u001B[38;5;255m█\u001B[38;5;231m███\u001B[38;5;255m█\u001B[38;5;16m  \u001B[38;5;243m█\u001B[38;5;231m████\u001B[38;5;241m█\u001B[38;5;243m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;16m   \u001B[38;5;232m█\u001B[38;5;231m█████\u001B[38;5;16m \u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m   \u001B[38;5;237m█\u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m \u001B[38;5;253m█\u001B[38;5;231m████\u001B[38;5;244m█\u001B[38;5;16m    \u001B[38;5;255m█\u001B[38;5;231m████\u001B[38;5;242m█\n\u001B[38;5;234m█\u001B[38;5;231m███████████\u001B[38;5;16m  \u001B[38;5;252m█\u001B[38;5;231m████████████\u001B[38;5;233m█\u001B[38;5;16m \u001B[38;5;231m████████████\u001B[38;5;255m█\u001B[38;5;16m   \u001B[38;5;250m█\u001B[38;5;231m████████████\u001B[38;5;242m█\u001B[38;5;16m \n\u001B[38;5;16m  \u001B[38;5;244m█\u001B[38;5;231m██████\u001B[38;5;255m█\u001B[38;5;233m█\u001B[38;5;16m     \u001B[38;5;249m█\u001B[38;5;231m████████\u001B[38;5;237m█\u001B[38;5;16m   \u001B[38;5;231m██████████\u001B[38;5;242m█\u001B[38;5;16m       \u001B[38;5;239m█\u001B[38;5;231m████████\u001B[38;5;236m█\u001B[38;5;16m   \n\u001B[0m"
	fmt.Print(judobanner)
	fmt.Println()
	
	fmt.Printf("\x1b[1;36m🚀 JUDO CLI Interactive Session\x1b[0m\n")
	fmt.Printf("\x1b[33mType 'help' for available commands, 'exit' to quit\x1b[0m\n\n")

	if state.ProjectInitialized {
		fmt.Printf("\x1b[32m✅ Project initialized in: %s\x1b[0m\n", state.CurrentDir)
	} else {
		fmt.Printf("\x1b[33m⚠️  No JUDO project found. Run 'init' to create one.\x1b[0m\n")
	}
	fmt.Println()

	// Create a root command for session mode
	rootCmd := createSessionRootCommand()

	// Create readline instance with tab completion
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "\x1b[1;34mjudo>\x1b[0m ",
		HistoryFile:     getHistoryFilePath(),
		AutoComplete:    getCompleter(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		fmt.Printf("\x1b[31mFailed to initialize readline: %v\x1b[0m\n", err)
		return
	}
	defer rl.Close()

	for {
		input, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				fmt.Println("\x1b[33mType 'exit' to quit or continue typing\x1b[0m")
				continue
			}
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Show command suggestions if input ends with '?' (legacy support)
		if strings.HasSuffix(input, "?") {
			partialInput := strings.TrimSuffix(input, "?")
			showEnhancedSuggestions(partialInput, state.CommandHistory)
			continue
		}

		// Add to command history
		state.CommandHistory = append(state.CommandHistory, input)

		// Handle session commands
		switch input {
		case "exit", "quit":
			fmt.Printf("\x1b[33m👋 Session duration: %s\x1b[0m\n", time.Since(state.StartTime).Round(time.Second))
			fmt.Printf("\x1b[33mCommands executed: %d\x1b[0m\n", len(state.CommandHistory))
			// Save session history before exiting
			saveSessionHistory(state.CommandHistory)
			return
		case "help":
			printSessionHelp()
			continue
		case "clear":
			fmt.Print("\033[H\033[2J")
			continue
		case "history":
			printCommandHistory(state.CommandHistory)
			continue
		case "status":
			updateSessionStatus(state)
			printSessionStatus(state)
			continue
		case "doctor":
			// Run doctor command in session mode
			doctorCmd := commands.CreateDoctorCommand()
			doctorCmd.SetArgs([]string{"--verbose"})
			if err := executeCommandInSession(doctorCmd, []string{}); err != nil {
				fmt.Printf("\x1b[31m❌ Doctor command failed: %v\x1b[0m\n", err)
			}
			continue
		}

		// Execute the command using cobra
		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}

		// Find and execute the command
		cmd, foundArgs, err := rootCmd.Find(args)
		if err != nil {
			fmt.Printf("\x1b[31mUnknown command: %s\x1b[0m\n", args[0])
			fmt.Printf("\x1b[33mType 'help' for available commands or '%s?' for suggestions\x1b[0m\n", args[0])
			continue
		}
		
		if cmd.Use == "session" {
			fmt.Printf("\x1b[33m⚠️  Warning: Cannot execute session command within session\x1b[0m\n")
			continue
		}

		// Show command execution feedback
		fmt.Printf("\x1b[36m⚡ Executing: %s\x1b[0m\n", input)
		startTime := time.Now()
		
		// Execute the command directly without going through the full Execute() flow
		// This avoids command parsing conflicts within the session context
		err = executeCommandInSession(cmd, foundArgs)
		if err != nil {
			fmt.Printf("\x1b[31m❌ Command failed after %s: %v\x1b[0m\n", time.Since(startTime).Round(time.Millisecond), err)
		} else {
			fmt.Printf("\x1b[32m✅ Command completed successfully in %s\x1b[0m\n", time.Since(startTime).Round(time.Millisecond))
		}

		// Update session state after command execution
		updateSessionStatus(state)
	}
}

func createSessionRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "judo",
		Short: "JUDO CLI",
		Long:  help.RootHelp(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config.LoadProperties()
			config.SetupEnvironment()
		},
	}

	// Add all commands
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
	)

	// Add session-specific flags
	rootCmd.PersistentFlags().StringVarP(&config.Profile, "env", "e", "judo", "Use alternate environment")

	return rootCmd
}

func printSessionHelp() {
	// Display JUDO banner in help
	judobanner := "\u001B[0m\u001B[38;5;16m        \u001B[38;5;167m█\u001B[38;5;209m███\u001B[38;5;232m█\u001B[38;5;16m                                               \n\u001B[38;5;16m       \u001B[38;5;209m██████\u001B[38;5;16m                                               \n\u001B[38;5;16m       \u001B[38;5;238m█\u001B[38;5;209m████\u001B[38;5;131m█\u001B[38;5;16m                                               \n\u001B[38;5;16m       \u001B[38;5;232m█\u001B[38;5;242m█\u001B[38;5;239m█\u001B[38;5;238m█\u001B[38;5;241m█\u001B[38;5;238m█\u001B[38;5;240m█\u001B[38;5;242m████\u001B[38;5;16m     \u001B[38;5;238m█\u001B[38;5;242m████\u001B[38;5;233m█\u001B[38;5;242m███████\u001B[38;5;241m█\u001B[38;5;238m█\u001B[38;5;234m█\u001B[38;5;16m         \u001B[38;5;234m█\u001B[38;5;244m█\u001B[38;5;253m█\u001B[38;5;231m█\u001B[38;5;255m█\u001B[38;5;251m█\u001B[38;5;59m█\u001B[38;5;233m█\u001B[38;5;16m    \n\u001B[38;5;16m       \u001B[38;5;233m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;252m█\u001B[38;5;231m████\u001B[38;5;232m█\u001B[38;5;16m    \u001B[38;5;247m█\u001B[38;5;231m████\u001B[38;5;235m█\u001B[38;5;231m████████████\u001B[38;5;239m█\u001B[38;5;16m   \u001B[38;5;234m█\u001B[38;5;231m███████████\u001B[38;5;254m█\u001B[38;5;16m  \n\u001B[38;5;16m       \u001B[38;5;233m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;252m█\u001B[38;5;231m████\u001B[38;5;232m█\u001B[38;5;16m    \u001B[38;5;247m█\u001B[38;5;231m████\u001B[38;5;235m█\u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;234m█\u001B[38;5;235m█\u001B[38;5;244m█\u001B[38;5;231m█████\u001B[38;5;250m█\u001B[38;5;16m \u001B[38;5;242m█\u001B[38;5;231m█████\u001B[38;5;239m█\u001B[38;5;233m██\u001B[38;5;246m█\u001B[38;5;231m█████\u001B[38;5;16m \n\u001B[38;5;16m       \u001B[38;5;233m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;252m█\u001B[38;5;231m████\u001B[38;5;232m█\u001B[38;5;16m    \u001B[38;5;247m█\u001B[38;5;231m████\u001B[38;5;235m█\u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m    \u001B[38;5;255m█\u001B[38;5;231m████\u001B[38;5;16m \u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m      \u001B[38;5;231m████\u001B[38;5;255m█\n\u001B[38;5;16m       \u001B[38;5;233m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;252m█\u001B[38;5;231m████\u001B[38;5;232m█\u001B[38;5;16m    \u001B[38;5;247m█\u001B[38;5;231m████\u001B[38;5;235m█\u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m    \u001B[38;5;249m█\u001B[38;5;231m████\u001B[38;5;16m \u001B[38;5;231m████\u001B[38;5;251m█\u001B[38;5;16m      \u001B[38;5;255m█\u001B[38;5;231m████\n\u001B[38;5;16m       \u001B[38;5;233m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;252m█\u001B[38;5;231m████\u001B[38;5;232m█\u001B[38;5;16m    \u001B[38;5;247m█\u001B[38;5;231m████\u001B[38;5;235m█\u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m    \u001B[38;5;254m█\u001B[38;5;231m████\u001B[38;5;16m \u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m      \u001B[38;5;231m█████\n\u001B[38;5;255m█\u001B[38;5;231m███\u001B[38;5;255m█\u001B[38;5;16m  \u001B[38;5;243m█\u001B[38;5;231m████\u001B[38;5;241m█\u001B[38;5;243m█\u001B[38;5;231m████\u001B[38;5;248m█\u001B[38;5;16m   \u001B[38;5;232m█\u001B[38;5;231m█████\u001B[38;5;16m \u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m   \u001B[38;5;237m█\u001B[38;5;231m████\u001B[38;5;255m█\u001B[38;5;16m \u001B[38;5;253m█\u001B[38;5;231m████\u001B[38;5;244m█\u001B[38;5;16m    \u001B[38;5;255m█\u001B[38;5;231m████\u001B[38;5;242m█\n\u001B[38;5;234m█\u001B[38;5;231m███████████\u001B[38;5;16m  \u001B[38;5;252m█\u001B[38;5;231m████████████\u001B[38;5;233m█\u001B[38;5;16m \u001B[38;5;231m████████████\u001B[38;5;255m█\u001B[38;5;16m   \u001B[38;5;250m█\u001B[38;5;231m████████████\u001B[38;5;242m█\u001B[38;5;16m \n\u001B[38;5;16m  \u001B[38;5;244m█\u001B[38;5;231m██████\u001B[38;5;255m█\u001B[38;5;233m█\u001B[38;5;16m     \u001B[38;5;249m█\u001B[38;5;231m████████\u001B[38;5;237m█\u001B[38;5;16m   \u001B[38;5;231m██████████\u001B[38;5;242m█\u001B[38;5;16m       \u001B[38;5;239m█\u001B[38;5;231m████████\u001B[38;5;236m█\u001B[38;5;16m   \n\u001B[0m"
	fmt.Print(judobanner)
	fmt.Println()
	
	fmt.Printf("\x1b[1;36m📋 JUDO CLI Session Commands:\x1b[0m\n")
	fmt.Printf("\x1b[32m  help\x1b[0m      - Show this help message\n")
	fmt.Printf("\x1b[32m  exit\x1b[0m      - Exit the interactive session\n")
	fmt.Printf("\x1b[32m  quit\x1b[0m      - Exit the interactive session\n")
	fmt.Printf("\x1b[32m  clear\x1b[0m     - Clear the terminal screen\n")
	fmt.Printf("\x1b[32m  history\x1b[0m   - Show command history\n")
	fmt.Printf("\x1b[32m  status\x1b[0m    - Show current session status\n")
	fmt.Printf("\x1b[32m  doctor\x1b[0m    - Run system health check\n")
	fmt.Println()
	fmt.Printf("\x1b[1;36m🔧 Project Commands:\x1b[0m\n")
	fmt.Printf("\x1b[32m  init\x1b[0m      - Initialize a new JUDO project\n")
	fmt.Printf("\x1b[32m  build\x1b[0m     - Build project\n")
	fmt.Printf("\x1b[32m  start\x1b[0m     - Start application\n")
	fmt.Printf("\x1b[32m  stop\x1b[0m      - Stop application\n")
	fmt.Printf("\x1b[32m  status\x1b[0m    - Show application status\n")
	fmt.Printf("\x1b[32m  clean\x1b[0m     - Clean project data\n")
	fmt.Printf("\x1b[32m  generate\x1b[0m  - Generate application from model\n")
	fmt.Printf("\x1b[32m  dump\x1b[0m      - Dump PostgreSQL database\n")
	fmt.Printf("\x1b[32m  import\x1b[0m    - Import PostgreSQL database dump\n")
	fmt.Printf("\x1b[32m  update\x1b[0m    - Update dependency versions\n")
	fmt.Printf("\x1b[32m  prune\x1b[0m     - Clean untracked files\n")
	fmt.Printf("\x1b[32m  reckless\x1b[0m  - Fast build & run mode\n")
	fmt.Println()
	fmt.Printf("\x1b[33m💡 Type any JUDO command directly to execute it\x1b[0m\n")
	fmt.Printf("\x1b[33m💡 Press TAB for auto-completion of commands and flags\x1b[0m\n")
	fmt.Printf("\x1b[33m💡 Type '<command>?' to see detailed suggestions\x1b[0m\n")
	fmt.Printf("\x1b[33m💡 Use 'history' to see previously executed commands\x1b[0m\n")
}

func printCommandHistory(history []string) {
	if len(history) == 0 {
		fmt.Printf("\x1b[33mNo commands in history\x1b[0m\n")
		return
	}

	fmt.Printf("\x1b[1;36m📜 Command History:\x1b[0m\n")
	for i, cmd := range history {
		fmt.Printf("\x1b[32m%3d\x1b[0m: %s\n", i+1, cmd)
	}
}

// getCommandSuggestions returns command suggestions for auto-completion
func getCommandSuggestions(input string) []string {
	commands := []string{
		"help", "exit", "quit", "clear", "history", "status", "doctor",
		"init", "build", "start", "stop", "clean", "prune", "update",
		"generate", "generate-root", "dump", "import", "schema-upgrade",
		"reckless",
	}

	var suggestions []string
	for _, cmd := range commands {
		if strings.HasPrefix(strings.ToLower(cmd), strings.ToLower(input)) {
			suggestions = append(suggestions, cmd)
		}
	}

	return suggestions
}

// getArgumentSuggestions returns argument suggestions for specific commands
func getArgumentSuggestions(commandLine string) []string {
	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		return []string{}
	}
	
	command := parts[0]
	
	switch command {
	case "build":
		return []string{
			"--build-parallel", "-p",
			"--build-app-module", "-a", 
			"--build-frontend-module", "-f",
			"--docker",
			"--skip-model",
			"--skip-backend", 
			"--skip-frontend",
			"--skip-karaf",
			"--skip-schema",
			"--build-schema-cli",
			"--version", "-v",
			"--maven-argument", "-m",
			"--quick", "-q",
			"--ignore-checksum", "-i",
		}
	case "start":
		return []string{
			"--skip-keycloak",
			"--skip-watch-bundles",
			"--options",
		}
	case "doctor":
		return []string{
			"--verbose", "-v",
		}
	case "init":
		return []string{
			"--group-id", "-g",
			"--model-name", "-n",
			"--type", "-t",
		}
	case "prune":
		return []string{
			"--frontend", "-f",
			"--yes", "-y",
		}
	case "generate", "generate-root", "update":
		return []string{
			"--ignore-checksum", "-i",
		}
	case "import":
		return []string{
			"--dump-name", "-n",
		}
	default:
		return []string{}
	}
}

// getHistoryBasedSuggestions returns suggestions based on command history
func getHistoryBasedSuggestions(input string, history []string) []string {
	var suggestions []string
	seen := make(map[string]bool)
	
	// Go through history in reverse order (most recent first)
	for i := len(history) - 1; i >= 0; i-- {
		cmd := history[i]
		if strings.HasPrefix(strings.ToLower(cmd), strings.ToLower(input)) && !seen[cmd] {
			suggestions = append(suggestions, cmd)
			seen[cmd] = true
			// Limit to 5 history-based suggestions
			if len(suggestions) >= 5 {
				break
			}
		}
	}
	
	return suggestions
}

// showCommandSuggestions displays command suggestions
func showCommandSuggestions(input string) {
	suggestions := getCommandSuggestions(input)
	if len(suggestions) > 0 {
		fmt.Printf("\n\x1b[36m💡 Command Suggestions:\x1b[0m\n")
		for _, suggestion := range suggestions {
			fmt.Printf("  \x1b[32m%s\x1b[0m\n", suggestion)
		}
		fmt.Println()
	} else {
		fmt.Printf("\x1b[33m❓ No matching commands found for '%s'\x1b[0m\n", input)
		fmt.Printf("\x1b[33m💡 Type 'help' to see all available commands\x1b[0m\n")
	}
}

// showEnhancedSuggestions displays both command and history-based suggestions
func showEnhancedSuggestions(input string, history []string) {
	cmdSuggestions := getCommandSuggestions(input)
	historySuggestions := getHistoryBasedSuggestions(input, history)
	argSuggestions := getArgumentSuggestions(input)
	
	if len(cmdSuggestions) > 0 || len(historySuggestions) > 0 || len(argSuggestions) > 0 {
		fmt.Printf("\n\x1b[36m💡 Suggestions for '%s':\x1b[0m\n", input)
		
		if len(cmdSuggestions) > 0 {
			fmt.Printf("\x1b[1;36m📝 Commands:\x1b[0m\n")
			for _, suggestion := range cmdSuggestions {
				fmt.Printf("  \x1b[32m%s\x1b[0m\n", suggestion)
			}
		}
		
		if len(argSuggestions) > 0 {
			fmt.Printf("\x1b[1;36m⚙️  Arguments:\x1b[0m\n")
			for _, suggestion := range argSuggestions {
				fmt.Printf("  \x1b[34m%s\x1b[0m\n", suggestion)
			}
		}
		
		if len(historySuggestions) > 0 {
			fmt.Printf("\x1b[1;36m📜 From History:\x1b[0m\n")
			for _, suggestion := range historySuggestions {
				fmt.Printf("  \x1b[33m%s\x1b[0m\n", suggestion)
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("\x1b[33m❓ No matching commands found for '%s'\x1b[0m\n", input)
		fmt.Printf("\x1b[33m💡 Type 'help' to see all available commands\x1b[0m\n")
	}
}

func updateSessionStatus(state *SessionState) {
	state.CurrentDir = utils.GetCurrentDir()
	state.ProjectInitialized = config.IsProjectInitialized()
}

// printSessionStatus displays detailed session information
func printSessionStatus(state *SessionState) {
	fmt.Printf("\x1b[1;36m📊 Session Status:\x1b[0m\n")
	fmt.Printf("\x1b[32m  Session Duration:\x1b[0m %s\n", time.Since(state.StartTime).Round(time.Second))
	fmt.Printf("\x1b[32m  Commands Executed:\x1b[0m %d\n", len(state.CommandHistory))
	fmt.Printf("\x1b[32m  Current Directory:\x1b[0m %s\n", state.CurrentDir)
	
	if state.ProjectInitialized {
		fmt.Printf("\x1b[32m  JUDO Project:\x1b[0m ✅ Initialized\n")
		
		// Show project info if available
		if cfg := config.GetConfig(); cfg != nil {
			fmt.Printf("\x1b[32m  App Name:\x1b[0m %s\n", cfg.AppName)
			fmt.Printf("\x1b[32m  Runtime:\x1b[0m %s\n", cfg.Runtime)
			fmt.Printf("\x1b[32m  Database:\x1b[0m %s\n", cfg.DBType)
		}
	} else {
		fmt.Printf("\x1b[33m  JUDO Project:\x1b[0m ⚠️  Not initialized (run 'init' to create)\n")
	}
	
	if len(state.CommandHistory) > 0 {
		lastCmd := state.CommandHistory[len(state.CommandHistory)-1]
		fmt.Printf("\x1b[32m  Last Command:\x1b[0m %s\n", lastCmd)
	}
	
	if state.Verbose {
		fmt.Printf("\x1b[32m  Verbose Mode:\x1b[0m Enabled\n")
	}
}

// GetCurrentDir returns the current working directory
func GetCurrentDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}

// getHistoryFilePath returns the path to the session history file
func getHistoryFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".judo", "session_history.json")
}

// loadSessionHistory loads command history from file
func loadSessionHistory() []string {
	filePath := getHistoryFilePath()
	if filePath == "" {
		return []string{}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return []string{}
	}

	var history []string
	if err := json.Unmarshal(data, &history); err != nil {
		return []string{}
	}

	// Keep only the last 100 commands
	if len(history) > 100 {
		history = history[len(history)-100:]
	}

	return history
}

// saveSessionHistory saves command history to file
func saveSessionHistory(history []string) {
	filePath := getHistoryFilePath()
	if filePath == "" {
		return
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	// Keep only the last 100 commands
	if len(history) > 100 {
		history = history[len(history)-100:]
	}

	data, err := json.Marshal(history)
	if err != nil {
		return
	}

	os.WriteFile(filePath, data, 0644)
}

// getCompleter returns a readline auto-completer for JUDO CLI commands
func getCompleter() readline.AutoCompleter {
	return readline.NewPrefixCompleter(
		readline.PcItem("help"),
		readline.PcItem("exit"),
		readline.PcItem("quit"),
		readline.PcItem("clear"),
		readline.PcItem("history"),
		readline.PcItem("status"),
		readline.PcItem("log",
			readline.PcItem("--tail", readline.PcItem("-t")),
			readline.PcItem("--follow", readline.PcItem("-f")),
			readline.PcItem("--lines", readline.PcItem("-n")),
		),
		readline.PcItem("doctor"),
		readline.PcItem("init"),
		readline.PcItem("build",
			readline.PcItem("--build-parallel", readline.PcItem("-p")),
			readline.PcItem("--build-app-module", readline.PcItem("-a")),
			readline.PcItem("--build-frontend-module", readline.PcItem("-f")),
			readline.PcItem("--docker"),
			readline.PcItem("--skip-model"),
			readline.PcItem("--skip-backend"),
			readline.PcItem("--skip-frontend"),
			readline.PcItem("--skip-karaf"),
			readline.PcItem("--skip-schema"),
			readline.PcItem("--build-schema-cli"),
			readline.PcItem("--version", readline.PcItem("-v")),
			readline.PcItem("--maven-argument", readline.PcItem("-m")),
			readline.PcItem("--quick", readline.PcItem("-q")),
			readline.PcItem("--ignore-checksum", readline.PcItem("-i")),
		),
		readline.PcItem("start",
			readline.PcItem("--skip-keycloak"),
			readline.PcItem("--skip-watch-bundles"),
			readline.PcItem("--options"),
		),
		readline.PcItem("stop"),
		readline.PcItem("clean"),
		readline.PcItem("prune",
			readline.PcItem("--frontend", readline.PcItem("-f")),
			readline.PcItem("--yes", readline.PcItem("-y")),
		),
		readline.PcItem("update",
			readline.PcItem("--ignore-checksum", readline.PcItem("-i")),
		),
		readline.PcItem("generate",
			readline.PcItem("--ignore-checksum", readline.PcItem("-i")),
		),
		readline.PcItem("generate-root",
			readline.PcItem("--ignore-checksum", readline.PcItem("-i")),
		),
		readline.PcItem("dump"),
		readline.PcItem("import",
			readline.PcItem("--dump-name", readline.PcItem("-n")),
		),
		readline.PcItem("schema-upgrade"),
		readline.PcItem("reckless"),
	)
}

// executeCommandInSession executes a command within the session context
// without going through the full cobra Execute() flow to avoid conflicts
func executeCommandInSession(cmd *cobra.Command, args []string) error {
	// Set up the command flags from the remaining arguments
	if err := cmd.ParseFlags(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	
	// Run the persistent pre-run hooks if they exist
	if cmd.PersistentPreRun != nil {
		cmd.PersistentPreRun(cmd, args)
	}
	if cmd.PreRun != nil {
		cmd.PreRun(cmd, args)
	}
	
	// Execute the main command function
	if cmd.RunE != nil {
		return cmd.RunE(cmd, args)
	} else if cmd.Run != nil {
		cmd.Run(cmd, args)
		return nil
	}
	
	return fmt.Errorf("command has no run function")
}
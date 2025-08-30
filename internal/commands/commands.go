package commands

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"judo-cli-module/internal/config"
	"judo-cli-module/internal/db"
	"judo-cli-module/internal/docker"
	"judo-cli-module/internal/help"
	"judo-cli-module/internal/karaf"
	"judo-cli-module/internal/utils"
)

func CreateGenerateCommand() *cobra.Command {
	var ignore bool
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate application based on model in JUDO project.",
		Long:  help.GenerateLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := config.GetConfig()

			// Stop mvnd daemon first (same as bash script)
			_ = utils.Run("mvnd", "--purge", "--stop")

			args := []string{
				"clean", "compile",
				"-DgenerateApplication",
				"-DskipApplicationBuild",
				"-f", cfg.ModelDir,
			}
			if ignore {
				args = append(args, "-DvalidateChecksum=false")
			}
			return utils.Run("mvnd", args...)
		},
	}
	cmd.Flags().BoolVarP(&ignore, "ignore-checksum", "i", false, "Ignore checksum errors and update checksums")
	return cmd
}

func CreateBuildCommand() *cobra.Command {
	// sensible defaults (match bash)
	config.Options.BuildModel = true
	config.Options.BuildBackend = true
	config.Options.BuildFrontend = true
	config.Options.BuildKaraf = true
	config.Options.SchemaBuilding = true
	config.Options.SchemaCliBuilding = false
	config.Options.DockerBuilding = false

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build project",
		Long:  help.BuildLongHelp(),
		Run:   runBuild,
	}

	cmd.Flags().BoolVarP(&config.Options.BuildParallel, "build-parallel", "p", false, "Parallel maven build")
	cmd.Flags().BoolVarP(&config.Options.BuildAppModule, "build-app-module", "a", false, "Build app module only")
	cmd.Flags().BoolVarP(&config.Options.BuildFrontend, "build-frontend-module", "f", false, "Build frontend module only")
	cmd.Flags().BoolVar(&config.Options.DockerBuilding, "docker", false, "Build Docker images")
	// skip-* flags flip the respective build toggles
	cmd.Flags().Bool("skip-model", false, "Skip model building")
	cmd.Flags().Bool("skip-backend", false, "Skip backend building")
	cmd.Flags().Bool("skip-frontend", false, "Skip frontend building")
	cmd.Flags().Bool("skip-karaf", false, "Skip Backend Karaf building")
	cmd.Flags().Bool("skip-schema", false, "Skip building schema migration image")
	cmd.Flags().Bool("build-schema-cli", false, "Build schema CLI standalone JAR file")
	cmd.Flags().StringVarP(&config.Options.VersionNumber, "version", "v", "SNAPSHOT", "Version number")
	cmd.Flags().StringVarP(&config.Options.ExtraMavenArgs, "maven-argument", "m", "", "Extra Maven args (quoted)")
	cmd.Flags().BoolVarP(&config.Options.QuickMode, "quick", "q", false, "Quick mode: cache + skip validations")
	cmd.Flags().BoolVarP(&config.Options.IgnoreChecksum, "ignore-checksum", "i", false, "Ignore checksum errors")

	return cmd
}

func CreateRecklessCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reckless",
		Short: "Build & run fast (skips validations, favors speed)",
		Long:  help.RecklessLongHelp(),
		Run: func(cmd *cobra.Command, _ []string) {
			config.Options.Reckless = true
			config.Options.QuickMode = true
			config.Options.BuildKaraf = false // match bash defaults for reckless build matrix
			runBuild(cmd, nil)
		},
	}
}

func CreateGenerateRootCommand() *cobra.Command {
	var ignore bool
	cmd := &cobra.Command{
		Use:   "generate-root",
		Short: "Generate application root structure based on model in JUDO project.",
		Long:  help.GenerateRootLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := config.GetConfig()

			_ = utils.Run("mvnd", "--purge", "--stop")

			args := []string{
				"clean", "compile",
				"-DgenerateRoot",
				"-DskipApplicationBuild",
				"-f", cfg.ModelDir,
				"-U",
			}
			if ignore {
				args = append(args, "-DvalidateChecksum=false")
			}
			return utils.Run("mvnd", args...)
		},
	}
	cmd.Flags().BoolVarP(&ignore, "ignore-checksum", "i", false, "Ignore checksum errors and update checksums")
	return cmd
}

func CreateStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print status of Karaf/Keycloak/PostgreSQL containers and resources",
		Long:  help.StatusLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			// Check if JUDO project is initialized
			if err := requireJudoProject(); err != nil {
				return err
			}
			cfg := config.GetConfig()
			fmt.Println("Runtime:", cfg.Runtime, " DB:", cfg.DBType)
			if cfg.Runtime == "karaf" {
				karafDir := filepath.Join(cfg.ModelDir, "application", ".karaf")
				// Karaf
				if karaf.KarafRunning(karafDir) {
					fmt.Println("Karaf is running")
				} else {
					fmt.Println("Karaf is not running")
				}

				// Postgres (if applicable)
				if cfg.DBType == "postgresql" {
					pgName := "postgres-" + cfg.SchemaName
					if docker.DockerInstanceRunning(pgName) {
						fmt.Println("PostgreSQL is running")
					} else {
						fmt.Println("PostgreSQL is not running")
						if docker.ContainerExists(pgName) {
							fmt.Println("PostgreSQL container exists")
						} else {
							fmt.Println("PostgreSQL container does not exist")
						}
						if docker.DockerVolumeExists(cfg.AppName + "_postgresql_db") {
							fmt.Println("PostgreSQL db volume exists")
						} else {
							fmt.Println("PostgreSQL db volume does not exist")
						}
						if docker.DockerVolumeExists(cfg.AppName + "_postgresql_data") {
							fmt.Println("PostgreSQL data volume exists")
						} else {
							fmt.Println("PostgreSQL data volume does not exist")
						}
					}
				}

				// Keycloak
				kcName := "keycloak-" + cfg.KeycloakName
				if docker.DockerInstanceRunning(kcName) {
					fmt.Println("Keycloak is running")
				} else {
					fmt.Println("Keycloak is not running")
					if docker.ContainerExists(kcName) {
						fmt.Println("Keycloak container exists")
					} else {
						fmt.Println("Keycloak container does not exist")
					}
				}
			}
			return nil
		},
	}
	return cmd
}

func CreateDumpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Dump PostgreSQL DB data (creates <schema>_dump_YYYYMMDD_HHMMSS.tar.gz).",
		Long:  help.DumpLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			// Check if JUDO project is initialized
			if err := requireJudoProject(); err != nil {
				return err
			}
			cfg := config.GetConfig()

			if cfg.DBType != "postgresql" {
				fmt.Println("Dump is only supported with PostgreSQL.")
				return nil
			}

			// Ensure DB is up, then dump, then stop it (like the bash script)
			docker.StartPostgres()
			name := "postgres-" + cfg.SchemaName
			file, err := db.DumpPostgresql(name, cfg.SchemaName)
			if err != nil {
				return err
			}
			fmt.Println("Database dumped to", file)
			_ = docker.StopDockerInstance(name)
			return nil
		},
	}
	return cmd
}

func CreateImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import PostgreSQL DB dump (pg_restore).",
		Long:  help.ImportLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			// Check if JUDO project is initialized
			if err := requireJudoProject(); err != nil {
				return err
			}
			cfg := config.GetConfig()

			if cfg.DBType != "postgresql" {
				fmt.Println("Import is only supported with PostgreSQL.")
				return nil
			}

			instance := "postgres-" + cfg.SchemaName
			// Fresh db state
			_ = docker.RemoveDockerInstance(instance)
			_ = docker.RemoveDockerVolume(cfg.SchemaName + "_postgresql_db")
			_ = docker.RemoveDockerVolume(cfg.SchemaName + "_postgresql_data")

			// Start DB and wait
			docker.StartPostgres()

			// Determine dump file
			dumpFile := config.Options.DumpName
			if strings.TrimSpace(dumpFile) == "" {
				var err error
				dumpFile, err = db.FindLatestDump(cfg.SchemaName)
				if err != nil {
					return err
				}
			}
			fmt.Println("Loading dump:", dumpFile)

			// Run pg_restore inside the container
			if err := db.ImportPostgresql(instance, cfg.SchemaName, dumpFile); err != nil {
				return err
			}

			// Bounce container (same as bash)
			_ = docker.StopDockerInstance(instance)
			docker.StartPostgres()
			return nil
		},
	}
	// Bash used -dn / --dump-name; we expose -n/--dump-name here.
	cmd.Flags().StringVarP(&config.Options.DumpName, "dump-name", "n", "", "Dump filename to import (defaults to latest <schema>_dump_*.tar.gz)")
	return cmd
}

func CreateSchemaUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema-upgrade",
		Short: "Apply RDBMS schema upgrade using current running database (PostgreSQL only).",
		Long:  help.SchemaUpgradeLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			// Check if JUDO project is initialized
			if err := requireJudoProject(); err != nil {
				return err
			}
			cfg := config.GetConfig()

			if cfg.DBType != "postgresql" {
				fmt.Println("Schema upgrade requires PostgreSQL.")
				return nil
			}

			// Ensure Postgres is started and reachable
			docker.StartPostgres()

			updateModel := filepath.Join(cfg.ModelDir, "model", "target", "generated-resources", "model",
				fmt.Sprintf("%s-rdbms_postgresql.model", cfg.SchemaName))
			schemaDir := filepath.Join(cfg.ModelDir, "schema")

			args := []string{
				"judo-rdbms-schema:apply",
				fmt.Sprintf("-DjdbcUrl=jdbc:postgresql://127.0.0.1:%d/%s", cfg.PostgresPort, cfg.SchemaName),
				"-DdbType=postgresql",
				"-DdbUser=" + cfg.SchemaName,
				"-DdbPassword=" + cfg.SchemaName,
				"-DschemaIgnoreModelDependency=true",
				"-DupdateModel=" + updateModel,
				"-f", schemaDir,
			}
			return utils.Run("mvnd", args...)
		},
	}
	return cmd
}

func CreateCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Stop postgresql docker container and clear data.",
		Long:  help.CleanLongHelp(),
		RunE: func(_ *cobra.Command, args []string) error {
			// Check if JUDO project is initialized
			if err := requireJudoProject(); err != nil {
				return err
			}
			cfg := config.GetConfig()
			for _, env := range docker.GetComposeEnvs(cfg) {
				_ = docker.StopCompose(cfg, env)
			}
			_ = docker.RemoveDockerInstance("postgres-" + cfg.SchemaName)
			_ = docker.RemoveDockerInstance("keycloak-" + cfg.KeycloakName)
			_ = docker.RemoveDockerNetwork(cfg.AppName)
			_ = docker.RemoveDockerVolume(cfg.AppName + "_certs")
			_ = docker.RemoveDockerVolume(cfg.SchemaName + "_postgresql_db")
			_ = docker.RemoveDockerVolume(cfg.SchemaName + "_postgresql_data")
			_ = docker.RemoveDockerVolume(cfg.AppName + "_filestore")
			if cfg.Runtime == "karaf" {
				karaf.StopKaraf(cfg.KarafDir)
				if err := os.RemoveAll(cfg.KarafDir); err != nil {
					// Return the error so the user sees it.
					return fmt.Errorf("failed to remove karaf directory: %w", err)
				}
			}
			return nil
		},
	}
	return cmd
}

func CreatePruneCommand() *cobra.Command {
	var frontend bool
	var yes bool
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Stop postgresql docker container and delete untracked files in this repository.",
		Long:  help.PruneLongHelp(),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if JUDO project is initialized
			if err := requireJudoProject(); err != nil {
				return err
			}
			cfg := config.GetConfig()
			st := &config.State{PruneFrontend: frontend, PruneConfirm: !yes}
			pruneApplication(cfg, st)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&frontend, "frontend", "f", false, "Clear only frontend data")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation")
	return cmd
}

func CreateUpdateCommand() *cobra.Command {
	var ignore bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update dependency versions in JUDO project.",
		Long:  help.UpdateLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := config.GetConfig()

			// Run SDKMAN steps (Unix or via WSL on Windows). Safe to skip if unavailable.
			_ = utils.SdkmanRun(
				"sdk selfupdate || true",
				"sdk env install || true",
				"sdk env || true",
			)

			// Stop mvnd daemon like the bash script
			_ = utils.Run("mvnd", "--purge", "--stop")

			mvnargs := []string{
				"clean", "compile",
				"-DgenerateRoot",
				"-DskipApplicationBuild",
				"-DupdateJudoVersions=true",
				"-f", cfg.ModelDir,
				"-U",
			}
			if ignore {
				mvnargs = append(mvnargs, "-DvalidateChecksum=false")
			}
			return utils.Run("mvnd", mvnargs...)
		},
	}
	cmd.Flags().BoolVarP(&ignore, "ignore-checksum", "i", false, "Ignore checksum errors and update checksums")
	return cmd
}

func CreateStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop application, postgresql and keycloak (if running)",
		Long:  help.StopLongHelp(),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if JUDO project is initialized
			if err := requireJudoProject(); err != nil {
				return err
			}
			cfg := config.GetConfig()
			if cfg.Runtime == "karaf" {
				karaf.StopKaraf(cfg.KarafDir)
				if cfg.DBType == "postgresql" {
					_ = docker.StopDockerInstance("postgres-" + cfg.SchemaName)
				}
				_ = docker.StopDockerInstance("keycloak-" + cfg.KeycloakName)
			}
			return nil
		},
	}
	return cmd
}

func runBuild(cmd *cobra.Command, args []string) {
	// Check if JUDO project is initialized
	if err := requireJudoProject(); err != nil {
		log.Fatal(err)
	}

	// reflect skip-* flags into options
	if v, _ := cmd.Flags().GetBool("skip-model"); v {
		config.Options.BuildModel = false
	}
	if v, _ := cmd.Flags().GetBool("skip-backend"); v {
		config.Options.BuildBackend = false
	}
	if v, _ := cmd.Flags().GetBool("skip-frontend"); v {
		config.Options.BuildFrontend = false
	}
	if v, _ := cmd.Flags().GetBool("skip-karaf"); v {
		config.Options.BuildKaraf = false
	}
	if v, _ := cmd.Flags().GetBool("skip-schema"); v {
		config.Options.SchemaBuilding = false
	}
	if v, _ := cmd.Flags().GetBool("build-schema-cli"); v {
		config.Options.SchemaCliBuilding = true
	}

	cfg := config.GetConfig()
	if config.Options.Reckless {
		// mirror bash: start local env first
		startLocalEnvironment()
	}

	// stop mvnd daemon as in bash (except reckless path which may run fast)
	if !config.Options.Reckless {
		_ = utils.Run("mvnd", "--purge", "--stop")
	}

	goal := "install"
	if config.Options.Reckless {
		goal = "package"
	}

	// base args
	buildArgs := []string{goal}
	if !config.Options.Reckless {
		buildArgs = append([]string{"clean"}, buildArgs...)
	}
	buildArgs = append(buildArgs, "-Dsmartbuilder.profiling=true")

	// version handling (-Drevision) when not SNAPSHOT
	if strings.TrimSpace(config.Options.VersionNumber) != "" && strings.ToUpper(strings.TrimSpace(config.Options.VersionNumber)) != "SNAPSHOT" {
		buildArgs = append(buildArgs, "-Drevision="+config.Options.VersionNumber)
	}

	if config.Options.IgnoreChecksum {
		buildArgs = append(buildArgs, "-DvalidateChecksum=false")
	}
	if config.Options.QuickMode {
		buildArgs = append(buildArgs,
			"-Dfrontend-build-type=quick",
			"-DvalidateModels=false",
			"-DuseCache=true",
			"-DskipPrepareNodeJS",
		)
	}
	// parallel? keep conservative default (one core per thread)
	if config.Options.BuildParallel {
		buildArgs = append(buildArgs, "-T", "1C")
	}

	// Apply component toggles
	if !config.Options.BuildFrontend {
		buildArgs = append(buildArgs, "-DskipReact", "-DskipFrontendModel", "-DskipPrepareNodeJS")
	}
	if !config.Options.BuildModel {
		buildArgs = append(buildArgs, "-DskipModels")
	}
	if !config.Options.BuildKaraf {
		buildArgs = append(buildArgs, "-DskipKaraf")
	}
	if !config.Options.DockerBuilding {
		buildArgs = append(buildArgs, "-DskipDocker", "-DskipSchemaDocker", "-DkarafOfflineZip=false")
	}
	if !config.Options.SchemaBuilding {
		buildArgs = append(buildArgs, "-DskipSchema")
	}
	if !config.Options.SchemaCliBuilding {
		buildArgs = append(buildArgs, "-DskipSchemaCli")
	}

	// extra user maven args (best-effort split)
	if s := strings.TrimSpace(config.Options.ExtraMavenArgs); s != "" {
		buildArgs = append(buildArgs, strings.Fields(s)...)
	}

	// Special target layouts (subset builds)
	switch {
	case config.Options.BuildBackend && config.Options.BuildAppModule:
		// build backend app module (+interceptors) only
		fmt.Println("Building backend app module only...")
		args := append([]string{}, buildArgs...)
		args = append(args, "-f", cfg.AppDir, "-pl", "app,interceptors", "-DskipModels=true")
		utils.CheckError(utils.Run("mvnd", args...))
		return

	case !config.Options.BuildBackend && config.Options.BuildFrontend:
		// frontend only
		fmt.Println("Building frontend only...")
		args := append([]string{}, buildArgs...)
		args = append(args, "-f", filepath.Join(cfg.AppDir, "frontend-react"))
		utils.CheckError(utils.Run("mvnd", args...))
		return

	default:
		// full (or mostly-full) build, starting at MODEL_DIR
		args := append([]string{}, buildArgs...)
		args = append(args, "-f", cfg.ModelDir)
		utils.CheckError(utils.Run("mvnd", args...))
	}

	// Reckless extras: optionally (light) post-steps
	if config.Options.Reckless {
		// Skipping schema upgrade + bundle hot-install here to keep it simple and stable.
		fmt.Println("Reckless build completed.")
	}
}

func CreateStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start application",
		Long:  help.StartLongHelp(),
		Run:   runStart,
	}
	cmd.Flags().Bool("skip-keycloak", false, "Skip starting Keycloak")
	cmd.Flags().Bool("skip-watch-bundles", false, "Disable watching of bundle changes")
	cmd.Flags().String("options", "", "Additional options: key=value,key2=value2 (e.g. runtime=compose,dbtype=postgresql,karaf_port=8181)")
	return cmd
}

func runStart(cmd *cobra.Command, _ []string) {
	// Check if JUDO project is initialized
	if err := requireJudoProject(); err != nil {
		log.Fatal(err)
	}

	cfg := config.GetConfig()
	// Pre-flight checks
	if !docker.IsDockerRunning() {
		log.Fatal("Docker daemon is not running. Please start Docker and try again.")
	}

	// Set default values
	config.Options.StartKeycloak = true
	config.Options.WatchBundles = true
	config.Options.StartKaraf = true

	// apply flags
	if v, _ := cmd.Flags().GetBool("skip-keycloak"); v {
		config.Options.StartKeycloak = false
	}
	if v, _ := cmd.Flags().GetBool("skip-watch-bundles"); v {
		config.Options.WatchBundles = false
	}

	// parse -o/--options: key=value,key2=value2
	if raw, _ := cmd.Flags().GetString("options"); strings.TrimSpace(raw) != "" {
		config.ApplyInlineOptions(raw)
	}

	// Port checks with warnings instead of errors
	if config.Options.StartKeycloak {
		if !utils.IsPortAvailable(cfg.KeycloakPort) {
			// Check if this is our own Keycloak instance using the port
			if docker.IsPortUsedByKeycloak(cfg.KeycloakPort) {
				fmt.Printf("\x1b[33m⚠️  Keycloak port %d is already in use by your running JUDO Keycloak instance. Skipping Keycloak start.\x1b[0m\n", cfg.KeycloakPort)
				config.Options.StartKeycloak = false // Skip Keycloak start
			} else {
				log.Fatalf("Keycloak port %d is already in use by another process.", cfg.KeycloakPort)
			}
		}
	}
	if cfg.DBType == "postgresql" {
		if !utils.IsPortAvailable(cfg.PostgresPort) {
			// Check if this is our own PostgreSQL instance using the port
			if docker.IsPortUsedByPostgres(cfg.PostgresPort) {
				fmt.Printf("\x1b[33m⚠️  PostgreSQL port %d is already in use by your running JUDO PostgreSQL instance. Skipping PostgreSQL start.\x1b[0m\n", cfg.PostgresPort)
				// We'll skip PostgreSQL start by not calling docker.StartPostgres() later
			} else {
				log.Fatalf("PostgreSQL port %d is already in use by another process.", cfg.PostgresPort)
			}
		}
	}

	runtime := cfg.Runtime
	if runtime != "compose" && runtime != "karaf" {
		fmt.Println("Unknown runtime:", runtime, " — defaulting to karaf")
		runtime = "karaf"
	}

	if runtime == "karaf" {
		// Karaf-specific checks
		if !utils.IsPortAvailable(cfg.KarafPort) {
			// Check if this is our own Karaf instance using the port
			karafDir := filepath.Join(cfg.ModelDir, "application", ".karaf")
			if utils.IsPortUsedByKaraf(cfg.KarafPort, karafDir) {
				fmt.Printf("\x1b[33m⚠️  Karaf port %d is already in use by your running JUDO application. Skipping Karaf start.\x1b[0m\n", cfg.KarafPort)
				// Skip Karaf start by not calling karaf.StartKaraf() later
				config.Options.StartKaraf = false
				config.Options.WatchBundles = false // Also disable bundle watching
			} else {
				log.Fatalf("Karaf port %d is already in use by another process.", cfg.KarafPort)
			}
		}
		ver := utils.GetProjectVersion()
		tarPath := filepath.Join(cfg.ModelDir, "application", "karaf-offline", "target",
			fmt.Sprintf("%s-application-karaf-offline-%s.tar.gz", cfg.AppName, ver),
		)
		if _, err := os.Stat(tarPath); os.IsNotExist(err) {
			log.Fatalf("Karaf archive not found at %s. Please run a build first.", tarPath)
		}
	}

	// Execution
	switch runtime {
	case "compose":
		docker.StartCompose()
	case "karaf":
		startLocalEnvironment()
	}
}

func startLocalEnvironment() {
	cfg := config.GetConfig()
	if cfg.DBType == "postgresql" {
		docker.StartPostgres()
	}

	if config.Options.StartKeycloak {
		docker.StartKeycloak()
	}

	if config.Options.StartKaraf {
		karaf.StartKaraf()
	}
}

func pruneApplication(cfg *config.Config, st *config.State) {
	canContinue := "Y"
	if st.PruneConfirm {
		location := "this repository"
		if st.PruneFrontend {
			location = "application/frontend-react"
		}
		fmt.Printf("Prune command deletes all untracked files in %s!\nAre you sure you want to continue? [Y/n]: ", location)
		scanner := utils.NewScanner(os.Stdin)
		if scanner.Scan() {
			canContinue = strings.TrimSpace(scanner.Text())
		}
	}
	if strings.ToUpper(canContinue) != "Y" {
		println("Aborting prune.")
		os.Exit(13)
	}

	if st.PruneFrontend {
		_ = utils.Run("git", "clean", "-dffx", filepath.Join(cfg.AppDir, "frontend-react"))
		return
	}

	if cfg.DBType == "postgresql" {
		_ = docker.StopDockerInstance("postgres-" + cfg.SchemaName)
	}
	_ = docker.StopDockerInstance("keycloak-" + cfg.KeycloakName)
	if cfg.Runtime == "karaf" {
		karaf.StopKaraf(cfg.KarafDir)
	}
	_ = utils.Run("git", "clean", "-dffx", cfg.ModelDir)
}

func CreateDoctorCommand() *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system health and required dependencies",
		Long:  help.DoctorLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			return runDoctor(verbose)
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output for all checks")
	return cmd
}

// doctorMessage prints colorized messages with emojis for the doctor command
func doctorMessage(emoji, color, message string) {
	colorCode := ""
	switch color {
	case "green":
		colorCode = "\x1b[1;32m" // Bold green
	case "red":
		colorCode = "\x1b[1;31m" // Bold red
	case "yellow":
		colorCode = "\x1b[1;33m" // Bold yellow
	case "blue":
		colorCode = "\x1b[1;34m" // Bold blue
	case "cyan":
		colorCode = "\x1b[1;36m" // Bold cyan
	default:
		colorCode = "\x1b[0m" // Reset
	}
	fmt.Printf("%s%s %s\x1b[0m\n", emoji, colorCode, message)
}

func runDoctor(verbose bool) error {
	doctorMessage("🩺", "cyan", "JUDO CLI Doctor - Checking system health...")
	fmt.Println()

	allPassed := true

	// Check Go
	//	if checkGo(verbose) {
	//		fmt.Printf("✅ Go: Available\n")
	//	} else {
	//		fmt.Printf("❌ Go: Not found\n")
	//		allPassed = false
	//	}

	// Check Docker
	if checkDocker(verbose) {
		doctorMessage("✅", "green", "Docker: Available and running")
	} else {
		doctorMessage("❌", "red", "Docker: Not available or not running")
		allPassed = false
	}

	// Check Maven
	if checkMaven(verbose) {
		doctorMessage("✅", "green", "Maven: Available")
	} else {
		doctorMessage("❌", "red", "Maven: Not found")
		allPassed = false
	}

	// Check Git
	if checkGit(verbose) {
		doctorMessage("✅", "green", "Git: Available")
	} else {
		doctorMessage("❌", "red", "Git: Not found")
		allPassed = false
	}

	// Check Java
	if checkJava(verbose) {
		doctorMessage("✅", "green", "Java: Available")
	} else {
		doctorMessage("⚠️ ", "yellow", "Java: Not found (optional for some operations)")
	}

	// Check Maven Daemon (mvnd)
	if checkMavenDaemon(verbose) {
		doctorMessage("✅", "green", "Maven Daemon (mvnd): Available")
	} else {
		doctorMessage("⚠️ ", "yellow", "Maven Daemon (mvnd): Not found")
	}

	// Check SDKMAN and install if missing
	sdkmanAvailable := checkSDKMAN(verbose)
	if sdkmanAvailable {
		doctorMessage("✅", "green", "SDKMAN: Available")
	} else {
		doctorMessage("⚠️ ", "yellow", "SDKMAN: Not found - installing now...")

		// Always install SDKMAN automatically (no prompt)
		fmt.Printf("   \x1b[33mInstalling SDKMAN...\x1b[0m\n")
		if err := utils.InstallSDKMAN(); err != nil {
			fmt.Printf("   \x1b[31m❌ Failed to install SDKMAN: %v\x1b[0m\n", err)
			if verbose {
				if utils.HaveWSL() {
					fmt.Printf("   \x1b[33mOn Windows, WSL is required for SDKMAN installation\x1b[0m\n")
				}
			}
		} else {
			fmt.Printf("   \x1b[32m✅ SDKMAN installed successfully\x1b[0m\n")
			sdkmanAvailable = true
			// SDKMAN is now available, so mark this check as passed
			allPassed = true
		}
	}

	// Check ports
	fmt.Println()
	doctorMessage("🔌", "blue", "Port availability checks:")
	checkPortAvailability(8080, "Keycloak (default)", verbose)
	checkPortAvailability(8181, "Karaf (default)", verbose)
	checkPortAvailability(5432, "PostgreSQL (default)", verbose)

	// Check if this is a JUDO project directory
	fmt.Println()
	isProjectInitialized := config.IsProjectInitialized()
	if isProjectInitialized {
		doctorMessage("✅", "green", "JUDO Project: Initialized")

		// If SDKMAN is available and we're in a JUDO project, install required tools
		if sdkmanAvailable {
			fmt.Println()
			fmt.Printf("\x1b[1;33m🔧 Installing required development tools via SDKMAN...\x1b[0m\n")
			if err := utils.InstallRequiredTools(); err != nil {
				fmt.Printf("\x1b[33m⚠️  Failed to install required tools: %v\x1b[0m\n", err)
				if verbose {
					fmt.Printf("   \x1b[33mYou can manually install tools using 'sdk install maven' and 'sdk install java'\x1b[0m\n")
				}
			} else {
				fmt.Printf("\x1b[32m✅ Development tools installed successfully\x1b[0m\n")
			}
		} else if verbose {
			fmt.Printf("   \x1b[33mSDKMAN not available - cannot auto-install Maven/Java\x1b[0m\n")
		}
	} else {
		doctorMessage("ℹ️ ", "blue", "JUDO Project: Not initialized in this directory")
		if verbose {
			fmt.Printf("   \x1b[36mRun 'judo init' to initialize a new JUDO project\x1b[0m\n")
		}
	}

	fmt.Println()
	if allPassed {
		doctorMessage("🎉", "green", "All essential tools are available! JUDO CLI should work properly.")
	} else {
		doctorMessage("🚨", "red", "Some essential tools are missing. Please install them before using JUDO CLI.")
		return fmt.Errorf("system health check failed")
	}

	return nil
}

func checkGo(verbose bool) bool {
	version, err := utils.RunCapture("go", "version")
	if verbose && err == nil {
		fmt.Printf("   Go version: %s\n", version)
	}
	return err == nil
}

func checkDocker(verbose bool) bool {
	// Check if docker command exists
	version, err := utils.RunCapture("docker", "--version")
	if err != nil {
		return false
	}
	if verbose {
		fmt.Printf("   Docker version: %s\n", version)
	}

	// Check if docker daemon is running
	_, err = utils.RunCapture("docker", "info")
	if err != nil {
		if verbose {
			fmt.Printf("   Docker daemon status: Not running\n")
		}
		return false
	}
	if verbose {
		fmt.Printf("   Docker daemon status: Running\n")
	}
	return true
}

func checkMaven(verbose bool) bool {
	// Check mvnd first (preferred)
	version, err := utils.RunCapture("mvnd", "--version")
	if err == nil {
		if verbose {
			fmt.Printf("   Maven Daemon (mvnd) version: %s\n", strings.Split(version, "\n")[0])
		}
		return true
	}

	// Fallback to mvn
	version, err = utils.RunCapture("mvn", "--version")
	if err == nil {
		if verbose {
			fmt.Printf("   Maven version: %s\n", strings.Split(version, "\n")[0])
		}
		return true
	}
	return false
}

func checkGit(verbose bool) bool {
	version, err := utils.RunCapture("git", "--version")
	if verbose && err == nil {
		fmt.Printf("   Git version: %s\n", version)
	}
	return err == nil
}

func checkJava(verbose bool) bool {
	version, err := utils.RunCapture("java", "-version")
	if verbose && err == nil {
		lines := strings.Split(version, "\n")
		if len(lines) > 0 {
			fmt.Printf("   Java version: %s\n", lines[0])
		}
	}
	return err == nil
}

func checkMavenDaemon(verbose bool) bool {
	// Check mvnd first (preferred)
	version, err := utils.RunCapture("mvnd", "--version")
	if err == nil {
		if verbose {
			fmt.Printf("   Maven Daemon (mvnd) version: %s\n", strings.Split(version, "\n")[0])
		}
		return true
	}

	// mvnd not found
	if verbose {
		fmt.Printf("   Maven Daemon (mvnd) not found\n")
	}
	return false
}

func checkSDKMAN(verbose bool) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	sdkmanDir := filepath.Join(home, ".sdkman")
	if _, err := os.Stat(sdkmanDir); os.IsNotExist(err) {
		return false
	}

	if verbose {
		fmt.Printf("   SDKMAN directory found: %s\n", sdkmanDir)
	}
	return true
}

func checkPortAvailability(port int, service string, verbose bool) {
	if utils.IsPortAvailable(port) {
		fmt.Printf("\x1b[32m✅ Port %d (%s): Available\x1b[0m\n", port, service)
	} else {
		// Check if this is a JUDO project and if our services are using the port
		cfg := config.GetConfig()
		karafUsingPort := false
		postgresUsingPort := false
		keycloakUsingPort := false

		if config.IsProjectInitialized() {
			if cfg.Runtime == "karaf" {
				karafDir := filepath.Join(cfg.ModelDir, "application", ".karaf")
				karafUsingPort = utils.IsPortUsedByKaraf(port, karafDir)
			}

			// Check if PostgreSQL is using the port (for port 5432)
			if port == 5432 && cfg.DBType == "postgresql" {
				postgresUsingPort = docker.IsPortUsedByPostgres(port)
			}

			// Check if Keycloak is using the port (for port 8080)
			if port == 8080 {
				keycloakUsingPort = docker.IsPortUsedByKeycloak(port)
			}
		}

		if karafUsingPort {
			fmt.Printf("\x1b[33m⚠️  Port %d (%s): In use by current Karaf instance\x1b[0m\n", port, service)
			if verbose {
				fmt.Printf("   \x1b[33mNote: This port is used by your running JUDO application\x1b[0m\n")
			}
		} else if postgresUsingPort {
			fmt.Printf("\x1b[33m⚠️  Port %d (%s): In use by current PostgreSQL instance\x1b[0m\n", port, service)
			if verbose {
				fmt.Printf("   \x1b[33mNote: This port is used by your running JUDO PostgreSQL database\x1b[0m\n")
			}
		} else if keycloakUsingPort {
			fmt.Printf("\x1b[33m⚠️  Port %d (%s): In use by current Keycloak instance\x1b[0m\n", port, service)
			if verbose {
				fmt.Printf("   \x1b[33mNote: This port is used by your running JUDO Keycloak instance\x1b[0m\n")
			}
		} else {
			fmt.Printf("\x1b[31m❌ Port %d (%s): In use by another process\x1b[0m\n", port, service)
			if verbose {
				fmt.Printf("   \x1b[31mWarning: This port is occupied by another application, which will cause conflicts\x1b[0m\n")
			}
		}
	}
}

// requireJudoProject checks if a JUDO project is initialized and returns an error if not
func requireJudoProject() error {
	if !config.IsProjectInitialized() {
		return fmt.Errorf("no JUDO project initialized in this directory\nRun 'judo init' to initialize a new JUDO project")
	}
	return nil
}

func CreateLogCommand() *cobra.Command {
	var tail bool
	var follow bool
	var lines int

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Display or tail Karaf console log",
		Long:  "Display the contents of the Karaf console.out log file with optional tailing and following",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if JUDO project is initialized
			if err := requireJudoProject(); err != nil {
				return err
			}

			cfg := config.GetConfig()

			if cfg.Runtime != "karaf" {
				return fmt.Errorf("log command is only supported for karaf runtime")
			}

			logFile := filepath.Join(cfg.KarafDir, "console.out")

			if _, err := os.Stat(logFile); os.IsNotExist(err) {
				return fmt.Errorf("log file not found: %s", logFile)
			}

			if tail || follow {
				return tailLogFile(logFile, lines, follow)
			}

			return displayLogFile(logFile, lines)
		},
	}

	cmd.Flags().BoolVarP(&tail, "tail", "t", false, "Show the end of the log file")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output (like tail -f)")
	cmd.Flags().IntVarP(&lines, "lines", "n", 50, "Number of lines to display")

	return cmd
}

func CreateSessionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "session",
		Short: "Start interactive JUDO CLI session",
		Long:  "Start an interactive session with command history, auto-completion, and persistent state",
		Run: func(cmd *cobra.Command, args []string) {
			// Import session package here to avoid circular imports
			// We'll use a direct function call instead
			fmt.Println("Starting interactive session...")
			// For now, just print a message since we can't import session here due to circular imports
			fmt.Println("Session command would start interactive mode here")
		},
	}
}

func CreateInitCommand() *cobra.Command {
	var projectGroupId string
	var modelName string
	var projectType string
	var pluginVersion string = "LATEST"

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new JUDO project.",
		Long:  "This command initializes a new JUDO project by checking for existing configuration files and, if necessary, generating a new project structure using Maven.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current working directory: %w", err)
			}

			judoVersionPropsPath := filepath.Join(cwd, "judo-version.properties")
			judoPropsPath := filepath.Join(cwd, "judo.properties")

			if utils.FileExists(judoVersionPropsPath) && utils.FileExists(judoPropsPath) {
				fmt.Println("Project already initialized. Both judo-version.properties and judo.properties exist.")
				return nil
			}

			fmt.Println("Project not initialized. Proceeding with initialization...")

			// Prompt for projectGroupId if not provided via flag
			if projectGroupId == "" {
				projectGroupId = utils.PromptForInput("Enter project GroupId (e.g., com.example): ")
				if projectGroupId == "" {
					projectGroupId = "com.example"
				}
			}

			// Prompt for modelName if not provided via flag
			if modelName == "" {
				modelName = utils.PromptForInput("Enter model Name (e.g., MyProject): ")
				if modelName == "" {
					modelName = "MyProject"
				}
			}

			// Determine projectType (default to ESM if not provided)
			if projectType == "" {
				projectType = utils.PromptForSelection("Select project type", []string{"ESM", "JSL"}, "ESM")
			} else {
				// Validate provided projectType
				if !(strings.EqualFold(projectType, "ESM") || strings.EqualFold(projectType, "JSL")) {
					return fmt.Errorf("unsupported project type: %s. Supported types are ESM and JSL.", projectType)
				}
			}

			// Find latest plugin version
			// The user explicitly asked to "search in maven the latest version".
			// I will use `mvn help:evaluate` to get the latest version.
			// This command will output the version to stdout.
			//fmt.Println("Searching for the latest version of judo-version-updater-maven-plugin...")
			//pluginVersion, err := utils.RunCapture("mvn",
			//	"org.apache.maven.plugins:maven-help-plugin:3.2.0:evaluate",
			//	"-Dexpression=hu.blackbelt.judo:judo-version-updater-maven-plugin:LATEST:version",
			//	"-q",
			//	"-DforceStdout",
			//)
			//if err != nil {
			//	fmt.Printf("Warning: Could not determine latest plugin version, using 'LATEST'. Error: %v\n", err)
			//	pluginVersion = "LATEST"
			//} else {
			//	pluginVersion = strings.TrimSpace(pluginVersion)
			//	fmt.Printf("Found latest plugin version: %s\n", pluginVersion)
			//}

			mavenCommand := "mvn"
			var mavenArgs []string

			switch strings.ToUpper(projectType) {
			case "ESM":
				mavenArgs = []string{
					fmt.Sprintf("hu.blackbelt.judo:judo-version-updater-maven-plugin:%s:create-judo-project", pluginVersion),
					fmt.Sprintf("-DmodelName=%s", modelName),
					fmt.Sprintf("-DprojectGroupId=%s", projectGroupId),
					"-U",
				}
			case "JSL":
				mavenArgs = []string{
					fmt.Sprintf("hu.blackbelt.judo:judo-version-updater-maven-plugin:%s:create-jsl-judo-project", pluginVersion),
					fmt.Sprintf("-DmodelName=%s", modelName),
					fmt.Sprintf("-DprojectGroupId=%s", projectGroupId),
					"-U",
				}
			default:
				// This case should ideally not be reached due to the validation above, but as a safeguard.
				return fmt.Errorf("unsupported project type: %s. Supported types are ESM and JSL.", projectType)
			}

			fmt.Printf("Executing Maven command: %s %s\n", mavenCommand, strings.Join(mavenArgs, " "))
			return utils.Run(mavenCommand, mavenArgs...)
		},
	}

	cmd.Flags().StringVarP(&projectGroupId, "group-id", "g", "", "Project GroupId (e.g., com.example)")
	cmd.Flags().StringVarP(&modelName, "model-name", "n", "", "Target model name (e.g., MyProject)")
	cmd.Flags().StringVarP(&projectType, "type", "t", "", "Type of project to create (ESM or JSL, default: ESM)")

	return cmd
}

// displayLogFile displays the contents of a log file with optional line limit
func displayLogFile(logFile string, lines int) error {
	content, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	logContent := string(content)
	logLines := strings.Split(logContent, "\n")

	// If lines is 0 or negative, show all lines
	if lines <= 0 || lines >= len(logLines) {
		fmt.Print(logContent)
		return nil
	}

	// Show only the last 'lines' number of lines
	start := len(logLines) - lines
	if start < 0 {
		start = 0
	}

	for i := start; i < len(logLines); i++ {
		fmt.Println(logLines[i])
	}

	return nil
}

// tailLogFile tails a log file with optional following
func tailLogFile(logFile string, lines int, follow bool) error {
	content, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	logContent := string(content)
	logLines := strings.Split(logContent, "\n")

	// Show the last 'lines' number of lines
	start := len(logLines) - lines
	if start < 0 {
		start = 0
	}

	for i := start; i < len(logLines); i++ {
		fmt.Println(logLines[i])
	}

	if !follow {
		return nil
	}

	// Follow the log file (like tail -f)
	fmt.Printf("\n\x1b[33mFollowing log file (Ctrl+C to stop)...\x1b[0m\n\n")

	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("failed to open log file for following: %w", err)
	}
	defer file.Close()

	// Get current file size and seek to the end
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}

	offset := stat.Size()
	reader := bufio.NewReader(file)

	for {
		// Check for new content
		newStat, err := file.Stat()
		if err != nil {
			return fmt.Errorf("failed to get updated file stats: %w", err)
		}

		if newStat.Size() > offset {
			// Read new content
			file.Seek(offset, 0)
			newContent, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				return fmt.Errorf("failed to read new log content: %w", err)
			}

			if len(newContent) > 0 {
				fmt.Print(string(newContent))
			}

			offset = newStat.Size()
		}

		time.Sleep(1 * time.Second)
	}
}

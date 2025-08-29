package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	cfg := config.GetConfig()
	// Pre-flight checks
	if !docker.IsDockerRunning() {
		log.Fatal("Docker daemon is not running. Please start Docker and try again.")
	}

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

	// Port checks
	if config.Options.StartKeycloak {
		if !utils.IsPortAvailable(cfg.KeycloakPort) {
			log.Fatalf("Keycloak port %d is already in use.", cfg.KeycloakPort)
		}
	}
	if cfg.DBType == "postgresql" {
		if !utils.IsPortAvailable(cfg.PostgresPort) {
			log.Fatalf("PostgreSQL port %d is already in use.", cfg.PostgresPort)
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
			log.Fatalf("Karaf port %d is already in use.", cfg.KarafPort)
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

	karaf.StartKaraf()
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

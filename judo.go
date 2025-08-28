package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io/fs"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	profile              string
	appName              string
	modelDir             string
	schemaName           string
	keycloakName         string
	karafPort            int
	postgresPort         int
	keycloakPort         int
	runtimeEnv           string
	dbType               string
	composeEnv           string
	composeAccessIP      string
	karafEnableAdminUser bool
	javaCompiler         string
)

type JudoOptions struct {
	Clean             bool
	Prune             bool
	Update            bool
	Generate          bool
	GenerateRoot      bool
	Dump              bool
	Import            bool
	Build             bool
	Reckless          bool
	Start             bool
	Stop              bool
	Status            bool
	SchemaUpgrade     bool
	SchemaBuilding    bool
	SchemaCliBuilding bool
	PruneFrontend     bool
	PruneConfirmation bool
	IgnoreChecksum    bool
	QuickMode         bool
	DockerBuilding    bool
	BuildParallel     bool
	BuildAppModule    bool
	BuildFrontend     bool
	BuildKaraf        bool
	BuildModel        bool
	BuildBackend      bool
	StartKeycloak     bool
	WatchBundles      bool
	VersionNumber     string
	ExtraMavenArgs    string
	DumpName          string
}

// State used by prune command flags
type State struct {
	pruneFrontend bool
	pruneConfirm  bool
}
type Config struct {
	AppName      string
	SchemaName   string
	KeycloakName string
	ModelDir     string
	AppDir       string
	KarafDir     string
	Runtime      string // "karaf" | "compose"
	DBType       string // "hsqldb" | "postgresql"
}

var options JudoOptions

func rootHelp() string {
	return `JUDO runner.

USAGE: judo COMMANDS... [OPTIONS...]
    env <env>                               Use alternate env (custom properties file). Default judo is used.
    clean                                   Stop postgresql docker container and clear data.
    prune                                   Stop postgresql docker container and delete untracked files in this repository.
        -f                                  Clear only frontend data.
        -y                                  Skip confirmation.

    update                                  Update dependency versions in JUDO project.
        -i --ignore-checksum                Ignores checksum errors and updates checksums according to new sources.

    generate                                Generate application based on model in JUDO project.
        -i --ignore-checksum                Ignores checksum errors and updates checksums according to new sources.

    generate-root                           Generate application root structure on model in JUDO project.
        -i --ignore-checksum                Ignores checksum errors and updates checksums according to new sources.

    dump                                    Dump postgresql db data before clearing and starting application.
    import                                  Import postgresql db data
        -dn --dump-name                     Import dump name when it's not defined loaded the last one
    schema-upgrade                          It can be used with persistent db (postgresql) only. It uses the current running database to
                                            generate the difference and after it applied.
    build                                   Build project.
        -v <VERSION> --version <VERSION>    Use given version as model and application version
        -p --build-parallel                 Parallel maven build. The log can be chaotic.
        -a --build-app-module               Build app module only.
        -f --build-fronted-module           Build fronted module only.
        -sc --build-schema-cli              Build schema CLI standalon JAR file.
        -d --docker                         Build docker images.
        -M --skip-model                     Skip model building.
        -B --skip-backend                   Skip backend building.
        -F --skip-frontend                  Skip frontend building.
        -KA --skip-karaf                    Skip Backend Karaf building.
        -S --skip-schema                    Skip building schema migration image.
        -ma --maven-argument                Add extra maven argument.
        -q --quick                          Quick mode which uses cache and ignores validations.
        -i --ignore-checksum                Ignores checksum errors and updates checksums according to new sources.

    reckless                                Build and run project in reckless mode. It is skipping validations, docker builds and run as fast as possible.
    start                                   Run application with postgresql and keycloak.
        -W --skip-watch-bundles             Disable watching of bundle changes
        -K --skip-keycloak                  Skip starting keycloak.
        -o "<name>=<value>,<name2>=<value2>, ... " --options "<name>=<value>,<name2>=<value2>, ..."
                                            Add options (defaults can be defined in judo.properties)
                                            Available options:
                                               runtime = karaf | compose
                                               dbtype = hsqldb | postgresql
                                               compose_env = compose-develop | compose-postgresql-https | or any directory defined in ${MODEL_DIR}/docker
                                               model_dir = model project directory. Default is the application's parent.
                                               karaf_port = <port>
                                               postgres_port = <port>
                                               keycloak_port = <port>
                                               compose_access_ip = <alternate ip address to access app>
                                               karaf_enable_admin_user = 1
                                               java_compiler = ejc | javac. Which compuler can be used, default is ejc
    stop                                    Stop application, postgresql and keycloak. (if running)
    status                                  Print status of containers


EXAMPLES:
    judo prune -f                           Clear untracked data in application/frontend if opening modeling project freezes in designer.

    judo build -a                           Build app module only for backend. Useful for updating custom operations for running backend.
    judo build -f -q                        Build model and fronted only in quick mode. Useful when frontend changes needs to be checked.
    judo build -F -KA                       Build model and backend without frontend. Useful when custom operations need to be implemented.

    judo prune build start                  Super fresh application build and start.
    judo build clean start                  Stop postgresql docker container then build and run application (including keycloak) with clean db.
    judo build start -K                     Stop postgresql docker container then build and restart application.
    judo build -M -F clean start            Stop postgresql docker container then rebuild app and start application with clean db.
    judo build -M -F start -K               Rebuild app and restart application.
    judo build -ma "-rf :${app_name}-application-karaf-offline"
                                            Continue build from module.
    judo start -o "runtime=compose,compose_env=compose-postgresql-https'"
                                            Run application in docker compose using the compose-postgresql-https's docker-compose.yaml

    judo env compose-dev build start        Build and run application with compose-dev env. (have to be described in compose-dev.properties)
`
}

func startLongHelp() string {
	return `Run application with postgresql and keycloak.

  -W --skip-watch-bundles   Disable watching of bundle changes
  -K --skip-keycloak        Skip starting keycloak.
  -o, --options "<k=v,k2=v2,...>"
                            Add options (defaults can be defined in judo.properties)

Available options:
  runtime = karaf | compose
  dbtype = hsqldb | postgresql
  compose_env = compose-develop | compose-postgresql-https | or any directory defined in ${MODEL_DIR}/docker
  model_dir = model project directory. Default is the application's parent.
  karaf_port = <port>
  postgres_port = <port>
  keycloak_port = <port>
  compose_access_ip = <alternate ip address to access app>
  karaf_enable_admin_user = 1
  java_compiler = ejc | javac (default ejc)
`
}

func buildLongHelp() string {
	return `Build project.

  -v <VER> --version <VER>  Use given version as model and application version
  -p --build-parallel       Parallel maven build (log can be chaotic)
  -a --build-app-module     Build app module only
  -f --build-frontend-module
                            Build frontend module only
  -sc --build-schema-cli    Build schema CLI standalone JAR file
  -d --docker               Build docker images
  -M --skip-model           Skip model building
  -B --skip-backend         Skip backend building
  -F --skip-frontend        Skip frontend building
  -KA --skip-karaf          Skip Backend Karaf building
  -S --skip-schema          Skip building schema migration image
  -ma --maven-argument ARG  Add extra maven argument
  -q --quick                Quick mode (cache + skip validations)
  -i --ignore-checksum      Ignore checksum errors
`
}

func cleanLongHelp() string {
	return `Stop postgresql docker container and clear data.

This removes:
  • All Docker containers for postgres-<schema>, keycloak-<keycloak>
  • The Docker network <app_name>
  • Volumes: <app>_certs, <schema>_postgresql_db, <schema>_postgresql_data, <app>_filestore
  • Karaf dir (application/.karaf) if running in local 'karaf' runtime.
`
}

func pruneLongHelp() string {
	return `Stop postgresql docker container and delete untracked files in this repository.

Options:
  -f, --frontend    Clear only frontend data (application/frontend-react)
  -y, --yes         Skip confirmation prompt

Notes:
  • Only supported inside a Git repository (uses 'git clean -dffx').
  • When not using --frontend, will also stop Karaf/Keycloak/PostgreSQL (if applicable) before cleaning.
`
}

func updateLongHelp() string {
	return `Update dependency versions in JUDO project.

This runs mvnd clean compile with:
  -DgenerateRoot -DskipApplicationBuild -DupdateJudoVersions=true -U

Options:
  -i, --ignore-checksum   Ignore checksum errors and update checksums
`
}

func generateLongHelp() string {
	return `Generate application based on model in JUDO project.

Runs:
  mvnd clean compile -DgenerateApplication -DskipApplicationBuild -f <MODEL_DIR>

Options:
  -i, --ignore-checksum   Ignore checksum errors and update checksums
`
}

func generateRootLongHelp() string {
	return `Generate application root structure based on model in JUDO project.

Runs:
  mvnd clean compile -DgenerateRoot -DskipApplicationBuild -U -f <MODEL_DIR>

Options:
  -i, --ignore-checksum   Ignore checksum errors and update checksums
`
}

func dumpLongHelp() string {
	return `Dump postgresql DB data before clearing/starting application.

Behavior:
  • Ensures PostgreSQL is running locally (docker) for <schema>.
  • Creates dump file: <schema>_dump_YYYYMMDD_HHMMSS.tar.gz (pg_dump -F c).
  • Stops the container afterward.

Notes:
  • Works only when dbtype=postgresql.
`
}

func importLongHelp() string {
	return `Import postgresql DB data.

Behavior:
  • Recreates the postgres container volumes for a fresh state.
  • Starts postgres and waits for readiness.
  • Restores a dump with pg_restore -Fc --clean (from given file or latest matching <schema>_dump_*.tar.gz).
  • Restarts the container at the end.

Options:
  -n, --dump-name <FILE>  Specific dump filename to import. If omitted, the latest <schema>_dump_*.tar.gz is used.

Notes:
  • Works only when dbtype=postgresql.
`
}

func schemaUpgradeLongHelp() string {
	return `Apply RDBMS schema upgrade using the current running database (PostgreSQL only).

Behavior:
  • Ensures local PostgreSQL is up.
  • Executes 'judo-rdbms-schema:apply' against jdbc:postgresql://127.0.0.1:<port>/<schema>
    with -DschemaIgnoreModelDependency=true and -DupdateModel pointing to the generated model.

Notes:
  • Works only when dbtype=postgresql.
  • Expects generated model at: model/target/generated-resources/model/<schema>-rdbms_postgresql.model
`
}

func stopLongHelp() string {
	return `Stop application, postgresql and keycloak (if running).

Behavior (karaf runtime):
  • Stops Karaf if running.
  • Stops postgres-<schema> (when dbtype=postgresql).
  • Stops keycloak-<keycloak>.
`
}

func statusLongHelp() string {
	return `Print status of containers and local Karaf.

Reports:
  • Karaf running/not running (based on application/.karaf/bin/status).
  • PostgreSQL running/not running + container/volume existence (if dbtype=postgresql).
  • Keycloak running/not running + container existence.
`
}

func recklessLongHelp() string {
	return `Build and run project in reckless mode.

Behavior:
  • Optimizes for speed: skips validations, schema/docker builds, favors 'package'.
  • Starts local environment first (Karaf runtime).
  • Useful for quick iteration; not for reproducible CI builds.
`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "judo",
		Short: "JUDO CLI",
		Long:  rootHelp(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			loadProperties()
			setupEnvironment()
		},
	}

	// Common flags
	rootCmd.PersistentFlags().StringVarP(&profile, "env", "e", "judo", "Use alternate environment")

	// Add commands
	rootCmd.AddCommand(
		createCleanCommand(),
		createPruneCommand(),
		createUpdateCommand(),
		createGenerateCommand(),
		createGenerateRootCommand(),
		createDumpCommand(),
		createImportCommand(),
		createSchemaUpgradeCommand(),
		createBuildCommand(),
		createRecklessCommand(),
		createStartCommand(),
		createStopCommand(),
		createStatusCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createGenerateCommand() *cobra.Command {
	var ignore bool
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate application based on model in JUDO project.",
		Long:  generateLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			cwd, _ := os.Getwd()
			cfg := defaultConfig(cwd)

			// Stop mvnd daemon first (same as bash script)
			_ = run("mvnd", "--purge", "--stop")

			args := []string{
				"clean", "compile",
				"-DgenerateApplication",
				"-DskipApplicationBuild",
				"-f", cfg.ModelDir,
			}
			if ignore {
				args = append(args, "-DvalidateChecksum=false")
			}
			return run("mvnd", args...)
		},
	}
	cmd.Flags().BoolVarP(&ignore, "ignore-checksum", "i", false, "Ignore checksum errors and update checksums")
	return cmd
}
func createBuildCommand() *cobra.Command {
	// sensible defaults (match bash)
	options.BuildModel = true
	options.BuildBackend = true
	options.BuildFrontend = true
	options.BuildKaraf = true
	options.SchemaBuilding = true
	options.SchemaCliBuilding = false
	options.DockerBuilding = false

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build project",
		Long:  buildLongHelp(),
		Run:   runBuild,
	}

	cmd.Flags().BoolVarP(&options.BuildParallel, "build-parallel", "p", false, "Parallel maven build")
	cmd.Flags().BoolVarP(&options.BuildAppModule, "build-app-module", "a", false, "Build app module only")
	cmd.Flags().BoolVarP(&options.BuildFrontend, "build-frontend-module", "f", false, "Build frontend module only")
	cmd.Flags().BoolVar(&options.DockerBuilding, "docker", false, "Build Docker images")
	// skip-* flags flip the respective build toggles
	cmd.Flags().Bool("skip-model", false, "Skip model building")
	cmd.Flags().Bool("skip-backend", false, "Skip backend building")
	cmd.Flags().Bool("skip-frontend", false, "Skip frontend building")
	cmd.Flags().Bool("skip-karaf", false, "Skip Backend Karaf building")
	cmd.Flags().Bool("skip-schema", false, "Skip building schema migration image")
	cmd.Flags().Bool("build-schema-cli", false, "Build schema CLI standalone JAR file")
	cmd.Flags().StringVarP(&options.VersionNumber, "version", "v", "SNAPSHOT", "Version number")
	cmd.Flags().StringVarP(&options.ExtraMavenArgs, "maven-argument", "m", "", "Extra Maven args (quoted)")
	cmd.Flags().BoolVarP(&options.QuickMode, "quick", "q", false, "Quick mode: cache + skip validations")
	cmd.Flags().BoolVarP(&options.IgnoreChecksum, "ignore-checksum", "i", false, "Ignore checksum errors")

	return cmd
}

func createRecklessCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reckless",
		Short: "Build & run fast (skips validations, favors speed)",
		Long:  recklessLongHelp(),
		Run: func(cmd *cobra.Command, _ []string) {
			options.Reckless = true
			options.QuickMode = true
			options.BuildKaraf = false // match bash defaults for reckless build matrix
			runBuild(cmd, nil)
		},
	}
}

func createGenerateRootCommand() *cobra.Command {
	var ignore bool
	cmd := &cobra.Command{
		Use:   "generate-root",
		Short: "Generate application root structure based on model in JUDO project.",
		Long:  generateRootLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			cwd, _ := os.Getwd()
			cfg := defaultConfig(cwd)

			_ = run("mvnd", "--purge", "--stop")

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
			return run("mvnd", args...)
		},
	}
	cmd.Flags().BoolVarP(&ignore, "ignore-checksum", "i", false, "Ignore checksum errors and update checksums")
	return cmd
}

func createStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Print status of Karaf/Keycloak/PostgreSQL containers and resources",
		Long:  statusLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			cwd, _ := os.Getwd()
			cfg := defaultConfig(cwd)

			fmt.Println("Runtime:", runtimeEnv, " DB:", cfg.DBType)
			if runtimeEnv == "karaf" {
				// Karaf
				if karafRunning(cfg.KarafDir) {
					fmt.Println("Karaf is running")
				} else {
					fmt.Println("Karaf is not running")
				}

				// Postgres (if applicable)
				if cfg.DBType == "postgresql" {
					pgName := "postgres-" + cfg.SchemaName
					if dockerInstanceRunning(pgName) {
						fmt.Println("PostgreSQL is running")
					} else {
						fmt.Println("PostgreSQL is not running")
						if containerExists(pgName) {
							fmt.Println("PostgreSQL container exists")
						} else {
							fmt.Println("PostgreSQL container does not exist")
						}
						if dockerVolumeExists(cfg.AppName + "_postgresql_db") {
							fmt.Println("PostgreSQL db volume exists")
						} else {
							fmt.Println("PostgreSQL db volume does not exist")
						}
						if dockerVolumeExists(cfg.AppName + "_postgresql_data") {
							fmt.Println("PostgreSQL data volume exists")
						} else {
							fmt.Println("PostgreSQL data volume does not exist")
						}
					}
				}

				// Keycloak
				kcName := "keycloak-" + cfg.KeycloakName
				if dockerInstanceRunning(kcName) {
					fmt.Println("Keycloak is running")
				} else {
					fmt.Println("Keycloak is not running")
					if containerExists(kcName) {
						fmt.Println("Keycloak container exists")
					} else {
						fmt.Println("Keycloak container does not exist")
					}
				}
			}

			return nil
		},
	}
}

func createDumpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Dump PostgreSQL DB data (creates <schema>_dump_YYYYMMDD_HHMMSS.tar.gz).",
		Long:  dumpLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			cwd, _ := os.Getwd()
			cfg := defaultConfig(cwd)

			if cfg.DBType != "postgresql" {
				fmt.Println("Dump is only supported with PostgreSQL.")
				return nil
			}

			// Ensure DB is up, then dump, then stop it (like the bash script)
			startPostgres()
			name := "postgres-" + cfg.SchemaName
			file, err := dumpPostgresql(name, cfg.SchemaName)
			if err != nil {
				return err
			}
			fmt.Println("Database dumped to", file)
			_ = stopDockerInstance(name)
			return nil
		},
	}
	return cmd
}

func createImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import PostgreSQL DB dump (pg_restore).",
		Long:  importLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			cwd, _ := os.Getwd()
			cfg := defaultConfig(cwd)

			if cfg.DBType != "postgresql" {
				fmt.Println("Import is only supported with PostgreSQL.")
				return nil
			}

			instance := "postgres-" + cfg.SchemaName
			// Fresh db state
			_ = removeDockerInstance(instance)
			_ = removeDockerVolume(cfg.SchemaName + "_postgresql_db")
			_ = removeDockerVolume(cfg.SchemaName + "_postgresql_data")

			// Start DB and wait
			startPostgres()

			// Determine dump file
			dumpFile := options.DumpName
			if strings.TrimSpace(dumpFile) == "" {
				var err error
				dumpFile, err = findLatestDump(cfg.SchemaName)
				if err != nil {
					return err
				}
			}
			fmt.Println("Loading dump:", dumpFile)

			// Run pg_restore inside the container
			if err := importPostgresql(instance, cfg.SchemaName, dumpFile); err != nil {
				return err
			}

			// Bounce container (same as bash)
			_ = stopDockerInstance(instance)
			startPostgres()
			return nil
		},
	}
	// Bash used -dn / --dump-name; we expose -n/--dump-name here.
	cmd.Flags().StringVarP(&options.DumpName, "dump-name", "n", "", "Dump filename to import (defaults to latest <schema>_dump_*.tar.gz)")
	return cmd
}

func createSchemaUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema-upgrade",
		Short: "Apply RDBMS schema upgrade using current running database (PostgreSQL only).",
		Long:  schemaUpgradeLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			cwd, _ := os.Getwd()
			cfg := defaultConfig(cwd)

			if cfg.DBType != "postgresql" {
				fmt.Println("Schema upgrade requires PostgreSQL.")
				return nil
			}

			// Ensure Postgres is started and reachable
			startPostgres()

			updateModel := filepath.Join(modelDir, "model", "target", "generated-resources", "model",
				fmt.Sprintf("%s-rdbms_postgresql.model", cfg.SchemaName))
			schemaDir := filepath.Join(modelDir, "schema")

			args := []string{
				"judo-rdbms-schema:apply",
				fmt.Sprintf("-DjdbcUrl=jdbc:postgresql://127.0.0.1:%d/%s", postgresPort, cfg.SchemaName),
				"-DdbType=postgresql",
				"-DdbUser=" + cfg.SchemaName,
				"-DdbPassword=" + cfg.SchemaName,
				"-DschemaIgnoreModelDependency=true",
				"-DupdateModel=" + updateModel,
				"-f", schemaDir,
			}
			return run("mvnd", args...)
		},
	}
	return cmd
}

// --- Helpers for dump/import ---

func dumpPostgresql(containerName, schema string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	file := fmt.Sprintf("%s_dump_%s.tar.gz", schema, timestamp)

	out, err := os.Create(file)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Use pg_dump custom format (-F c), same as bash
	cmd := exec.Command("docker", "exec", "-i", containerName,
		"/bin/bash", "-c",
		fmt.Sprintf("PGPASSWORD=%s pg_dump --username=%s -F c %s", schema, schema, schema))
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return file, cmd.Run()
}

func importPostgresql(containerName, schema, dumpFile string) error {
	in, err := os.Open(dumpFile)
	if err != nil {
		return err
	}
	defer in.Close()

	cmd := exec.Command("docker", "exec", "-i", containerName,
		"/bin/bash", "-c",
		fmt.Sprintf("PGPASSWORD=%s pg_restore -Fc --clean -U %s -d %s", schema, schema, schema))
	cmd.Stdin = in
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findLatestDump(schema string) (string, error) {
	pattern := fmt.Sprintf("%s_dump_*.tar.gz", schema)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no dump files found matching %q", pattern)
	}

	// Filenames include timestamp; lexicographic max is the latest
	latest := matches[0]
	for _, m := range matches[1:] {
		if m > latest {
			latest = m
		}
	}
	return latest, nil
}

func defaultConfig(cwd string) *Config {
	cfg := &Config{
		AppName:  filepath.Base(cwd),
		ModelDir: cwd,
		AppDir:   filepath.Join(cwd, "application"),
		KarafDir: filepath.Join(cwd, "application", ".karaf"),
		Runtime:  "karaf",
		DBType:   "hsqldb",
	}
	if cfg.SchemaName == "" {
		cfg.SchemaName = cfg.AppName
	}
	if cfg.KeycloakName == "" {
		cfg.KeycloakName = cfg.AppName
	}
	return cfg
}

func getComposeEnvs(cfg *Config) []string {
	root := filepath.Join(cfg.AppDir, "docker")
	envs := []string{}
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Base(path) == "docker-compose.yml" {
			envs = append(envs, filepath.Base(filepath.Dir(path)))
		}
		return nil
	})
	return envs
}

func stopCompose(cfg *Config, env string) error {
	composeFile := filepath.Join(cfg.AppDir, "docker", env, "docker-compose.yml")
	cmd := exec.Command("docker", "compose", "-f", composeFile, "down", "--volumes")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func removeDockerInstance(name string) error {
	if name == "" {
		return nil
	}
	cmd := exec.Command("docker", "rm", "-f", name)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	_ = cmd.Run() // ignore if it doesn't exist
	return nil
}
func createDockerNetwork(name string) {
	out, _ := runCapture("docker", "network", "ls", "--format", "{{.Name}}")
	for _, n := range strings.Split(out, "\n") {
		if n == name {
			return
		}
	}
	_ = executeCommand("docker", "network", "create", name)
}

func removeDockerNetwork(name string) error {
	if name == "" {
		return nil
	}
	cmd := exec.Command("docker", "network", "rm", name)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	_ = cmd.Run() // ignore if it doesn't exist
	return nil
}

func removeDockerVolume(name string) error {
	if name == "" {
		return nil
	}
	cmd := exec.Command("docker", "volume", "rm", name)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	_ = cmd.Run() // ignore if it doesn't exist
	return nil
}

func dockerVolumeExists(name string) bool {
	if name == "" {
		return false
	}
	out, _ := runCapture("docker", "volume", "ls", "--format", "{{.Name}}")
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == name {
			return true
		}
	}
	return false
}

func stopKaraf(karafDir string) {
	if karafDir == "" {
		return
	}
	_ = exec.Command(filepath.Join(karafDir, "bin", "stop")).Run()
}

// Docker stop helper (no-op if not running)
func dockerInstanceRunning(name string) bool {
	out, _ := runCapture("docker", "ps", "--format", "{{.Names}}")
	for _, n := range strings.Split(out, "\n") {
		if n == name {
			return true
		}
	}
	return false
}

func stopDockerInstance(name string) error {
	if dockerInstanceRunning(name) {
		return run("docker", "stop", name)
	}
	return nil
}

// Clean
func createCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Stop postgresql docker container and clear data.",
		Long:  cleanLongHelp(),
		RunE: func(_ *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			cfg := defaultConfig(cwd)
			for _, env := range getComposeEnvs(cfg) {
				_ = stopCompose(cfg, env)
			}
			_ = removeDockerInstance("postgres-" + cfg.SchemaName)
			_ = removeDockerInstance("keycloak-" + cfg.KeycloakName)
			_ = removeDockerNetwork(cfg.AppName)
			_ = removeDockerVolume(cfg.AppName + "_certs")
			_ = removeDockerVolume(cfg.SchemaName + "_postgresql_db")
			_ = removeDockerVolume(cfg.SchemaName + "_postgresql_data")
			_ = removeDockerVolume(cfg.AppName + "_filestore")
			if cfg.Runtime == "karaf" {
				stopKaraf(cfg.KarafDir)
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

// Prune
func createPruneCommand() *cobra.Command {
	var frontend bool
	var yes bool
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Stop postgresql docker container and delete untracked files in this repository.",
		Long:  pruneLongHelp(),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			cfg := defaultConfig(cwd)
			st := &State{pruneFrontend: frontend, pruneConfirm: !yes}
			pruneApplication(cfg, st)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&frontend, "frontend", "f", false, "Clear only frontend data")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation")
	return cmd
}

// Update
func createUpdateCommand() *cobra.Command {
	var ignore bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update dependency versions in JUDO project.",
		Long:  updateLongHelp(),
		RunE: func(_ *cobra.Command, _ []string) error {
			cwd, _ := os.Getwd()
			cfg := defaultConfig(cwd)

			// Run SDKMAN steps (Unix or via WSL on Windows). Safe to skip if unavailable.
			_ = sdkmanRun(
				"sdk selfupdate || true",
				"sdk env install || true",
				"sdk env || true",
			)

			// Stop mvnd daemon like the bash script
			_ = run("mvnd", "--purge", "--stop")

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
			return run("mvnd", mvnargs...)
		},
	}
	cmd.Flags().BoolVarP(&ignore, "ignore-checksum", "i", false, "Ignore checksum errors and update checksums")
	return cmd
}

// Stop
func createStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop application, postgresql and keycloak (if running)",
		Long:  stopLongHelp(),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			cfg := defaultConfig(cwd)
			if cfg.Runtime == "karaf" {
				stopKaraf(cfg.KarafDir)
				if cfg.DBType == "postgresql" {
					_ = stopDockerInstance("postgres-" + cfg.SchemaName)
				}
				_ = stopDockerInstance("keycloak-" + cfg.KeycloakName)
			}
			return nil
		},
	}
	return cmd
}

func runBuild(cmd *cobra.Command, args []string) {
	// reflect skip-* flags into options
	if v, _ := cmd.Flags().GetBool("skip-model"); v {
		options.BuildModel = false
	}
	if v, _ := cmd.Flags().GetBool("skip-backend"); v {
		options.BuildBackend = false
	}
	if v, _ := cmd.Flags().GetBool("skip-frontend"); v {
		options.BuildFrontend = false
	}
	if v, _ := cmd.Flags().GetBool("skip-karaf"); v {
		options.BuildKaraf = false
	}
	if v, _ := cmd.Flags().GetBool("skip-schema"); v {
		options.SchemaBuilding = false
	}
	if v, _ := cmd.Flags().GetBool("build-schema-cli"); v {
		options.SchemaCliBuilding = true
	}

	if options.Reckless {
		// mirror bash: start local env first
		startLocalEnvironment()
	}

	// stop mvnd daemon as in bash (except reckless path which may run fast)
	if !options.Reckless {
		_ = run("mvnd", "--purge", "--stop")
	}

	cwd, _ := os.Getwd()
	cfg := defaultConfig(cwd)

	goal := "install"
	if options.Reckless {
		goal = "package"
	}

	// base args
	buildArgs := []string{goal}
	if !options.Reckless {
		buildArgs = append([]string{"clean"}, buildArgs...)
	}
	buildArgs = append(buildArgs, "-Dsmartbuilder.profiling=true")

	// version handling (-Drevision) when not SNAPSHOT
	if strings.TrimSpace(options.VersionNumber) != "" && strings.ToUpper(strings.TrimSpace(options.VersionNumber)) != "SNAPSHOT" {
		buildArgs = append(buildArgs, "-Drevision="+options.VersionNumber)
	}

	if options.IgnoreChecksum {
		buildArgs = append(buildArgs, "-DvalidateChecksum=false")
	}
	if options.QuickMode {
		buildArgs = append(buildArgs,
			"-Dfrontend-build-type=quick",
			"-DvalidateModels=false",
			"-DuseCache=true",
			"-DskipPrepareNodeJS",
		)
	}
	// parallel? keep conservative default (one core per thread)
	if options.BuildParallel {
		buildArgs = append(buildArgs, "-T", "1C")
	}

	// Apply component toggles
	if !options.BuildFrontend {
		buildArgs = append(buildArgs, "-DskipReact", "-DskipFrontendModel", "-DskipPrepareNodeJS")
	}
	if !options.BuildModel {
		buildArgs = append(buildArgs, "-DskipModels")
	}
	if !options.BuildKaraf {
		buildArgs = append(buildArgs, "-DskipKaraf")
	}
	if !options.DockerBuilding {
		buildArgs = append(buildArgs, "-DskipDocker", "-DskipSchemaDocker", "-DkarafOfflineZip=false")
	}
	if !options.SchemaBuilding {
		buildArgs = append(buildArgs, "-DskipSchema")
	}
	if !options.SchemaCliBuilding {
		buildArgs = append(buildArgs, "-DskipSchemaCli")
	}

	// extra user maven args (best-effort split)
	if s := strings.TrimSpace(options.ExtraMavenArgs); s != "" {
		buildArgs = append(buildArgs, strings.Fields(s)...)
	}

	// Special target layouts (subset builds)
	switch {
	case options.BuildBackend && options.BuildAppModule:
		// build backend app module (+interceptors) only
		fmt.Println("Building backend app module only...")
		args := append([]string{}, buildArgs...)
		args = append(args, "-f", cfg.AppDir, "-pl", "app,interceptors", "-DskipModels=true")
		checkError(run("mvnd", args...))
		return

	case !options.BuildBackend && options.BuildFrontend:
		// frontend only
		fmt.Println("Building frontend only...")
		args := append([]string{}, buildArgs...)
		args = append(args, "-f", filepath.Join(cfg.AppDir, "frontend-react"))
		checkError(run("mvnd", args...))
		return

	default:
		// full (or mostly-full) build, starting at MODEL_DIR
		args := append([]string{}, buildArgs...)
		args = append(args, "-f", cfg.ModelDir)
		checkError(run("mvnd", args...))
	}

	// Reckless extras: optionally (light) post-steps
	if options.Reckless {
		// Skipping schema upgrade + bundle hot-install here to keep it simple and stable.
		fmt.Println("Reckless build completed.")
	}
}

func createStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start application",
		Long:  startLongHelp(),
		Run:   runStart,
	}
	cmd.Flags().Bool("skip-keycloak", false, "Skip starting Keycloak")
	cmd.Flags().Bool("skip-watch-bundles", false, "Disable watching of bundle changes")
	cmd.Flags().String("options", "", "Additional options: key=value,key2=value2 (e.g. runtime=compose,dbtype=postgresql,karaf_port=8181)")
	return cmd
}

func runStart(cmd *cobra.Command, _ []string) {
	// apply flags
	if v, _ := cmd.Flags().GetBool("skip-keycloak"); v {
		options.StartKeycloak = false
	}
	if v, _ := cmd.Flags().GetBool("skip-watch-bundles"); v {
		options.WatchBundles = false
	}

	// parse -o/--options: key=value,key2=value2
	if raw, _ := cmd.Flags().GetString("options"); strings.TrimSpace(raw) != "" {
		applyInlineOptions(raw)
	}

	switch runtimeEnv {
	case "compose":
		startCompose()
	case "karaf":
		startLocalEnvironment()
	default:
		fmt.Println("Unknown runtime:", runtimeEnv, " — defaulting to karaf")
		startLocalEnvironment()
	}
}

func startCompose() {
	fmt.Println("Starting Docker compose environment...")
	ecmd := exec.Command("docker", "compose", "-f", fmt.Sprintf("%s/docker/%s/docker-compose.yml", modelDir, composeEnv), "up")
	ecmd.Stdout = os.Stdout
	ecmd.Stderr = os.Stderr
	checkError(ecmd.Run())
}

func startLocalEnvironment() {
	if dbType == "postgresql" {
		startPostgres()
	}

	if options.StartKeycloak {
		startKeycloak()
	}

	startKaraf()
}

func startPostgres() {
	fmt.Println("Starting PostgreSQL...")
	name := "postgres-" + schemaName

	if !containerExists(name) {
		createDockerNetwork(appName)
		_ = executeCommand(
			"docker", "run", "-d",
			"-v", fmt.Sprintf("%s_postgresql_db:/var/lib/postgresql/pgdata", schemaName),
			"-v", fmt.Sprintf("%s_postgresql_data:/var/lib/postgresql/data", schemaName),
			"--network", appName,
			"--name", name,
			"-e", "PGDATA=/var/lib/postgresql/pgdata",
			"-e", fmt.Sprintf("POSTGRES_USER=%s", schemaName),
			"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", schemaName),
			"-p", fmt.Sprintf("%d:5432", postgresPort),
			"postgres:16.2",
		)
	} else {
		startContainer(name)
	}
	waitForPort("localhost", postgresPort, 30*time.Second)
}

func startKeycloak() {
	fmt.Println("Starting Keycloak...")
	name := "keycloak-" + keycloakName

	if !containerExists(name) {
		if dbType == "postgresql" {
			createDockerNetwork(appName)
		}
		args := []string{
			"run", "-d",
			"--name", name,
			"-e", "KEYCLOAK_ADMIN=admin",
			"-e", "KEYCLOAK_ADMIN_PASSWORD=judo",
			"-p", fmt.Sprintf("%d:%d", keycloakPort, keycloakPort),
		}
		// DB wiring like in the bash script
		if dbType == "postgresql" {
			args = append(args,
				"--network", appName,
				"-e", "KC_DB=postgres",
				"-e", "KC_DB_URL_HOST=postgres-"+schemaName,
				"-e", "KC_DB_URL_DATABASE="+schemaName,
				"-e", "KC_DB_PASSWORD="+schemaName,
				"-e", "KC_DB_USERNAME="+schemaName,
				"-e", "KC_DB_SCHEMA=public",
			)
		}
		args = append(args,
			"-it", "quay.io/keycloak/keycloak:23.0",
			"start-dev",
			fmt.Sprintf("--http-port=%d", keycloakPort),
			"--http-relative-path", "/auth",
		)
		_ = executeCommand("docker", args...)
	} else {
		startContainer(name)
	}
	waitForPort("localhost", keycloakPort, 30*time.Second)
}

func startKaraf() {
	fmt.Println("Starting Karaf...")

	// env like in the bash
	os.Setenv("JUDO_PLATFORM_RDBMS_DIALECT", dbType)
	if dbType == "postgresql" {
		os.Setenv("JUDO_PLATFORM_RDBMS_DB_HOST", "localhost")
		os.Setenv("JUDO_PLATFORM_RDBMS_DB_PORT", fmt.Sprintf("%d", postgresPort))
	}
	os.Setenv("JUDO_PLATFORM_RDBMS_DB_DATABASE", schemaName)
	os.Setenv("JUDO_PLATFORM_RDBMS_DB_USER", schemaName)
	os.Setenv("JUDO_PLATFORM_RDBMS_DB_PASSWORD", schemaName)
	os.Setenv("JUDO_PLATFORM_KEYCLOAK_AUTH_SERVER_URL", fmt.Sprintf("http://localhost:%d/auth", keycloakPort))
	if !options.WatchBundles {
		os.Setenv("JUDO_PLATFORM_BUNDLE_WATCHER", "false")
	}
	os.Setenv("EXTRA_JAVA_OPTS", "-Xms1024m -Xmx1024m -Dfile.encoding=UTF-8 -Dsun.jnu.encoding=UTF-8")

	karafDir := filepath.Join(modelDir, "application", ".karaf")
	_ = os.RemoveAll(karafDir)
	_ = os.MkdirAll(karafDir, 0o755)

	ver := getProjectVersion()
	tarPath := filepath.Join(modelDir, "application", "karaf-offline", "target",
		fmt.Sprintf("%s-application-karaf-offline-%s.tar.gz", appName, ver),
	)
	// extract
	_ = executeCommand("tar", "xzf", tarPath, "-C", karafDir)

	// flatten top-level dir if present
	entries, _ := os.ReadDir(karafDir)
	if len(entries) == 1 && entries[0].IsDir() {
		top := filepath.Join(karafDir, entries[0].Name())
		children, _ := os.ReadDir(top)
		for _, ch := range children {
			_ = os.Rename(filepath.Join(top, ch.Name()), filepath.Join(karafDir, ch.Name()))
		}
		_ = os.RemoveAll(top)
	}

	// tweak http port
	pax := filepath.Join(karafDir, "etc", "org.ops4j.pax.web.cfg")
	_ = replaceInFile(pax, `org\.osgi\.service\.http\.port\s*=\s*\d+`, fmt.Sprintf("org.osgi.service.http.port = %d", karafPort))

	// optionally enable admin user
	if karafEnableAdminUser {
		users := filepath.Join(karafDir, "etc", "users.properties")
		_ = replaceInFile(users, `#karaf\s*=\s*`, "karaf = ")
		_ = replaceInFile(users, `#_g_/`, "_g_/")
	}

	// start in background, write logs to console.out
	consoleOut, _ := os.Create(filepath.Join(karafDir, "console.out"))
	ecmd := exec.Command(filepath.Join(karafDir, "bin", "karaf"), "debug", "run", "clean")
	ecmd.Stdout = consoleOut
	ecmd.Stderr = consoleOut
	_ = ecmd.Start()

	fmt.Printf("Karaf started (pid %d). Logs: %s\n", ecmd.Process.Pid, consoleOut.Name())
}

func karafRunning(karafDir string) bool {
	if karafDir == "" {
		return false
	}
	status := filepath.Join(karafDir, "bin", "status")
	if _, err := os.Stat(status); err != nil {
		return false
	}
	out, _ := runCapture(status)
	return strings.Contains(out, "Running")
}

// Helper functions
func executeMaven(args []string) {
	cmd := exec.Command("mvnd", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	checkError(cmd.Run())
}

func containerExists(name string) bool {
	cmd := exec.Command("docker", "ps", "-a", "-f", fmt.Sprintf("name=%s", name))
	output, _ := cmd.Output()
	return strings.Contains(string(output), name)
}

func startContainer(name string) {
	checkError(exec.Command("docker", "start", name).Run())
}

// Prune logic used by createPruneCommand()
func pruneApplication(cfg *Config, st *State) {
	canContinue := "Y"
	if st.pruneConfirm {
		location := "this repository"
		if st.pruneFrontend {
			location = "application/frontend-react"
		}
		print("Prune command deletes all untracked files in " + location + "!\nAre you sure you want to continue? [Y/n]: ")
		sc := bufio.NewScanner(os.Stdin)
		if sc.Scan() {
			canContinue = strings.TrimSpace(sc.Text())
		}
	}
	if strings.ToUpper(canContinue) != "Y" {
		println("Aborting prune.")
		os.Exit(13)
	}

	if st.pruneFrontend {
		_ = run("git", "clean", "-dffx", filepath.Join(cfg.AppDir, "frontend-react"))
		return
	}

	if cfg.DBType == "postgresql" {
		_ = stopDockerInstance("postgres-" + cfg.SchemaName)
	}
	_ = stopDockerInstance("keycloak-" + cfg.KeycloakName)
	if cfg.Runtime == "karaf" {
		stopKaraf(cfg.KarafDir)
	}
	_ = run("git", "clean", "-dffx", cfg.ModelDir)
}

func waitForPort(host string, port int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)

	fmt.Printf("Wait for port %d on %s.\n", port, host)
	for {
		c, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 2*time.Second)
		if err == nil {
			_ = c.Close()
			fmt.Println()
			return
		}
		if time.Now().After(deadline) {
			fmt.Println("\nWait timed out.")
			log.Fatal("waitForPort timed out")
		}
		fmt.Print(".")
		time.Sleep(1 * time.Second)
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Small exec helpers
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return cmd.Run()
}

func runCapture(name string, args ...string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

// Optional shim if your code calls executeCommand(...)
func executeCommand(name string, args ...string) error { return run(name, args...) }

// tiny .properties reader
func readProperties(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	props := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if i := strings.IndexAny(line, "=:"); i >= 0 {
			k := strings.TrimSpace(line[:i])
			v := strings.TrimSpace(line[i+1:])
			props[k] = v
		}
	}
	return props
}

func loadProperties() {
	// default MODEL_DIR = current working dir
	wd, _ := os.Getwd()
	modelDir = wd

	// prefer <profile>.properties, then judo.properties
	var props map[string]string
	candidates := []string{
		filepath.Join(modelDir, profile+".properties"),
		filepath.Join(modelDir, "judo.properties"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			props = readProperties(p)
			break
		}
	}
	if props == nil {
		return
	}

	if v := props["model_dir"]; v != "" {
		if filepath.IsAbs(v) {
			modelDir = v
		} else {
			modelDir = filepath.Clean(filepath.Join(modelDir, v))
		}
	}
	if v := props["app_name"]; v != "" {
		appName = v
	}
	if v := props["schema_name"]; v != "" {
		schemaName = v
	}
	if v := props["keycloak_name"]; v != "" {
		keycloakName = v
	}
	if v := props["runtime"]; v != "" {
		runtimeEnv = v
	}
	if v := props["dbtype"]; v != "" {
		dbType = v
	}
	if v := props["compose_env"]; v != "" {
		composeEnv = v
	}
	if v := props["compose_access_ip"]; v != "" {
		composeAccessIP = v
	}
	if v := props["karaf_port"]; v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			karafPort = n
		}
	}
	if v := props["postgres_port"]; v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			postgresPort = n
		}
	}
	if v := props["keycloak_port"]; v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			keycloakPort = n
		}
	}
	if v := props["karaf_enable_admin_user"]; v != "" {
		karafEnableAdminUser = (v == "1" || strings.EqualFold(v, "true"))
	}
	if v := props["java_compiler"]; v != "" {
		javaCompiler = v
	}
}

func setupEnvironment() {
	if appName == "" {
		appName = filepath.Base(modelDir)
	}
	if schemaName == "" {
		schemaName = appName
	}
	if keycloakName == "" {
		keycloakName = appName
	}
	if runtimeEnv == "" {
		runtimeEnv = "karaf"
	}
	if dbType == "" {
		dbType = "hsqldb"
	}
	if composeEnv == "" {
		composeEnv = "compose-develop"
	}
	if karafPort == 0 {
		karafPort = 8181
	}
	if postgresPort == 0 {
		postgresPort = 5432
	}
	if keycloakPort == 0 {
		keycloakPort = 8080
	}
	// sensible defaults from the bash script
	options.StartKeycloak = true
	options.WatchBundles = true
}

func getProjectVersion() string {
	var out bytes.Buffer
	c := exec.Command("mvn",
		"org.apache.maven.plugins:maven-help-plugin:3.2.0:evaluate",
		"-Dexpression=project.version", "-q", "-DforceStdout",
	)
	c.Dir = modelDir
	c.Stdout = &out
	c.Stderr = &out
	if err := c.Run(); err != nil {
		return "SNAPSHOT"
	}
	return strings.TrimSpace(out.String())
}

func replaceInFile(path, pattern, repl string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(pattern)
	b = re.ReplaceAll(b, []byte(repl))
	return os.WriteFile(path, b, 0o644)
}

// Pick a POSIX shell on Unix without assuming bash.
func defaultShell() (prog string, args []string) {
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "sh"
	}
	return sh, []string{"-lc"}
}

// Run a small POSIX shell script on Unix (macOS/Linux).
func runShell(script string) error {
	prog, argv := defaultShell()
	argv = append(argv, script)
	return run(prog, argv...)
}

// --- WSL support (Windows) ---

func haveWSL() bool {
	_, err := exec.LookPath("wsl.exe")
	return err == nil
}

// Convert a Windows path like C:\work\proj to /mnt/c/work/proj for WSL.
func winPathToWSL(p string) string {
	if p == "" {
		return ""
	}
	p = filepath.Clean(p)
	// Expect a drive letter path like C:\...
	if len(p) >= 2 && p[1] == ':' {
		drive := strings.ToLower(string(p[0]))
		rest := strings.ReplaceAll(p[2:], `\`, `/`)
		return "/mnt/" + drive + "/" + strings.TrimPrefix(rest, "/")
	}
	// Fallback: replace backslashes
	return strings.ReplaceAll(p, `\`, `/`)
}

// Run a script inside WSL, optionally cd into the Windows cwd mapped to WSL.
func runWSL(script string, winCwd string) error {
	wslCwd := winPathToWSL(winCwd)
	if wslCwd != "" {
		script = fmt.Sprintf("cd %q && %s", wslCwd, script)
	}
	// Use a POSIX shell inside WSL
	cmd := exec.Command("wsl.exe", "sh", "-lc", script)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return cmd.Run()
}

// Run SDKMAN steps cross-platform:
// - macOS/Linux: source ~/.sdkman/bin/sdkman-init.sh then run the lines
// - Windows: if WSL exists and ~/.sdkman exists there, run inside WSL in the current project dir
func sdkmanRun(lines ...string) error {
	body := strings.Join(lines, " && ")

	if runtime.GOOS == "windows" {
		if !haveWSL() {
			fmt.Println("WSL not found — skipping SDKMAN steps.")
			return nil
		}
		// Check SDKMAN inside WSL, then run
		wd, _ := os.Getwd()
		script := fmt.Sprintf(`
if [ -f "$HOME/.sdkman/bin/sdkman-init.sh" ]; then
  . "$HOME/.sdkman/bin/sdkman-init.sh"
  %s
fi`, body)
		return runWSL(script, wd)
	}

	// Unix (macOS/Linux): source SDKMAN init if present
	home, _ := os.UserHomeDir()
	initScript := filepath.Join(home, ".sdkman", "bin", "sdkman-init.sh")
	if _, err := os.Stat(initScript); err != nil {
		// SDKMAN not installed; skip quietly
		return nil
	}
	script := fmt.Sprintf(`. %q; %s`, initScript, body)
	return runShell(script)
}

func applyInlineOptions(s string) {
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		key := strings.TrimSpace(kv[0])
		val := ""
		if len(kv) == 2 {
			val = strings.TrimSpace(kv[1])
		}
		switch key {
		case "runtime":
			runtimeEnv = val // "karaf" | "compose"
		case "dbtype":
			dbType = val // "hsqldb" | "postgresql"
		case "compose_env":
			composeEnv = val
		case "model_dir":
			if val != "" {
				if filepath.IsAbs(val) {
					modelDir = filepath.Clean(val)
				} else {
					modelDir = filepath.Clean(filepath.Join(modelDir, val))
				}
			}
		case "karaf_port":
			if n, err := strconv.Atoi(val); err == nil {
				karafPort = n
			}
		case "postgres_port":
			if n, err := strconv.Atoi(val); err == nil {
				postgresPort = n
			}
		case "keycloak_port":
			if n, err := strconv.Atoi(val); err == nil {
				keycloakPort = n
			}
		case "compose_access_ip":
			composeAccessIP = val
		case "karaf_enable_admin_user":
			karafEnableAdminUser = (val == "1" || strings.EqualFold(val, "true"))
		case "java_compiler":
			javaCompiler = val
		}
	}
}

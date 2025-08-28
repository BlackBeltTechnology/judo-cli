package config

import (
	"path/filepath"
	"strconv"
	"strings"
	"os"
	"io"
	"bufio"
)

var (
	Profile              string
	AppName              string
	ModelDir             string
	SchemaName           string
	KeycloakName         string
	KarafPort            int
	PostgresPort         int
	KeycloakPort         int
	RuntimeEnv           string
	DBType               string
	ComposeEnv           string
	ComposeAccessIP      string
	KarafEnableAdminUser bool
	JavaCompiler         string
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
	PruneFrontend bool
	PruneConfirm  bool
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

var Options JudoOptions

func DefaultConfig(cwd string) *Config {
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

func readProperties(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	props := map[string]string{}
	scanner := NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
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

func LoadProperties() {
	// default MODEL_DIR = current working dir
	wd, _ := os.Getwd()
	ModelDir = wd

	// prefer <profile>.properties, then judo.properties
	var props map[string]string
	candidates := []string{
		filepath.Join(ModelDir, Profile+".properties"),
		filepath.Join(ModelDir, "judo.properties"),
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
			ModelDir = v
		} else {
			ModelDir = filepath.Clean(filepath.Join(ModelDir, v))
		}
	}
	if v := props["app_name"]; v != "" {
		AppName = v
	}
	if v := props["schema_name"]; v != "" {
		SchemaName = v
	}
	if v := props["keycloak_name"]; v != "" {
		KeycloakName = v
	}
	if v := props["runtime"]; v != "" {
		RuntimeEnv = v
	}
	if v := props["dbtype"]; v != "" {
		DBType = v
	}
	if v := props["compose_env"]; v != "" {
		ComposeEnv = v
	}
	if v := props["compose_access_ip"]; v != "" {
		ComposeAccessIP = v
	}
	if v := props["karaf_port"]; v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			KarafPort = n
		}
	}
	if v := props["postgres_port"]; v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			PostgresPort = n
		}
	}
	if v := props["keycloak_port"]; v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			KeycloakPort = n
		}
	}
	if v := props["karaf_enable_admin_user"]; v != "" {
		KarafEnableAdminUser = (v == "1" || strings.EqualFold(v, "true"))
	}
	if v := props["java_compiler"]; v != "" {
		JavaCompiler = v
	}
}

func SetupEnvironment() {
	if AppName == "" {
		AppName = filepath.Base(ModelDir)
	}
	if SchemaName == "" {
		SchemaName = AppName
	}
	if KeycloakName == "" {
		KeycloakName = AppName
	}
	if RuntimeEnv == "" {
		RuntimeEnv = "karaf"
	}
	if DBType == "" {
		DBType = "hsqldb"
	}
	if ComposeEnv == "" {
		ComposeEnv = "compose-develop"
	}
	if KarafPort == 0 {
		KarafPort = 8181
	}
	if PostgresPort == 0 {
		PostgresPort = 5432
	}
	if KeycloakPort == 0 {
		KeycloakPort = 8080
	}
	// sensible defaults from the bash script
	Options.StartKeycloak = true
	Options.WatchBundles = true
}

func ApplyInlineOptions(s string) {
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
			RuntimeEnv = val // "karaf" | "compose"
		case "dbtype":
			DBType = val // "hsqldb" | "postgresql"
		case "compose_env":
			ComposeEnv = val
		case "model_dir":
			if val != "" {
				if filepath.IsAbs(val) {
					ModelDir = filepath.Clean(val)
				} else {
					ModelDir = filepath.Clean(filepath.Join(ModelDir, val))
				}
			}
		case "karaf_port":
			if n, err := strconv.Atoi(val); err == nil {
				KarafPort = n
			}
		case "postgres_port":
			if n, err := strconv.Atoi(val); err == nil {
				PostgresPort = n
			}
		case "keycloak_port":
			if n, err := strconv.Atoi(val); err == nil {
				KeycloakPort = n
			}
		case "compose_access_ip":
			ComposeAccessIP = val
		case "karaf_enable_admin_user":
			KarafEnableAdminUser = (val == "1" || strings.EqualFold(val, "true"))
		case "java_compiler":
			JavaCompiler = val
		}
	}
}

// NewScanner is a placeholder for bufio.NewScanner.
// This is to avoid importing bufio in config package, as it's only used here.
// The actual bufio.NewScanner will be used in the utils package.
type Scanner interface {
	Scan() bool
	Text() string
}

type bufioScanner struct {
	*bufio.Scanner
}

func NewScanner(r io.Reader) Scanner {
	return &bufioScanner{bufio.NewScanner(r)}
}

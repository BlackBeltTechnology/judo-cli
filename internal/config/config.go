package config

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	instance *Config
)

func GetConfig() *Config {
	if instance == nil {
		cwd, _ := os.Getwd()
		instance = &Config{
			AppName:      filepath.Base(cwd),
			ModelDir:     cwd,
			AppDir:       filepath.Join(cwd, "application"),
			KarafDir:     filepath.Join(cwd, "application", ".karaf"),
			Runtime:      "karaf",
			DBType:       "hsqldb",
			KarafPort:    8181,
			PostgresPort: 5432,
			KeycloakPort: 8080,
		}
		instance.SchemaName = instance.AppName
		instance.KeycloakName = instance.AppName
		instance.loadProperties()
	}
	return instance
}

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
	AppName              string
	SchemaName           string
	KeycloakName         string
	ModelDir             string
	AppDir               string
	KarafDir             string
	Runtime              string // "karaf" | "compose"
	DBType               string // "hsqldb" | "postgresql"
	ComposeEnv           string
	ComposeAccessIP      string
	KarafEnableAdminUser bool
	JavaCompiler         string
	KarafPort            int
	PostgresPort         int
	KeycloakPort         int
	Profile              string
}

var Options JudoOptions

func (c *Config) loadProperties() {
	// prefer <profile>.properties, then judo.properties
	var props map[string]string
	candidates := []string{
		filepath.Join(c.ModelDir, c.Profile+".properties"),
		filepath.Join(c.ModelDir, "judo.properties"),
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
			c.ModelDir = v
		} else {
			c.ModelDir = filepath.Clean(filepath.Join(c.ModelDir, v))
		}
	}
	if v := props["app_name"]; v != "" {
		c.AppName = v
	}
	if v := props["schema_name"]; v != "" {
		c.SchemaName = v
	}
	if v := props["keycloak_name"]; v != "" {
		c.KeycloakName = v
	}
	if v := props["runtime"]; v != "" {
		c.Runtime = v
	}
	if v := props["dbtype"]; v != "" {
		if v == "postgres" {
			c.DBType = "postgresql"
		} else {
			c.DBType = v
		}
	}
	if v := props["compose_env"]; v != "" {
		c.ComposeEnv = v
	}
	if v := props["compose_access_ip"]; v != "" {
		c.ComposeAccessIP = v
	}
	if v := props["karaf_port"]; v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			c.KarafPort = n
		}
	}
	if v := props["postgres_port"]; v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			c.PostgresPort = n
		}
	}
	if v := props["keycloak_port"]; v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			c.KeycloakPort = n
		}
	}
	if v := props["karaf_enable_admin_user"]; v != "" {
		c.KarafEnableAdminUser = (v == "1" || strings.EqualFold(v, "true"))
	}
	if v := props["java_compiler"]; v != "" {
		c.JavaCompiler = v
	}
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

func ApplyInlineOptions(s string) {
	cfg := GetConfig()
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
			cfg.Runtime = val // "karaf" | "compose"
		case "dbtype":
			if val == "postgres" {
				cfg.DBType = "postgresql"
			} else {
				cfg.DBType = val
			}
		case "compose_env":
			cfg.ComposeEnv = val
		case "model_dir":
			if val != "" {
				if filepath.IsAbs(val) {
					cfg.ModelDir = filepath.Clean(val)
				} else {
					cfg.ModelDir = filepath.Clean(filepath.Join(cfg.ModelDir, val))
				}
			}
		case "karaf_port":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.KarafPort = n
			}
		case "postgres_port":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.PostgresPort = n
			}
		case "keycloak_port":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.KeycloakPort = n
			}
		case "compose_access_ip":
			cfg.ComposeAccessIP = val
		case "karaf_enable_admin_user":
			cfg.KarafEnableAdminUser = (val == "1" || strings.EqualFold(val, "true"))
		case "java_compiler":
			cfg.JavaCompiler = val
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

// LoadProperties loads the configuration properties for the current profile
func LoadProperties() {
	// This function is called in PersistentPreRun to ensure config is loaded
	// The actual loading happens in GetConfig() via loadProperties()
	GetConfig()
}

// SetupEnvironment sets up the environment based on loaded properties
func SetupEnvironment() {
	// Environment setup is handled by the individual commands
	// This function exists for compatibility with the main.go structure
}

// Profile is the global profile variable used by the CLI
var Profile string
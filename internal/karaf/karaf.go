package karaf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"judo-cli-module/internal/config"
	"judo-cli-module/internal/utils"
)

func StopKaraf(karafDir string) {
	if karafDir == "" {
		return
	}
	_ = utils.ExecuteCommand(filepath.Join(karafDir, "bin", "stop")).Run()
}

func KarafRunning(karafDir string) bool {
	if karafDir == "" {
		return false
	}
	status := filepath.Join(karafDir, "bin", "status")
	if _, err := os.Stat(status); err != nil {
		return false
	}
	out, _ := utils.RunCapture(status)
	return strings.Contains(out, "Running")
}

func StartKaraf() {
	fmt.Println("Starting Karaf...")

	// env like in the bash
	os.Setenv("JUDO_PLATFORM_RDBMS_DIALECT", config.DBType)
	if config.DBType == "postgresql" {
		os.Setenv("JUDO_PLATFORM_RDBMS_DB_HOST", "localhost")
		os.Setenv("JUDO_PLATFORM_RDBMS_DB_PORT", fmt.Sprintf("%d", config.PostgresPort))
	}
	os.Setenv("JUDO_PLATFORM_RDBMS_DB_DATABASE", config.SchemaName)
	os.Setenv("JUDO_PLATFORM_RDBMS_DB_USER", config.SchemaName)
	os.Setenv("JUDO_PLATFORM_RDBMS_DB_PASSWORD", config.SchemaName)
	os.Setenv("JUDO_PLATFORM_KEYCLOAK_AUTH_SERVER_URL", fmt.Sprintf("http://localhost:%d/auth", config.KeycloakPort))
	if !config.Options.WatchBundles {
		os.Setenv("JUDO_PLATFORM_BUNDLE_WATCHER", "false")
	}
	os.Setenv("EXTRA_JAVA_OPTS", "-Xms1024m -Xmx1024m -Dfile.encoding=UTF-8 -Dsun.jnu.encoding=UTF-8")

	karafDir := filepath.Join(config.ModelDir, "application", ".karaf")
	_ = os.RemoveAll(karafDir)
	_ = os.MkdirAll(karafDir, 0o755)

	ver := utils.GetProjectVersion()
	tarPath := filepath.Join(config.ModelDir, "application", "karaf-offline", "target",
		fmt.Sprintf("%s-application-karaf-offline-%s.tar.gz", config.AppName, ver),
	)
	// extract
	_ = utils.ExecuteCommand("tar", "xzf", tarPath, "-C", karafDir)

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
	_ = utils.ReplaceInFile(pax, `org\.osgi\.service\.http\.port\s*=\s*\d+`, fmt.Sprintf("org.osgi.service.http.port = %d", config.KarafPort))

	// optionally enable admin user
	if config.KarafEnableAdminUser {
		users := filepath.Join(karafDir, "etc", "users.properties")
		_ = utils.ReplaceInFile(users, `#karaf\s*=\s*`, "karaf = ")
		_ = utils.ReplaceInFile(users, `#_g_/`, "_g_/")
	}

	// start in background, write logs to console.out
	consoleOut, _ := os.Create(filepath.Join(karafDir, "console.out"))
	ecmd := utils.ExecuteCommand(filepath.Join(karafDir, "bin", "karaf"), "debug", "run", "clean")
	ecmd.Stdout = consoleOut
	ecmd.Stderr = consoleOut
	_ = ecmd.Start()

	fmt.Printf("Karaf started (pid %d). Logs: %s\n", ecmd.Process.Pid, consoleOut.Name())
}

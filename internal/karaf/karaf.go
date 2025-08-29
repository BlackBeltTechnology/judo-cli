package karaf

import (
	"fmt"
	"log"
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
	// The status script outputs "Not Running ..." when stopped.
	// A simple strings.Contains("Running") is not sufficient.
	return strings.Contains(out, "Running") && !strings.Contains(out, "Not")
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
	if err := utils.UntarGz(tarPath, karafDir, 1); err != nil {
		log.Fatalf("Failed to extract Karaf archive: %v", err)
	}

	// Ensure karaf script is executable
	karafScript := filepath.Join(karafDir, "bin", "karaf")
	if _, err := os.Stat(karafScript); err == nil {
		if err := os.Chmod(karafScript, 0755); err != nil {
			log.Printf("Warning: could not make karaf script executable: %v", err)
		}
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
	ecmd := utils.ExecuteCommand(filepath.Join(karafDir, "bin", "karaf"), "run", "clean")
	ecmd.Stdout = consoleOut
	ecmd.Stderr = consoleOut
	if err := ecmd.Start(); err != nil {
		log.Fatalf("Failed to start Karaf: %v", err)
	}

	fmt.Printf("Karaf started (pid %d). Logs: %s\n", ecmd.Process.Pid, consoleOut.Name())
}

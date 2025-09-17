package help

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootHelp(t *testing.T) {
	help := RootHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "USAGE: judo COMMANDS...")
	assert.Contains(t, help, "doctor")
	assert.Contains(t, help, "build")
	assert.Contains(t, help, "start")
}

func TestStartLongHelp(t *testing.T) {
	help := StartLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Run application with postgresql and keycloak")
	assert.Contains(t, help, "runtime = karaf | compose")
	assert.Contains(t, help, "dbtype = hsqldb | postgresql")
}

func TestBuildLongHelp(t *testing.T) {
	help := BuildLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Build project")
	assert.Contains(t, help, "--build-parallel")
	assert.Contains(t, help, "--build-app-module")
}

func TestCleanLongHelp(t *testing.T) {
	help := CleanLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Stop postgresql docker container and clear data")
	assert.Contains(t, help, "Docker containers")
}

func TestPruneLongHelp(t *testing.T) {
	help := PruneLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Stop postgresql docker container and delete untracked files")
	assert.Contains(t, help, "--frontend")
	assert.Contains(t, help, "--yes")
}

func TestUpdateLongHelp(t *testing.T) {
	help := UpdateLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Update dependency versions in JUDO project")
	assert.Contains(t, help, "--ignore-checksum")
}

func TestGenerateLongHelp(t *testing.T) {
	help := GenerateLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Generate application based on model in JUDO project")
	assert.Contains(t, help, "--ignore-checksum")
}

func TestGenerateRootLongHelp(t *testing.T) {
	help := GenerateRootLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Generate application root structure based on model")
	assert.Contains(t, help, "--ignore-checksum")
}

func TestDumpLongHelp(t *testing.T) {
	help := DumpLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Dump postgresql DB data before clearing/starting application")
	assert.Contains(t, help, "postgresql")
}

func TestImportLongHelp(t *testing.T) {
	help := ImportLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Import postgresql DB data")
	assert.Contains(t, help, "--dump-name")
}

func TestSchemaUpgradeLongHelp(t *testing.T) {
	help := SchemaUpgradeLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Apply RDBMS schema upgrade using the current running database")
	assert.Contains(t, help, "postgresql")
}

func TestStopLongHelp(t *testing.T) {
	help := StopLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Stop application, postgresql and keycloak")
	assert.Contains(t, help, "karaf runtime")
}

func TestStatusLongHelp(t *testing.T) {
	help := StatusLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Print status of containers and local Karaf")
	assert.Contains(t, help, "PostgreSQL")
}

func TestRecklessLongHelp(t *testing.T) {
	help := RecklessLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Build and run project in reckless mode")
	assert.Contains(t, help, "Optimizes for speed")
}

func TestDoctorLongHelp(t *testing.T) {
	help := DoctorLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Check system health and required dependencies")
	assert.Contains(t, help, "Essential Tools")
	assert.Contains(t, help, "Docker")
}

func TestInitLongHelp(t *testing.T) {
	help := InitLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Initialize a new JUDO project or check if already initialized")
	assert.Contains(t, help, "Project GroupId")
}

func TestServerLongHelp(t *testing.T) {
	help := ServerLongHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "Start JUDO CLI web server with browser-based interface")
	assert.Contains(t, help, "--port")
}

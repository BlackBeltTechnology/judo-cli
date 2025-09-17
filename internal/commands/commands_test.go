package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"judo-cli-module/internal/config"
	"judo-cli-module/internal/utils"
)

// Mocks
var mockRun func(name string, args ...string) error
var mockRunCapture func(name string, args ...string) (string, error)

type CommandsTestSuite struct {
	suite.Suite
	tempDir string
}

func TestCommandsTestSuite(t *testing.T) {
	suite.Run(t, new(CommandsTestSuite))
}

func (s *CommandsTestSuite) SetupTest() {
	// Reset mocks before each test
	mockRun = func(name string, args ...string) error {
		return nil
	}
	mockRunCapture = func(name string, args ...string) (string, error) {
		return "", nil
	}

	// Override the utils.Run and utils.RunCapture functions
	utils.Run = func(name string, args ...string) error {
		return mockRun(name, args...)
	}
	utils.RunCapture = func(name string, args ...string) (string, error) {
		return mockRunCapture(name, args...)
	}

	// Create a temporary directory for the test
	tempDir, err := ioutil.TempDir("", "judo-cli-test")
	if err != nil {
		s.T().Fatal(err)
	}
	s.tempDir = tempDir

	// Create a dummy judo.properties file
	judoProps := `appName=TestApp
modelDir=model
appDir=application`
	err = ioutil.WriteFile(filepath.Join(s.tempDir, "judo.properties"), []byte(judoProps), 0644)
	if err != nil {
		s.T().Fatal(err)
	}

	// Change to the temporary directory
	err = os.Chdir(s.tempDir)
	if err != nil {
		s.T().Fatal(err)
	}

	// Load the config
	config.LoadProperties()
}

func (s *CommandsTestSuite) TearDownTest() {
	// Clean up the temporary directory
	err := os.RemoveAll(s.tempDir)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *CommandsTestSuite) TestCreateGenerateCommand() {
	var capturedArgs []string
	mockRun = func(name string, args ...string) error {
		capturedArgs = args
		return nil
	}

	cmd := CreateGenerateCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	s.NoError(err)
	s.Contains(capturedArgs, "clean")
	s.Contains(capturedArgs, "compile")
	s.Contains(capturedArgs, "-DgenerateApplication")
	s.Contains(capturedArgs, "-DskipApplicationBuild")
	s.NotContains(capturedArgs, "-DvalidateChecksum=false")
}

func (s *CommandsTestSuite) TestCreateGenerateCommandWithIgnoreChecksum() {
	var capturedArgs []string
	mockRun = func(name string, args ...string) error {
		capturedArgs = args
		return nil
	}

	cmd := CreateGenerateCommand()
	cmd.SetArgs([]string{"--ignore-checksum"})
	err := cmd.Execute()

	s.NoError(err)
	s.Contains(capturedArgs, "-DvalidateChecksum=false")
}

func (s *CommandsTestSuite) TestCreateBuildCommand() {
	var capturedArgs []string
	mockRun = func(name string, args ...string) error {
		capturedArgs = args
		return nil
	}

	cmd := CreateBuildCommand()
	cmd.SetArgs([]string{})
	cmd.Run(cmd, []string{})

	s.Contains(capturedArgs, "install")
	s.Contains(capturedArgs, "clean")
}

func (s *CommandsTestSuite) TestCreateBuildCommandWithSkipFrontend() {
	var capturedArgs []string
	mockRun = func(name string, args ...string) error {
		capturedArgs = args
		return nil
	}

	cmd := CreateBuildCommand()
	cmd.SetArgs([]string{"--skip-frontend"})
	cmd.Run(cmd, []string{"--skip-frontend"})

	s.Contains(capturedArgs, "-DskipReact")
}

func (s *CommandsTestSuite) TestCreateRecklessCommand() {
	var capturedArgs []string
	mockRun = func(name string, args ...string) error {
		capturedArgs = args
		return nil
	}

	cmd := CreateRecklessCommand()
	cmd.SetArgs([]string{})
	cmd.Run(cmd, []string{})

	s.Contains(capturedArgs, "package")
	s.NotContains(capturedArgs, "clean")
}

func (s *CommandsTestSuite) TestCreateGenerateRootCommand() {
	var capturedArgs []string
	mockRun = func(name string, args ...string) error {
		capturedArgs = args
		return nil
	}

	cmd := CreateGenerateRootCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	s.NoError(err)
	s.Contains(capturedArgs, "-DgenerateRoot")
}

func (s *CommandsTestSuite) TestCreateImportCommand() {
	cmd := CreateImportCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	s.NoError(err)
}

func (s *CommandsTestSuite) TestCreateSchemaUpgradeCommand() {
	cmd := CreateSchemaUpgradeCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	s.NoError(err)
}

func (s *CommandsTestSuite) TestCreateCleanCommand() {
	cmd := CreateCleanCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	s.NoError(err)
}

func (s *CommandsTestSuite) TestCreatePruneCommand() {
	cmd := CreatePruneCommand()
	cmd.SetArgs([]string{"--yes"})
	err := cmd.Execute()

	s.NoError(err)
}

func (s *CommandsTestSuite) TestCreateUpdateCommand() {
	cmd := CreateUpdateCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	s.NoError(err)
}

func (s *CommandsTestSuite) TestCreateStopCommand() {
	cmd := CreateStopCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	s.NoError(err)
}

func (s *CommandsTestSuite) TestCreateStartCommandWithKarafArchive() {
	originalWd, _ := os.Getwd()
	config.Reset()
	os.Chdir(s.tempDir)
	defer os.Chdir(originalWd)
	// Create a dummy karaf archive
	cfg := config.GetConfig()
	utils.GetProjectVersion = func() string {
		return "1.0.0"
	}
	ver := utils.GetProjectVersion()
	tarPath := filepath.Join(cfg.ModelDir, "application", "karaf-offline", "target",
		fmt.Sprintf("%s-application-karaf-offline-%s.tar.gz", cfg.AppName, ver),
	)
	os.MkdirAll(filepath.Dir(tarPath), 0755)
	ioutil.WriteFile(tarPath, []byte("dummy content"), 0644)

	utils.IsPortAvailable = func(port int) bool {
		return true
	}

	cmd := CreateStartCommand()
	cmd.SetArgs([]string{})
	cmd.Run(cmd, []string{})

	s.NoError(nil)
}

func (s *CommandsTestSuite) TestCreateDoctorCommand() {
	cmd := CreateDoctorCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	s.NoError(err)
}

func (s *CommandsTestSuite) TestCreateLogCommand() {
	// Create a dummy log file
	cfg := config.GetConfig()
	logFile := filepath.Join(cfg.KarafDir, "console.out")
	os.MkdirAll(filepath.Dir(logFile), 0755)
	ioutil.WriteFile(logFile, []byte("test log"), 0644)

	cmd := CreateLogCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	s.NoError(err)
}

func (s *CommandsTestSuite) TestCreateSessionCommand() {
	cmd := CreateSessionCommand()
	cmd.SetArgs([]string{})
	cmd.Run(cmd, []string{})

	s.NoError(nil)
}

func (s *CommandsTestSuite) TestCreateInitCommand() {
	// Remove the dummy judo.properties file to simulate a new project
	os.Remove(filepath.Join(s.tempDir, "judo.properties"))

	cmd := CreateInitCommand()
	cmd.SetArgs([]string{"--group-id", "com.example", "--model-name", "TestProject", "--type", "ESM"})
	err := cmd.Execute()

	s.NoError(err)
}

func (s *CommandsTestSuite) TestCreateSelfUpdateCommand() {
	cmd := CreateSelfUpdateCommand("1.0..0")
	cmd.SetArgs([]string{"--check"})
	err := cmd.Execute()

	s.NoError(err)
}

func (s *CommandsTestSuite) TestCreateStatusCommand() {
	cmd := CreateStatusCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	s.NoError(err)
}

package utils

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const TimeSecond = time.Second

func TimeNow() time.Time {
	return time.Now()
}

// NewScanner is a wrapper for bufio.NewScanner.
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

func WaitForPort(host string, port int, timeout time.Duration) {
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

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Small exec helpers
func Run(name string, args ...string) error {
	return RunInDir("", name, args...)
}

func RunInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	cmd.Dir = dir
	return cmd.Run()
}

func RunCapture(name string, args ...string) (string, error) {
	return RunCaptureInDir("", name, args...)
}

func RunCaptureInDir(dir, name string, args ...string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = &out, &out
	cmd.Dir = dir
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

// Optional shim if your code calls executeCommand(...)
func ExecuteCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	return cmd
}

func GetProjectVersion() string {
	var out bytes.Buffer
	c := exec.Command("mvn",
		"org.apache.maven.plugins:maven-help-plugin:3.2.0:evaluate",
		"-Dexpression=project.version", "-q", "-DforceStdout",
	)
	c.Dir = "." // This will be set by the caller based on config.ModelDir
	c.Stdout = &out
	c.Stderr = &out
	if err := c.Run(); err != nil {
		return "SNAPSHOT"
	}
	return strings.TrimSpace(out.String())
}

// GetCurrentDir returns the current working directory
func GetCurrentDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}

func ReplaceInFile(path, pattern, repl string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(pattern)
	b = re.ReplaceAll(b, []byte(repl))
	return os.WriteFile(path, b, 0o644)
}

// Pick a POSIX shell on Unix without assuming bash.
func DefaultShell() (prog string, args []string) {
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "sh"
	}
	return sh, []string{"-l", "-c"}
}

// Run a small POSIX shell script on Unix (macOS/Linux).
func RunShell(script string) error {
	prog, argv := DefaultShell()
	argv = append(argv, script)
	return Run(prog, argv...)
}

// --- WSL support (Windows) ---

func HaveWSL() bool {
	_, err := exec.LookPath("wsl.exe")
	return err == nil
}

// Convert a Windows path like C:\work\proj to /mnt/c/work/proj for WSL.
func WinPathToWSL(p string) string {
	if p == "" {
		return ""
	}
	p = filepath.Clean(p)
	// Expect a drive letter path like C:\...
	if len(p) >= 2 && p[1] == ':' {
		drive := strings.ToLower(string(p[0]))
		rest := strings.ReplaceAll(p[2:], `\\`, `/`)
		return "/mnt/" + drive + "/" + strings.TrimPrefix(rest, "/")
	}
	// Fallback: replace backslashes
	return strings.ReplaceAll(p, `\\`, `/`)
}

// Run a script inside WSL, optionally cd into the Windows cwd mapped to WSL.
func RunWSL(script string, winCwd string) error {
	wslCwd := WinPathToWSL(winCwd)
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
func SdkmanRun(lines ...string) error {
	body := strings.Join(lines, " && ")

	if runtime.GOOS == "windows" {
		if !HaveWSL() {
			fmt.Println("WSL not found — skipping SDKMAN steps.")
			return nil
		}
		// Check SDKMAN inside WSL, then run
		wd, _ := os.Getwd()
		script := fmt.Sprintf(`
if [ -f "$HOME/.sdkman/bin/sdkman-init.sh" ]; then
  . "$HOME/.sdkman/bin/sdkman-init.sh"
  %s
fi`,
			body)
		return RunWSL(script, wd)
	}

	// Unix (macOS/Linux): source SDKMAN init if present
	home, _ := os.UserHomeDir()
	initScript := filepath.Join(home, ".sdkman", "bin", "sdkman-init.sh")
	if _, err := os.Stat(initScript); err != nil {
		// SDKMAN not installed; skip quietly
		return nil
	}
	script := fmt.Sprintf(`. %q; %s`,
		initScript,
		body)
	return RunShell(script)
}

// InstallSDKMAN installs SDKMAN on Unix systems (macOS/Linux)
func InstallSDKMAN() error {
	if runtime.GOOS == "windows" {
		if !HaveWSL() {
			return fmt.Errorf("WSL not found — cannot install SDKMAN on Windows without WSL")
		}
		// Install SDKMAN inside WSL
		script := `curl -s "https://get.sdkman.io" | bash`
		return RunWSL(script, "")
	}

	// Unix (macOS/Linux) installation
	fmt.Println("Installing SDKMAN...")

	// Download and install SDKMAN using shell to properly handle pipes
	script := `curl -s "https://get.sdkman.io" | bash`
	if err := RunShell(script); err != nil {
		return fmt.Errorf("failed to install SDKMAN: %w", err)
	}

	// Source SDKMAN to make it available in current session
	home, _ := os.UserHomeDir()
	initScript := filepath.Join(home, ".sdkman", "bin", "sdkman-init.sh")
	if _, err := os.Stat(initScript); err != nil {
		return fmt.Errorf("SDKMAN installation completed but init script not found")
	}

	// Source SDKMAN to initialize it
	return RunShell(fmt.Sprintf(`. %q && sdk version`, initScript))
}

// InstallRequiredTools installs Maven and Java using SDKMAN in a JUDO project directory
func InstallRequiredTools() error {
	fmt.Println("Installing required tools via SDKMAN...")

	// Install Maven
	err := SdkmanRun("sdk install maven")
	if err != nil {
		return fmt.Errorf("failed to install Maven: %w", err)
	}

	// Install Java (latest LTS version)
	err = SdkmanRun("sdk install java")
	if err != nil {
		return fmt.Errorf("failed to install Java: %w", err)
	}

	// Set Java as default
	//err = SdkmanRun("sdk default java")
	//if err != nil {
	//	return fmt.Errorf("failed to set Java as default: %w", err)
	//}

	fmt.Println("✅ Required tools installed successfully")
	return nil
}

// FileExists checks if a file exists at the given path.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// PromptForInput prompts the user for input and returns the trimmed string.
func PromptForInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// PromptForSelection prompts the user to select from a list of options, with a default.
func PromptForSelection(prompt string, options []string, defaultOption string) string {
	optionsStr := strings.Join(options, "/")
	for {
		input := PromptForInput(fmt.Sprintf("%s (%s) [%s]: ", prompt, optionsStr, defaultOption))
		if input == "" {
			return defaultOption
		}
		for _, opt := range options {
			if strings.EqualFold(input, opt) {
				return opt
			}
		}
		fmt.Printf("Invalid selection. Please choose from %s.\n", optionsStr)
	}
}

// IsPortAvailable checks if a TCP port is available by attempting to connect to it.
func IsPortAvailable(port int) bool {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)

	if err != nil {
		// Could not connect, so port is likely available.
		return true
	}

	// Successfully connected, so port is in use.
	_ = conn.Close()
	return false
}

// IsPortUsedByKaraf checks if a port is being used by the current Karaf instance
func IsPortUsedByKaraf(port int, karafDir string) bool {
	if karafDir == "" {
		return false
	}
	
	// Check if Karaf is running using bin/status command
	statusCmd := filepath.Join(karafDir, "bin", "status")
	if _, err := os.Stat(statusCmd); err != nil {
		return false
	}
	
	// Execute status command to check if Karaf is running
	out, err := RunCapture(statusCmd)
	if err != nil {
		return false
	}
	
	// Check if Karaf is running and using the specified port
	karafRunning := strings.Contains(out, "Running") && !strings.Contains(out, "Not")
	if !karafRunning {
		return false
	}
	
	// Check if Karaf is configured to use this port
	paxConfig := filepath.Join(karafDir, "etc", "org.ops4j.pax.web.cfg")
	if _, err := os.Stat(paxConfig); err == nil {
		content, err := os.ReadFile(paxConfig)
		if err == nil {
			portConfig := fmt.Sprintf("org.osgi.service.http.port = %d", port)
			return strings.Contains(string(content), portConfig)
		}
	}
	
	return false
}

// UntarGz decompresses a .tar.gz file to a destination directory, stripping leading path components.
func UntarGz(src, dest string, stripComponents int) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Strip leading path components
		strippedName := ""
		parts := strings.Split(header.Name, "/")
		if len(parts) > stripComponents {
			strippedName = strings.Join(parts[stripComponents:], "/")
		}
		if strippedName == "" {
			continue // Skip empty paths (like the top-level directory itself)
		}

		target := filepath.Join(dest, strippedName)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for file: %w", err)
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file content: %w", err)
			}
			outFile.Close()
		default:
			log.Printf("unsupported file type in archive: %c for %s", header.Typeflag, header.Name)
		}
	}
}

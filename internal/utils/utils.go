package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"io"
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


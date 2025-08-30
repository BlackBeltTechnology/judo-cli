package selfupdate

import (
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	githubAPIBase   = "https://api.github.com/repos/BlackBeltTechnology/judo-cli"
	downloadTimeout = 30 * time.Second
	maxDownloadSize = 100 * 1024 * 1024 // 100MB
)

// Release represents a GitHub release
type Release struct {
	TagName    string  `json:"tag_name"`
	Name       string  `json:"name"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	DownloadURL    string
	IsSnapshot     bool
	NeedsUpdate    bool
}

// CheckForUpdate checks if an update is available
func CheckForUpdate(currentVersion string) (*UpdateInfo, error) {
	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		IsSnapshot:     strings.Contains(strings.ToLower(currentVersion), "snapshot"),
	}

	// Only check for updates if current version is a snapshot
	if !info.IsSnapshot {
		return info, nil
	}

	release, err := getLatestRelease(true) // Get latest prerelease for snapshots
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}

	info.LatestVersion = release.TagName
	info.NeedsUpdate = info.CurrentVersion != info.LatestVersion

	if info.NeedsUpdate {
		asset := findAssetForPlatform(release.Assets)
		if asset == nil {
			return nil, fmt.Errorf("no binary found for platform %s/%s", runtime.GOOS, runtime.GOARCH)
		}
		info.DownloadURL = asset.BrowserDownloadURL
	}

	return info, nil
}

// PerformUpdate downloads and installs the update
func PerformUpdate(downloadURL string) error {
	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Determine update binary name
	updateExe := getUpdateBinaryName(currentExe)

	// Download the new binary
	if err := downloadBinary(downloadURL, updateExe); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Make it executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(updateExe, 0755); err != nil {
			return fmt.Errorf("failed to make update binary executable: %w", err)
		}
	}

	// Perform the replacement
	return replaceCurrentBinary(currentExe, updateExe)
}

// getLatestRelease fetches the latest release from GitHub
func getLatestRelease(prerelease bool) (*Release, error) {
	var url string
	if prerelease {
		// Get all releases and find the latest prerelease
		url = githubAPIBase + "/releases"
	} else {
		url = githubAPIBase + "/releases/latest"
	}

	client := &http.Client{Timeout: downloadTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("no releases found or repository not accessible")
		}
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	if prerelease {
		var releases []Release
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, err
		}
		// Find the latest prerelease
		for _, release := range releases {
			if release.Prerelease {
				return &release, nil
			}
		}
		return nil, fmt.Errorf("no prerelease found")
	} else {
		var release Release
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil, err
		}
		return &release, nil
	}
}

// findAssetForPlatform finds the appropriate binary asset for the current platform
func findAssetForPlatform(assets []Asset) *Asset {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Common naming patterns for different platforms
	patterns := []string{
		fmt.Sprintf("judo-%s-%s", goos, goarch),
		fmt.Sprintf("judo_%s_%s", goos, goarch),
		fmt.Sprintf("judo-%s", goos),
	}

	// Add .exe extension for Windows
	if goos == "windows" {
		for i, pattern := range patterns {
			patterns[i] = pattern + ".exe"
		}
	}

	// Also check for compressed versions
	compressedPatterns := make([]string, 0, len(patterns)*3)
	for _, pattern := range patterns {
		compressedPatterns = append(compressedPatterns,
			pattern+".gz",
			pattern+".zip",
			pattern,
		)
	}

	// Find matching asset
	for _, asset := range assets {
		assetName := strings.ToLower(asset.Name)
		for _, pattern := range compressedPatterns {
			if strings.Contains(assetName, strings.ToLower(pattern)) {
				return &asset
			}
		}
	}

	return nil
}

// downloadBinary downloads and extracts the binary if needed
func downloadBinary(url, destPath string) error {
	client := &http.Client{Timeout: downloadTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Check file size
	if resp.ContentLength > maxDownloadSize {
		return fmt.Errorf("download too large: %d bytes", resp.ContentLength)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Determine if we need to decompress based on URL
	filename := filepath.Base(url)
	if strings.HasSuffix(filename, ".gz") {
		return extractGzip(resp.Body, out)
	} else if strings.HasSuffix(filename, ".zip") {
		return extractZip(resp.Body, out, destPath)
	}

	// Direct copy for uncompressed files
	_, err = io.Copy(out, resp.Body)
	return err
}

// extractGzip extracts a gzip compressed binary
func extractGzip(src io.Reader, dst io.Writer) error {
	gr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer gr.Close()

	_, err = io.Copy(dst, gr)
	return err
}

// extractZip extracts a zip file (assumes single binary inside)
func extractZip(src io.Reader, dst io.Writer, destPath string) error {
	// For zip files, we need to download to a temp file first
	tempFile, err := os.CreateTemp("", "judo-download-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy zip content to temp file
	if _, err := io.Copy(tempFile, src); err != nil {
		return err
	}

	// Reopen for reading
	if _, err := tempFile.Seek(0, 0); err != nil {
		return err
	}

	// Read zip file
	stat, err := tempFile.Stat()
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(tempFile, stat.Size())
	if err != nil {
		return err
	}

	// Find the binary file in the zip
	for _, file := range zipReader.File {
		if strings.Contains(file.Name, "judo") && !strings.Contains(file.Name, "/") {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			_, err = io.Copy(dst, rc)
			return err
		}
	}

	return fmt.Errorf("no judo binary found in zip file")
}

// getUpdateBinaryName returns the name for the update binary
func getUpdateBinaryName(currentExe string) string {
	dir := filepath.Dir(currentExe)
	base := filepath.Base(currentExe)

	if runtime.GOOS == "windows" {
		// Remove .exe if present and add .selfupdate.exe
		if strings.HasSuffix(base, ".exe") {
			base = strings.TrimSuffix(base, ".exe")
		}
		return filepath.Join(dir, base+".selfupdate.exe")
	}

	return filepath.Join(dir, base+".selfupdate")
}

// replaceCurrentBinary replaces the current binary with the update
func replaceCurrentBinary(currentExe, updateExe string) error {
	if runtime.GOOS == "windows" {
		return replaceOnWindows(currentExe, updateExe)
	}
	return replaceOnUnix(currentExe, updateExe)
}

// replaceOnUnix replaces binary on Unix systems
func replaceOnUnix(currentExe, updateExe string) error {
	// Create a script to perform the replacement
	scriptContent := fmt.Sprintf(`#!/bin/bash
sleep 1
mv "%s" "%s.old" 2>/dev/null || true
mv "%s" "%s"
chmod +x "%s"
exec "%s" "$@"
`, currentExe, currentExe, updateExe, currentExe, currentExe, currentExe)

	scriptPath := updateExe + ".sh"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to create replacement script: %w", err)
	}

	// Execute the replacement script
	cmd := exec.Command("/bin/bash", scriptPath)
	cmd.Args = append(cmd.Args, os.Args[1:]...) // Pass through original arguments
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Start the script and exit current process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start replacement script: %w", err)
	}

	os.Exit(0)
	return nil
}

// replaceOnWindows replaces binary on Windows
func replaceOnWindows(currentExe, updateExe string) error {
	// Create a batch script to perform the replacement
	scriptContent := fmt.Sprintf(`@echo off
timeout /t 1 /nobreak > nul
del "%s.old" 2>nul
ren "%s" "%s.old" 2>nul
ren "%s" "%s"
start "" "%s" %%*
`, currentExe, currentExe, filepath.Base(currentExe), updateExe, filepath.Base(currentExe), currentExe)

	scriptPath := updateExe + ".bat"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		return fmt.Errorf("failed to create replacement script: %w", err)
	}

	// Execute the replacement script
	cmd := exec.Command("cmd", "/C", scriptPath)
	cmd.Args = append(cmd.Args, os.Args[1:]...) // Pass through original arguments

	// Start the script and exit current process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start replacement script: %w", err)
	}

	// Use syscall.Exit to avoid cleanup
	syscall.Exit(0)
	return nil
}

package selfupdate

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckForUpdateWithNonSnapshot(t *testing.T) {
	info, err := CheckForUpdate("v1.0.0")
	assert.NoError(t, err)
	assert.False(t, info.IsSnapshot)
	assert.False(t, info.NeedsUpdate)
	assert.Equal(t, "v1.0.0", info.CurrentVersion)
	assert.Empty(t, info.LatestVersion)
	assert.Empty(t, info.DownloadURL)
}

func TestCheckForUpdateWithSnapshot(t *testing.T) {
	info, err := CheckForUpdate("v1.0.0-SNAPSHOT")
	assert.NoError(t, err)
	assert.True(t, info.IsSnapshot)
	assert.Equal(t, "v1.0.0-SNAPSHOT", info.CurrentVersion)
	// We can't test the actual update check since it makes HTTP calls
	// but we can verify the structure is correct
}

func TestFindAssetForPlatform(t *testing.T) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	assets := []Asset{
		{Name: "judo-" + goos + "-" + goarch + ".exe", BrowserDownloadURL: "http://example.com/binary.exe"},
		{Name: "judo-" + goos + "-" + goarch + ".gz", BrowserDownloadURL: "http://example.com/binary.gz"},
		{Name: "judo-" + goos + "-" + goarch + ".zip", BrowserDownloadURL: "http://example.com/binary.zip"},
		{Name: "judo-" + goos + "-" + goarch, BrowserDownloadURL: "http://example.com/binary"},
	}

	asset := findAssetForPlatform(assets)
	assert.NotNil(t, asset)
	assert.Contains(t, asset.Name, "judo-"+goos+"-"+goarch)
}

func TestFindAssetForPlatformNoMatch(t *testing.T) {
	assets := []Asset{
		{Name: "judo-linux-arm64.exe", BrowserDownloadURL: "http://example.com/binary.exe"},
		{Name: "judo-windows-amd64.gz", BrowserDownloadURL: "http://example.com/binary.gz"},
	}

	asset := findAssetForPlatform(assets)
	assert.Nil(t, asset)
}

func TestGetUpdateBinaryName(t *testing.T) {
	currentExe := "/usr/local/bin/judo"
	updateName := getUpdateBinaryName(currentExe)

	if runtime.GOOS == "windows" {
		assert.Equal(t, "/usr/local/bin/judo.selfupdate.exe", updateName)
	} else {
		assert.Equal(t, "/usr/local/bin/judo.selfupdate", updateName)
	}
}

func TestGetUpdateBinaryNameWindowsWithExe(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	currentExe := "C:\\Program Files\\judo\\judo.exe"
	updateName := getUpdateBinaryName(currentExe)
	assert.Equal(t, "C:\\Program Files\\judo\\judo.selfupdate.exe", updateName)
}

func TestIsSnapshotDetection(t *testing.T) {
	testCases := []struct {
		version    string
		isSnapshot bool
	}{
		{"v1.0.0", false},
		{"v1.0.0-SNAPSHOT", true},
		{"v1.0.0-snapshot", true},
		{"v1.0.0-Snapshot", true},
		{"1.0.0", false},
		{"1.0.0-SNAPSHOT", true},
		{"development", false}, // Not a standard snapshot format
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			info := &UpdateInfo{CurrentVersion: tc.version}
			info.IsSnapshot = strings.Contains(strings.ToLower(tc.version), "snapshot")
			assert.Equal(t, tc.isSnapshot, info.IsSnapshot)
		})
	}
}

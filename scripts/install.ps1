#!/usr/bin/env pwsh
#Requires -Version 7.0

# JUDO CLI Installer for Windows
# Downloads and installs the latest JUDO CLI binary from GitHub releases

# ---------- Configuration ----------
$Repo = "BlackBeltTechnology/judo-cli"
$Project = "judo-cli"
$BinaryName = "judo.exe"

# Optional overrides from environment
$Version = if ($env:JUDO_VERSION) { $env:JUDO_VERSION } else { "latest" }
$InstallDir = if ($env:JUDO_INSTALL_DIR) { $env:JUDO_INSTALL_DIR } else { "" }

# ---------- OS/Arch detection ----------
$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { 
        Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
        exit 1
    }
}

# ---------- Install destination ----------
if ($InstallDir) {
    $dest = $InstallDir
} elseif (Test-Path "C:\Program Files") {
    $dest = "C:\Program Files\JUDO-CLI\bin"
} else {
    $dest = "$HOME\AppData\Local\JUDO-CLI\bin"
}

# Create destination directory if it doesn't exist
if (!(Test-Path $dest)) {
    New-Item -ItemType Directory -Path $dest -Force | Out-Null
}

# ---------- Build download URL ----------
if ($Version -eq "latest") {
    $asset = "${Project}_windows_${arch}.zip"
    $url = "https://github.com/${Repo}/releases/latest/download/${asset}"
} else {
    $asset = "${Project}_${Version}_windows_${arch}.zip"
    $url = "https://github.com/${Repo}/releases/download/${Version}/${asset}"
}

# ---------- Download + extract ----------
$tempDir = Join-Path $env:TEMP "judo-install-$(Get-Random)"
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

try {
    Write-Host "Downloading $url"
    $archive = Join-Path $tempDir "pkg.zip"
    
    # Download with progress
    Invoke-WebRequest -Uri $url -OutFile $archive -UseBasicParsing
    
    # Extract archive
    Expand-Archive -Path $archive -DestinationPath $tempDir -Force
    
    # Find the binary
    $src = Get-ChildItem -Path $tempDir -Recurse -Filter $BinaryName | Select-Object -First 1
    if (!$src) {
        Write-Error "Binary '$BinaryName' not found in archive"
        exit 1
    }
    
    # ---------- Move to destination ----------
    $installPath = Join-Path $dest $BinaryName
    Move-Item -Path $src.FullName -Destination $installPath -Force
    
    # ---------- Add to PATH if needed ----------
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$dest*") {
        $newPath = "$dest;$currentPath"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Host "Added $dest to user PATH. Restart your terminal or run: `$env:PATH = \"$dest;`$env:PATH\""
    }
    
    # ---------- Verify installation ----------
    if (Test-Path $installPath) {
        Write-Host " Installed $BinaryName to $dest" -ForegroundColor Green
        try {
            & $installPath --version
        } catch {
            Write-Host "Installed to $installPath, but version check failed. Ensure it runs on your platform." -ForegroundColor Yellow
        }
    } else {
        Write-Error "Installation failed - binary not found at destination"
        exit 1
    }
} finally {
    # Cleanup
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}
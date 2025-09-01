#!/usr/bin/env bash
set -euo pipefail

# ---------- Config (change owner/repo + binary name if needed) ----------
REPO="BlackBeltTechnology/judo-cli"             # e.g., acme/judo-cli
PROJECT="judo"                    # archive base name inside Releases (matches GoReleaser project_name)
BINARY_NAME="judo"                # installed command name (no extension)

# Optional overrides from env: JUDO_VERSION (e.g. v1.2.3), JUDO_INSTALL_DIR
VERSION="${JUDO_VERSION:-latest}"
INSTALL_DIR="${JUDO_INSTALL_DIR:-}"

# ---------- Prereqs ----------
need() { command -v "$1" >/dev/null 2>&1 || { echo "Missing dependency: $1"; exit 1; }; }
if command -v curl >/dev/null 2>&1; then DL="curl -fL --proto '=https' --tlsv1.2"; need tar; else
  need wget; need tar; DL="wget -qO-"; fi

# ---------- OS/Arch detection ----------
uname_s=$(uname -s)
case "$uname_s" in
  Linux)   os="Linux";;
  Darwin)  os="Darwin";;
  *) echo "Unsupported OS: $uname_s"; exit 1;;
 esac

uname_m=$(uname -m)
case "$uname_m" in
  x86_64|amd64) arch="x86_64";;
  arm64|aarch64) arch="arm64";;
  i386|i686) echo "Unsupported architecture: $uname_m (386 is not supported)"; exit 1;;
  *) echo "Unsupported architecture: $uname_m"; exit 1;;
 esac

# ---------- Install destination ----------
if [[ -n "$INSTALL_DIR" ]]; then dest="$INSTALL_DIR";
elif [[ -w "/usr/local/bin" ]]; then dest="/usr/local/bin";
else dest="$HOME/.local/bin"; fi
mkdir -p "$dest"

# ---------- Build download URL ----------
if [[ "$VERSION" == "latest" ]]; then
  asset="${PROJECT}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/latest/download/${asset}"
else
  asset="${PROJECT}_${VERSION}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${VERSION}/${asset}"
fi

# ---------- Download + extract ----------
workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT

archive="$workdir/pkg.tgz"
if [[ "$DL" == curl* ]]; then
  echo "Downloading $url"
  curl -fL --proto '=https' --tlsv1.2 "$url" -o "$archive"
else
  echo "Downloading $url"
  wget -q "$url" -O "$archive"
fi

tar -xzf "$archive" -C "$workdir"

# find the binary (either at root or in a folder)
src="$(find "$workdir" -maxdepth 3 -type f -name "$BINARY_NAME" -perm -u+x | head -n1)"
if [[ -z "$src" ]]; then echo "Binary '$BINARY_NAME' not found in archive"; exit 1; fi
chmod +x "$src"

# ---------- Move + PATH handling ----------
install_path="$dest/$BINARY_NAME"
mv -f "$src" "$install_path"

if ! command -v "$install_path" >/dev/null 2>&1 && ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
  # ensure PATH for future shells
  if [[ -n "${ZSH_VERSION-}" ]]; then rc="$HOME/.zshrc"; elif [[ -n "${BASH_VERSION-}" ]]; then rc="$HOME/.bashrc"; else rc="$HOME/.profile"; fi
  echo "export PATH=\"$dest:\$PATH\"" >> "$rc"
  echo "Added $dest to PATH in $rc. Restart your shell or run: export PATH=\"$dest:$PATH\""
fi

# ---------- Verify ----------
if "$install_path" --version >/dev/null 2>&1; then
  echo "✅ Installed $BINARY_NAME to $dest"
  "$install_path" --version
else
  echo "Installed to $install_path, but version check failed. Ensure it runs on your platform."
fi

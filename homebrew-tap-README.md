# BlackBelt Technology Homebrew Tap

This tap contains Homebrew formulas for BlackBelt Technology tools.

## Usage

```bash
# Add the tap
brew tap blackbelttechnology/tap

# Install tools
brew install judo

# Upgrade tools
brew upgrade judo
```

## Available Formulas

- **judo** - JUDO CLI for managing JUDO application lifecycle

## Tools

### JUDO CLI

The JUDO CLI is a command-line tool for managing the complete lifecycle of JUDO applications.

**Features:**
- Application management and deployment
- Docker container operations
- Karaf server management
- Database operations
- Session management
- Self-update capabilities

**Installation:**
```bash
brew install blackbelttechnology/tap/judo
```

**Verification:**
```bash
# Check installation
judo version

# Test basic functionality
judo doctor
```

## How It Works

This tap is automatically maintained through GitHub Actions:

1. **Release Process**: When a new release is created in the [judo-cli repository](https://github.com/BlackBeltTechnology/judo-cli), GoReleaser:
   - Builds binaries for multiple platforms (macOS Intel/Apple Silicon, Linux x86_64/arm64)
   - Creates release assets on GitHub
   - Automatically updates the Homebrew formula in this repository

2. **Formula Generation**: The formula file (`Formula/judo.rb`) is automatically generated with:
   - Download URLs for the release assets
   - SHA256 checksums for verification
   - Installation and test instructions
   - Dependencies (Docker and Maven as optional)

3. **User Installation**: Users can install via Homebrew as shown above

## Manual Installation (Alternative)

If you prefer not to use Homebrew, you can download binaries directly from the [JUDO CLI releases page](https://github.com/BlackBeltTechnology/judo-cli/releases).

## Troubleshooting

### Common Issues

1. **Tap not found**: Ensure you've added the tap correctly:
   ```bash
   brew tap blackbelttechnology/tap
   ```

2. **Formula not found**: The formula may not exist yet if no releases have been created

3. **Permission issues**: Ensure you have write permissions to `/usr/local` or use `brew install --build-from-source`

### Testing Formula Locally

```bash
# Clone this tap
brew tap blackbelttechnology/tap

# Install from tap
brew install judo

# Or install specific version
brew install blackbelttechnology/tap/judo@0.1.5

# Upgrade to latest
brew upgrade judo
```

## Support

- **Issues**: Report issues on the [judo-cli repository](https://github.com/BlackBeltTechnology/judo-cli/issues)
- **Documentation**: Full documentation available at [judo-cli documentation](https://blackbelttechnology.github.io/judo-cli/)

## License

All formulas in this tap are distributed under the same licenses as their respective projects.

---

*This tap is automatically maintained by GitHub Actions and GoReleaser.*
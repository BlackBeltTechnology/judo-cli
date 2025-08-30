# Homebrew Setup Guide

This document explains the steps needed to set up Homebrew distribution for the JUDO CLI.

## Overview

The JUDO CLI is configured to automatically publish to a Homebrew tap when releases are created. This provides an easy installation method for macOS and Linux users.

## Required Steps

### 1. Create the Homebrew Tap Repository

Create a new GitHub repository named `homebrew-tap` under the BlackBeltTechnology organization:

```bash
# Repository URL should be:
https://github.com/BlackBeltTechnology/homebrew-tap
```

**Repository Settings:**
- Name: `homebrew-tap`
- Description: "Homebrew tap for BlackBelt Technology tools"
- Public repository (required for Homebrew)
- Initialize with README

### 2. Set Up GitHub Token

Create a GitHub Personal Access Token (PAT) for the automated formula updates:

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Set the following:
   - **Note**: "JUDO CLI Homebrew Tap"
   - **Expiration**: No expiration (or long-term)
   - **Scopes**: 
     - `public_repo` (for public repositories)
     - `workflow` (for GitHub Actions)

### 3. Configure Repository Secrets

In the **judo-cli** repository, add the GitHub token as a secret:

1. Go to Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add:
   - **Name**: `HOMEBREW_TAP_TOKEN`
   - **Value**: The PAT created in step 2

### 4. Repository Structure

The homebrew-tap repository will automatically be populated with formula files when releases are created. The structure will look like:

```
homebrew-tap/
├── README.md
└── Formula/
    └── judo.rb
```

### 5. Initial Tap Setup (Optional)

Create an initial README.md in the homebrew-tap repository:

```markdown
# BlackBelt Technology Homebrew Tap

This tap contains Homebrew formulas for BlackBelt Technology tools.

## Usage

```bash
# Add the tap
brew tap blackbelttechnology/tap

# Install tools
brew install judo
```

## Available Formulas

- **judo** - JUDO CLI for managing JUDO application lifecycle
```

## How It Works

1. **Release Process**: When a new release is created in the judo-cli repository, GoReleaser:
   - Builds binaries for multiple platforms
   - Creates release assets
   - Automatically updates the Homebrew formula in `homebrew-tap`

2. **Formula Generation**: The formula file (`Formula/judo.rb`) is automatically generated with:
   - Download URLs for the release assets
   - SHA256 checksums for verification
   - Installation and test instructions
   - Dependencies (Docker and Maven as optional)

3. **User Installation**: Users can then install via:
   ```bash
   brew tap blackbelttechnology/tap
   brew install judo
   ```

## Testing the Setup

After the first release, test the Homebrew installation:

```bash
# Add the tap
brew tap blackbelttechnology/tap

# Install JUDO CLI
brew install judo

# Verify installation
judo version

# Test basic functionality
judo doctor
```

## Maintenance

- **Automatic Updates**: Formula updates are automatic with each release
- **Manual Updates**: If needed, formula can be manually edited in the homebrew-tap repository
- **Token Rotation**: Update the `HOMEBREW_TAP_TOKEN` secret when the GitHub token expires

## Troubleshooting

### Common Issues

1. **Token Permissions**: Ensure the GitHub token has `public_repo` and `workflow` scopes
2. **Repository Access**: The token user must have write access to the homebrew-tap repository
3. **Formula Validation**: Homebrew validates formulas automatically - check for syntax errors
4. **Naming Conflicts**: Ensure the formula name `judo` doesn't conflict with existing formulas

### Testing Formula Locally

```bash
# Clone the tap
git clone https://github.com/BlackBeltTechnology/homebrew-tap.git

# Test the formula
brew install --build-from-source ./homebrew-tap/Formula/judo.rb

# Or test without installing
brew audit --strict homebrew-tap/Formula/judo.rb
```

## Documentation Links

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [GoReleaser Homebrew Documentation](https://goreleaser.com/customization/homebrew/)
- [GitHub Token Documentation](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
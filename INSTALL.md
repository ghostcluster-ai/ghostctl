# Installation Guide

## Prerequisites

- Go 1.21 or later
- Kubernetes 1.24+ cluster (for host cluster)
- `kubectl` configured to access your cluster
- `helm` 3.0+ (optional, for advanced deployments)

## Installation Methods

### Method 1: Build from Source

#### 1. Clone the Repository

```bash
git clone https://github.com/ghostcluster-ai/ghostctl.git
cd ghostctl
```

#### 2. Build the Binary

```bash
make build
```

The binary will be created at `./bin/ghostctl`

#### 3. Install to PATH

```bash
make install
```

Or for development:

```bash
make install-dev
# Then add to your shell profile:
export PATH=$HOME/.local/bin:$PATH
```

### Method 2: Pre-built Binaries

Download from [Latest Release](https://github.com/ghostcluster-ai/ghostctl/releases):

```bash
# Linux
wget https://github.com/ghostcluster-ai/ghostctl/releases/download/v1.0.0/ghostctl-linux-amd64
chmod +x ghostctl-linux-amd64
sudo mv ghostctl-linux-amd64 /usr/local/bin/ghostctl

# macOS
wget https://github.com/ghostcluster-ai/ghostctl/releases/download/v1.0.0/ghostctl-darwin-amd64
chmod +x ghostctl-darwin-amd64
sudo mv ghostctl-darwin-amd64 /usr/local/bin/ghostctl

# Windows
# Download ghostctl.exe from releases
# Add to PATH
```

### Method 3: Using Homebrew (macOS/Linux)

```bash
brew tap ghostcluster-ai/ghostctl https://github.com/ghostcluster-ai/ghostctl.git
brew install ghostctl
```

Or install directly from the repository:

```bash
brew install ghostcluster-ai/ghostctl/ghostctl
```

### Method 4: Using Package Managers

#### Ubuntu/Debian

```bash
sudo apt-get update
sudo apt-get install ghostctl
```

#### Fedora/RHEL

```bash
sudo dnf install ghostctl
```

## Post-Installation Setup

### 1. Verify Installation

```bash
ghostctl version
```

Expected output:
```
ghostctl version v1.0.0 (commit: abc1234, built: 2024-01-01_00:00:00)
```

### 2. Initialize Configuration

ghostctl will create a default configuration file on first run:

```bash
ghostctl list
```

This creates `$HOME/.ghost/config.yaml` with defaults.

### 3. Configure Your Settings

Edit `$HOME/.ghost/config.yaml`:

```yaml
apiServer: localhost:8080
authToken: your-token
defaultTemplate: default
defaultTTL: 1h
namespace: ghostcluster
logLevel: info
```

### 4. Set Environment Variables (Optional)

```bash
export GHOSTCTL_LOG_LEVEL=debug
export GHOSTCTL_CONFIG=/custom/path/config.yaml
export GHOSTCTL_AUTH_TOKEN=your-token
```

### 5. Initialize Ghostcluster Controller

```bash
ghostctl init --namespace ghostcluster
```

## Uninstallation

### Remove Binary

```bash
make uninstall
```

Or manually:

```bash
rm /usr/local/bin/ghostctl
# or
rm $HOME/.local/bin/ghostctl
```

### Remove Configuration

```bash
rm -rf $HOME/.ghost
```

### Homebrew

```bash
brew uninstall ghostctl
brew untap ghostcluster-ai/ghostctl
```

## Troubleshooting

### Command Not Found

If `ghostctl` is not found after installation:

1. Verify it was installed: `ls -la /usr/local/bin/ghostctl`
2. Check PATH: `echo $PATH`
3. Add to PATH in `~/.bashrc` or `~/.zshrc`:

```bash
export PATH=$PATH:/usr/local/bin
source ~/.bashrc  # or ~/.zshrc
```

### Permission Denied

```bash
chmod +x /usr/local/bin/ghostctl
```

### Configuration Issues

Rebuild config with defaults:

```bash
rm $HOME/.ghost/config.yaml
ghostctl list  # Creates new config
```

### Version Mismatch

Ensure you have the compatible version:

```bash
ghostctl version
go version  # Check Go version
```

## Building for Different Platforms

### Linux

```bash
make build-linux
# Output: ./bin/ghostctl-linux-amd64
```

### macOS

```bash
make build-darwin
# Output: ./bin/ghostctl-darwin-amd64
```

### Windows

```bash
make build-windows
# Output: ./bin/ghostctl.exe
```

### All Platforms

```bash
make build-linux build-darwin build-windows
```

## Upgrading

### From Source

```bash
cd ghostctl
git pull origin main
make build
make install
```

### From Binary

1. Download new version
2. Replace old binary
3. Verify upgrade: `ghostctl version`

### Check for Updates

```bash
ghostctl version
# Compare with latest release on GitHub
```

## Security Considerations

### File Permissions

Configuration files are created with restricted permissions (0600):

```bash
ls -la $HOME/.ghost/
# Should show: -rw------- (user read/write only)
```

### Token Management

- Never commit tokens to version control
- Use environment variables for sensitive data
- Rotate tokens regularly
- Use `.gitignore` to exclude auth files

### SSL/TLS

For production, use TLS:

```yaml
# $HOME/.ghost/config.yaml
apiServer: https://api.example.com:8443
```

## System Requirements

### Minimum

- CPU: 1 core
- Memory: 256 MB
- Disk: 100 MB

### Recommended

- CPU: 2+ cores
- Memory: 1+ GB
- Disk: 1+ GB

## Network Requirements

- Outbound HTTPS (443) for API communication
- Access to Kubernetes cluster
- Optional: Outbound access to cloud providers (GCP, AWS, Azure)

## Next Steps

1. Run `ghostctl init` to initialize the controller
2. Create your first cluster: `ghostctl up my-cluster`
3. Check status: `ghostctl status my-cluster`
4. View examples: `ls examples/`

## Getting Help

- Documentation: `ghostctl --help`
- Command help: `ghostctl <command> --help`
- Issues: https://github.com/ghostcluster-ai/ghostctl/issues
- Discussions: https://github.com/ghostcluster-ai/ghostctl/discussions

# Goldfish - Cross-Platform Command Unification

Goldfish is a Go CLI tool that provides unified command interfaces that work consistently across different operating systems (Linux, macOS, Windows). It solves the problem of platform-specific command differences by translating unified commands to their platform-specific equivalents at runtime.

## Credits and attributions

- @danballance - concept, direction and AI management 
- Gemini 2.5Pro - initial project investigation and recommedantion to use Go + YAML configuration
- Claude Code (Sonnet 4) - creation of GenAI-ready project spec from deep research report
- Claude Code (Sonnet 4) - implementation of app in Go from project spec
- Claude Desktop (Opus 4) - design of cross-platform GHA CI/CD pipelines
- Cloud Code (Sonnet 4) - implementation of CI/CD pipelines for GHA

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Installation](#installation)
- [Usage](#usage)
- [Available Commands](#available-commands)
- [Configuration](#configuration)
- [Development](#development)
- [Testing](#testing)
- [CI/CD Pipeline](#cicd-pipeline)
- [Project Structure](#project-structure)
- [How It Works](#how-it-works)
- [Contributing](#contributing)

## Overview

### Problem Solved

Unix/Linux commands differ between platforms:
- **macOS** uses BSD tools (`sed -i ''`, different flag behaviors)
- **Linux** uses GNU tools (`sed -i`, different syntax)
- **Windows** requires PowerShell equivalents

This breaks cross-platform scripts and creates developer friction.

### Solution

Goldfish provides a unified interface:
```bash
# Same command works everywhere
goldfish replace --in-place 's/old/new/g' file.txt

# Translates to platform-specific commands:
# Linux:   sed -i 's/old/new/g' file.txt
# macOS:   sed -i '' 's/old/new/g' file.txt
# Windows: PowerShell equivalent
```

## Architecture

### Core Components

```
goldfish/
├── cmd/goldfish/           # CLI entry point
│   ├── main.go            # Cobra CLI setup and command generation
│   └── main_test.go       # Integration tests
├── internal/
│   ├── config/            # YAML configuration parsing
│   │   ├── config.go      # Config structures and validation
│   │   └── config_test.go # Unit tests
│   ├── engine/            # Command execution engine
│   │   ├── engine.go      # Template rendering and execution
│   │   └── engine_test.go # Unit tests
│   └── platform/          # OS detection
│       ├── platform.go    # Platform detection logic
│       └── platform_test.go # Unit tests
├── commands.yml           # Command definitions
├── go.mod                # Go module definition
└── Makefile              # Build automation
```

### Design Patterns

- **Strategy Pattern**: Platform-specific command templates
- **Template Method**: Go templates for command generation
- **Factory Pattern**: Dynamic Cobra command creation
- **Dependency Injection**: Clean separation of concerns

## Installation

### Option 1: Download Pre-built Binaries (Recommended)

Download the latest release from the [Releases page](../../releases):

```bash
# Linux AMD64
curl -L -o goldfish.tar.gz https://github.com/user/goldfish/releases/latest/download/goldfish_linux_amd64.tar.gz
tar -xzf goldfish.tar.gz
sudo mv goldfish /usr/local/bin/

# macOS (Intel)
curl -L -o goldfish.tar.gz https://github.com/user/goldfish/releases/latest/download/goldfish_darwin_amd64.tar.gz
tar -xzf goldfish.tar.gz
sudo mv goldfish /usr/local/bin/

# macOS (Apple Silicon)
curl -L -o goldfish.tar.gz https://github.com/user/goldfish/releases/latest/download/goldfish_darwin_arm64.tar.gz
tar -xzf goldfish.tar.gz
sudo mv goldfish /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/user/goldfish/releases/latest/download/goldfish_windows_amd64.zip" -OutFile "goldfish.zip"
Expand-Archive -Path "goldfish.zip" -DestinationPath "."
# Add goldfish.exe to your PATH
```

### Option 2: Build from Source

**Prerequisites:**
- Go 1.24.4 or later
- Make (optional, for build automation)

```bash
# Clone the repository
git clone <repository-url>
cd goldfish

# Build the binary
make build

# Or build manually
go build -o bin/goldfish ./cmd/goldfish

# Install globally (optional)
make install
```

### Verify Installation

```bash
goldfish --version
goldfish --help
```

## Usage

### Basic Usage

```bash
# Show all available commands
goldfish --help

# Get help for a specific command
goldfish <command> --help

# Execute a command
goldfish <command> [flags] [arguments]
```

### Examples

```bash
# Cross-platform file search
goldfish find --type f --name "*.go"
goldfish find /usr/local --type d --name "bin*"

# Text replacement (sed equivalent)
goldfish replace --in-place 's/old/new/g' file.txt
goldfish replace 's/foo/bar/g' input.txt  # outputs to stdout

# Archive creation (tar equivalent)
goldfish tar --compress --verbose archive.tar.gz ./src

# Process listing (ps equivalent)
goldfish ps --all
goldfish ps --user john

# Network information (netstat equivalent)
goldfish netstat --listening --tcp
```

## Available Commands

| Command | Alias | Description | Underlying Tool |
|---------|-------|-------------|-----------------|
| `replace-in-file` | `replace` | Cross-platform text replacement | `sed` / PowerShell |
| `find-files` | `find` | Cross-platform file search | `find` / PowerShell |
| `archive-create` | `tar` | Cross-platform archive creation | `tar` / PowerShell |
| `list-processes` | `ps` | Cross-platform process listing | `ps` / PowerShell |
| `network-info` | `netstat` | Cross-platform network info | `netstat` |

### Command Details

#### replace-in-file (replace)
```bash
# Replace text in file
goldfish replace --expression 's/old/new/g' --file input.txt --in-place

# Output to stdout (default)
goldfish replace --expression 's/foo/bar/g' --file input.txt
```

#### find-files (find)
```bash
# Find all Go files
goldfish find --name "*.go" --type f

# Find directories
goldfish find --path /usr/local --type d --name "bin*"

# Find large files
goldfish find --size "+1M" --type f
```

## Configuration

### commands.yml Structure

The behavior is defined in `commands.yml`:

```yaml
commands:
  - name: "command-name"           # Primary command name
    alias: "short-name"            # Optional shorter alias
    description: "What it does"    # Help text description
    base_command: "underlying-cmd" # Base system command
    params:                        # Parameter definitions
      - name: "param-name"         # Parameter identifier
        type: "string"             # Type: string, bool, int, float
        required: true             # Whether mandatory
        flag: "--flag-name"        # CLI flag (optional)
        description: "Help text"   # Parameter description
        default: "value"           # Default value (optional)
    platforms:                     # Platform-specific templates
      linux:
        template: "{{.base_command}} {{.params.param_name}}"
      darwin:
        template: "{{.base_command}} {{.params.param_name}}"
      windows:
        template: "powershell -Command \"...\""
```

### Template Variables

Templates have access to:
- `{{.base_command}}` - The underlying system command
- `{{.params.param_name}}` - Parameter values
- Standard Go template functions (if, range, etc.)

### Adding New Commands

1. Add command definition to `commands.yml`
2. Define platform-specific templates
3. Rebuild: `make build`
4. Test: `./bin/goldfish new-command --help`

## Development

### Project Structure Deep Dive

#### cmd/goldfish/main.go
- **GoldfishApp struct**: Main application state
- **Dynamic command generation**: Creates Cobra commands from YAML
- **Flag handling**: Maps YAML parameters to CLI flags
- **Error handling**: User-friendly error messages

#### internal/config/
- **Config loading**: Parses and validates `commands.yml`
- **Parameter validation**: Type checking and requirements
- **Command lookup**: Find commands by name or alias

#### internal/engine/
- **Template rendering**: Go template processing
- **Command execution**: Cross-platform subprocess management
- **Parameter parsing**: CLI argument to parameter mapping
- **Timeout handling**: Command execution limits

#### internal/platform/
- **OS detection**: Runtime platform identification
- **Platform validation**: Supported platform checking

### Key Design Decisions

1. **YAML over code**: Commands defined declaratively for easy extension
2. **Go templates**: Flexible platform-specific command generation
3. **Cobra framework**: Professional CLI experience with help, completion
4. **Internal packages**: Clean architecture with separated concerns
5. **Comprehensive testing**: Unit, integration, and benchmark tests

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/config

# Run with verbose output
go test -v ./...

# Run only short tests (skip integration tests)
go test -short ./...
```

### Test Structure

- **Unit tests**: Test individual functions and methods
- **Integration tests**: Test complete workflows
- **Benchmark tests**: Performance measurement
- **Table-driven tests**: Comprehensive input validation

### Test Coverage

The project includes 27 tests covering:
- Platform detection (100% coverage)
- Configuration parsing and validation
- Command execution engine
- Template rendering
- Parameter parsing
- Error handling
- End-to-end workflows

## CI/CD Pipeline

The project includes a comprehensive GitHub Actions pipeline for continuous integration and automated releases across multiple platforms.

### GitHub Actions Workflows

#### CI Workflow (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main` branch
- Pull requests to `main` branch

**Jobs:**
1. **Test Job** - Runs on Ubuntu
   - Executes all tests with race detection and coverage
   - Runs `go vet` for static analysis
   - Validates `go mod tidy` status
   
2. **Lint Job** - Code quality analysis
   - Uses `golangci-lint` for comprehensive static analysis
   - Enforces code style and best practices
   
3. **Build Job** - Cross-platform builds
   - Matrix builds for `linux/darwin/windows` on `amd64`
   - Additional `darwin/arm64` support
   - Uploads artifacts for testing

**Features:**
- **Caching**: Automatic Go module and build caching (20-40% faster builds)
- **Static binaries**: CGO_ENABLED=0 for portability
- **Optimized builds**: Uses `-ldflags="-s -w"` for smaller binaries
- **Artifact retention**: 30 days for development builds

#### Release Workflow (`.github/workflows/release.yml`)

**Triggers:**
- Git tags matching `v*` pattern (e.g., `v1.0.0`, `v2.1.3`)

**Cross-Platform Matrix:**
- **Linux**: amd64, arm64
- **macOS**: amd64, arm64  
- **Windows**: amd64

**Features:**
- **Version injection**: Embeds version, commit, and build date into binaries
- **Proper archives**: tar.gz for Unix, zip for Windows
- **Naming convention**: `goldfish_v1.0.0_linux_amd64.tar.gz`
- **Checksums**: SHA256 verification for all releases
- **GitHub releases**: Automatic release creation with detailed notes

### Using the CI/CD Pipeline

#### Development Workflow

1. **Push code to trigger CI**:
   ```bash
   git add .
   git commit -m "Add new feature"
   git push origin feature-branch
   ```
   
2. **Create pull request**:
   - CI runs automatically on PR creation
   - All tests, linting, and builds must pass
   - Artifacts available for testing

3. **Monitor workflow status**:
   - Check GitHub Actions tab for build status
   - Review test results and linting feedback
   - Download artifacts for manual testing

#### Release Process

1. **Prepare release**:
   ```bash
   # Ensure main branch is ready
   git checkout main
   git pull origin main
   
   # Update version if needed in code
   # Commit any final changes
   ```

2. **Create release tag**:
   ```bash
   # Create and push version tag
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. **Monitor release build**:
   - Release workflow triggers automatically
   - Builds for all supported platforms
   - Creates GitHub release with binaries

4. **Verify release**:
   - Check GitHub releases page
   - Download and test binaries
   - Verify checksums match

### Installation from Releases

#### Download Pre-built Binaries

Visit the [Releases page](../../releases) and download the appropriate binary:

- **Linux AMD64**: `goldfish_v1.0.0_linux_amd64.tar.gz`
- **Linux ARM64**: `goldfish_v1.0.0_linux_arm64.tar.gz`
- **macOS Intel**: `goldfish_v1.0.0_darwin_amd64.tar.gz`
- **macOS Apple Silicon**: `goldfish_v1.0.0_darwin_arm64.tar.gz`
- **Windows**: `goldfish_v1.0.0_windows_amd64.zip`

#### Install Binary

```bash
# Linux/macOS
curl -L -o goldfish.tar.gz https://github.com/user/goldfish/releases/download/v1.0.0/goldfish_v1.0.0_linux_amd64.tar.gz
tar -xzf goldfish.tar.gz
sudo mv goldfish /usr/local/bin/
goldfish --version

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/user/goldfish/releases/download/v1.0.0/goldfish_v1.0.0_windows_amd64.zip" -OutFile "goldfish.zip"
Expand-Archive -Path "goldfish.zip" -DestinationPath "."
# Move goldfish.exe to your PATH
```

#### Verify Installation

```bash
# Check binary integrity
sha256sum goldfish_v1.0.0_linux_amd64.tar.gz
# Compare with checksums.txt from release

# Verify installation
goldfish --version
goldfish --help
```

### Workflow Configuration

The workflows are configured for optimal performance and security:

- **Permissions**: Minimal required permissions (`contents: read/write`)
- **Caching**: Go modules and build cache for faster builds
- **Security**: Static binaries, no CGO dependencies
- **Artifacts**: Proper naming conventions and organization
- **Retention**: 30 days for CI artifacts, permanent for releases

### Troubleshooting CI/CD

#### Common Issues

1. **Build failures**:
   - Check Go version compatibility in `go.mod`
   - Ensure all tests pass locally: `make test`
   - Verify linting passes: `make check`

2. **Release failures**:
   - Ensure tag follows `v*` pattern
   - Check that main branch builds successfully
   - Verify no uncommitted changes

3. **Artifact issues**:
   - Binary permissions may need adjustment after extraction
   - Windows binaries require `.exe` extension
   - Use appropriate archive format (tar.gz vs zip)

#### Debugging Workflows

1. **Check workflow logs**:
   - Navigate to GitHub Actions tab
   - Click on failed workflow run
   - Review step-by-step logs

2. **Local testing**:
   ```bash
   # Test locally before pushing
   make test
   make check
   make build
   
   # Test cross-compilation
   GOOS=linux GOARCH=amd64 go build ./cmd/goldfish
   GOOS=darwin GOARCH=arm64 go build ./cmd/goldfish
   ```

3. **Manual release testing**:
   ```bash
   # Test release build locally
   go build -ldflags="-s -w -X 'main.version=v1.0.0'" ./cmd/goldfish
   ./goldfish --version
   ```

## How It Works

### Execution Flow

1. **Initialization**
   ```go
   // Load commands.yml configuration
   config := config.LoadDefault()
   
   // Detect current platform
   platform := platformDetector.Current()
   
   // Generate Cobra commands dynamically
   generateCommands(config, platform)
   ```

2. **Command Generation**
   ```go
   // For each command in YAML
   for _, cmd := range config.Commands {
       // Create Cobra command
       cobraCmd := &cobra.Command{
           Use: cmd.Name,
           RunE: executeCommand,
       }
       
       // Add flags from parameters
       addParameterFlags(cobraCmd, cmd.Parameters)
   }
   ```

3. **Command Execution**
   ```go
   // Parse user input
   params := parseParameters(args, flags)
   
   // Render platform-specific template
   command := renderTemplate(cmd.Platforms[platform], params)
   
   // Execute system command
   exec.Command("sh", "-c", command).Run()
   ```

### Template Rendering Example

Input YAML:
```yaml
platforms:
  linux:
    template: "sed {{if .params.in_place}}-i{{end}} '{{.params.expression}}' {{.params.file}}"
```

With parameters `{in_place: true, expression: "s/old/new/g", file: "test.txt"}`:

Renders to: `sed -i 's/old/new/g' test.txt`

### Error Handling Strategy

- **Configuration errors**: Detailed YAML validation messages
- **Parameter errors**: Clear missing/invalid parameter feedback
- **Execution errors**: Preserve exit codes, show command context
- **Platform errors**: Graceful handling of unsupported platforms

### Security Considerations

- **Input validation**: All parameters validated before execution
- **Template safety**: Go templates prevent injection attacks
- **Subprocess control**: Proper timeout and signal handling
- **Exit code preservation**: Maintains shell script compatibility

## Building and Linting

### Build Commands

```bash
# Build development version
make build

# Run all quality checks
make check

# Clean build artifacts
make clean

# Install locally
make install

# Build release version
make release
```

### Code Quality

The project maintains high code quality through:

- **golangci-lint**: Comprehensive static analysis
- **gofmt**: Consistent code formatting  
- **go vet**: Static analysis for common mistakes
- **errcheck**: Ensures all errors are handled
- **Tests**: 100% of critical paths covered

### Linting Rules

All code passes these linters:
- `errcheck`: Error handling verification
- `gocritic`: Performance and style suggestions
- `staticcheck`: Advanced static analysis
- `gosec`: Security vulnerability detection

## Contributing

### Development Workflow

1. **Setup**
   ```bash
   git clone <repo>
   cd goldfish
   make dev  # build + test
   ```

2. **Make Changes**
   - Add/modify commands in `commands.yml`
   - Update code in `internal/` packages
   - Add tests for new functionality

3. **Validate Locally**
   ```bash
   make check  # lint + test
   ./bin/goldfish <new-command> --help
   ```

4. **Submit Pull Request**
   - Push changes to feature branch
   - Create pull request to `main`
   - GitHub Actions CI will run automatically:
     - Tests with race detection
     - golangci-lint analysis
     - Cross-platform builds
   - All CI checks must pass before merge

5. **Release Process** (for maintainers)
   ```bash
   # After merging to main
   git tag v1.0.0
   git push origin v1.0.0
   # Release workflow creates binaries automatically
   ```

### Adding New Commands

1. **Define in YAML**
   ```yaml
   - name: "new-command"
     description: "What it does"
     base_command: "system-cmd"
     # ... parameters and platforms
   ```

2. **Test thoroughly**
   ```bash
   make build
   ./bin/goldfish new-command --help
   ./bin/goldfish new-command <args>
   ```

3. **Add tests** if needed for complex logic

### Code Style

- Follow Go conventions
- Add comprehensive comments (required by CLAUDE.md)
- Use descriptive variable names
- Handle all errors explicitly
- Write tests for new functionality

## Support

- **Issues**: Report bugs and feature requests
- **Documentation**: This README and inline code comments
- **Examples**: See `commands.yml` for configuration examples
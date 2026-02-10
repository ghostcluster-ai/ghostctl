# Contributing to ghostctl

Thank you for your interest in contributing to ghostctl! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful and professional in all interactions with other contributors and maintainers.

## Getting Started

### 1. Fork and Clone

```bash
git clone https://github.com/your-username/ghostctl.git
cd ghostctl
```

### 2. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number
```

### 3. Set Up Development Environment

```bash
make deps
make build
```

## Development Guidelines

### Code Style

- Follow idiomatic Go conventions
- Use `gofmt` and `golangci-lint`
- Run `make fmt` before committing

### Testing

- Write tests for new features
- Ensure all tests pass: `make test`
- Maintain >80% code coverage where possible

### Commit Messages

- Use clear, descriptive commit messages
- Reference issues when applicable: `Fixes #123`
- Follow conventional commits format when possible

Examples:
```
feat: add cluster template filtering
fix: handle GPU allocation errors
docs: update installation instructions
```

### Pull Request Process

1. Ensure all tests pass: `make test`
2. Ensure code is formatted: `make fmt`
3. Ensure linting passes: `make lint`
4. Update documentation if needed
5. Create a descriptive PR with context

## Adding New Commands

### 1. Create Command File

```bash
# Create cmd/mycommand.go
touch cmd/mycommand.go
```

### 2. Implement Command

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Short description",
    Long:  "Longer description with examples",
    RunE:  runMyCmd,
}

func init() {
    myCmd.Flags().StringVar(&myVar, "flag", "default", "Help text")
    RootCmd.AddCommand(myCmd)
}

func runMyCmd(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

### 3. Register Command

Add to `cmd/root.go`:

```go
RootCmd.AddCommand(myCmd)
```

## Adding New Packages

### Keep It Modular

- Place business logic in `internal/` packages
- Place utilities in `pkg/` packages
- Use clear package names that describe functionality

### Package Structure

```go
package mypackage

// Exported types and functions
type MyType struct {
    // Fields
}

func NewMyType() *MyType {
    return &MyType{}
}

// Internal helper functions
func internalHelper() {
    // Implementation
}
```

## Documentation

- Update README.md for user-facing changes
- Add inline code comments for complex logic
- Update help text in commands
- Include examples in command long descriptions

## Reporting Issues

When reporting bugs:

1. Check if the issue already exists
2. Include Go version: `go version`
3. Include OS and architecture
4. Provide minimal reproducible example
5. Include error messages and logs

## Build and Release Process

### Building

```bash
make build              # Build for current platform
make build-linux        # Build for Linux
make build-darwin       # Build for macOS
make build-windows      # Build for Windows
```

### Local Installation

```bash
make install-dev
# Add to PATH: export PATH=$HOME/.local/bin:$PATH
```

## Testing

### Run All Tests

```bash
make test
```

### Run Specific Test

```bash
go test -v -run TestNamePattern ./...
```

### Coverage Report

```bash
make test-coverage
```

## Performance Considerations

- Minimize API calls
- Cache when appropriate
- Use goroutines for concurrent operations
- Log at appropriate levels

## Security Considerations

- Never commit credentials or tokens
- Use environment variables for sensitive data
- Validate user input
- Sanitize output
- Keep dependencies updated

## Asking for Help

- Check existing issues and discussions
- Read the documentation carefully
- Ask in discussions or open an issue
- Provide context and examples

## Recognition

Contributors will be recognized in:
- Commit history
- Release notes
- GitHub contributors page

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to ghostctl! 🚀

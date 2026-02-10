package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CommandResult represents the result of a shell command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// ExecuteCommand executes a shell command and captures output
func ExecuteCommand(command string, args ...string) (*CommandResult, error) {
	cmd := exec.Command(command, args...)
	
	// Capture output
	output, err := cmd.CombinedOutput()
	
	result := &CommandResult{
		Stdout:   string(output),
		ExitCode: cmd.ProcessState.ExitCode(),
	}
	
	if err != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	
	return result, nil
}

// ExecuteCommandWithEnv executes a command with custom environment variables
func ExecuteCommandWithEnv(env []string, command string, args ...string) (*CommandResult, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = env
	
	output, err := cmd.CombinedOutput()
	
	result := &CommandResult{
		Stdout:   string(output),
		ExitCode: cmd.ProcessState.ExitCode(),
	}
	
	if err != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	
	return result, nil
}

// ExecuteCommandStreaming executes a command with real-time output streaming
func ExecuteCommandStreaming(command string, args ...string) (int, error) {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	err := cmd.Run()
	
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return 1, err
		}
	}
	
	return exitCode, nil
}

// ExecuteCommandStreamingWithEnv executes a command with environment and real-time output
func ExecuteCommandStreamingWithEnv(env []string, command string, args ...string) (int, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	err := cmd.Run()
	
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return 1, err
		}
	}
	
	return exitCode, nil
}

// CommandExists checks if a command is available in PATH
func CommandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// ParseCommand parses a command string into program and arguments
func ParseCommand(cmdStr string) (string, []string, error) {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("empty command")
	}
	return parts[0], parts[1:], nil
}

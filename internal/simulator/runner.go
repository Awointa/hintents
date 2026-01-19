package simulator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Runner handles the execution of the Rust simulator binary
type Runner struct {
	BinaryPath string
}

// NewRunner creates a new simulator runner.
// It checks for the binary in common locations.
func NewRunner() (*Runner, error) {
	// 1. Check environment variable
	if envPath := os.Getenv("ERST_SIMULATOR_PATH"); envPath != "" {
		return &Runner{BinaryPath: envPath}, nil
	}

	// 2. Check current directory (for Docker/Production)
	cwd, err := os.Getwd()
	if err == nil {
		localPath := filepath.Join(cwd, "erst-sim")
		if _, err := os.Stat(localPath); err == nil {
			return &Runner{BinaryPath: localPath}, nil
		}
	}

	// 3. Check development path (assuming running from sdk root)
	devPath := filepath.Join("simulator", "target", "release", "erst-sim")
	if _, err := os.Stat(devPath); err == nil {
		return &Runner{BinaryPath: devPath}, nil
	}

	// 4. Check global PATH
	if path, err := exec.LookPath("erst-sim"); err == nil {
		return &Runner{BinaryPath: path}, nil
	}

	return nil, fmt.Errorf("simulator binary 'erst-sim' not found. Please build it or set ERST_SIMULATOR_PATH")
}

// Run executes the simulation with the given request
func (r *Runner) Run(req *SimulationRequest) (*SimulationResponse, error) {
	// Serialize Request
	inputBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Prepare Command
	cmd := exec.Command(r.BinaryPath)
	cmd.Stdin = bytes.NewReader(inputBytes)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("simulator execution failed: %w, stderr: %s", err, stderr.String())
	}

	// Deserialize Response
	var resp SimulationResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w, output: %s", err, stdout.String())
	}

	// Check logic error from simulator
	if resp.Status == "error" {
		return nil, fmt.Errorf("simulation error: %s", resp.Error)
	}

	return &resp, nil
}

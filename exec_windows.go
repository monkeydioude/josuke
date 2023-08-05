package josuke

import (
	"bytes"
	"fmt"
	"os/exec"
)

// NativeExecuteCommand executes a command.
func NativeExecuteCommand(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err, cmd.Stderr)
	}
	return nil
}

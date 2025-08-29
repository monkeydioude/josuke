package josuke

import (
	"os/exec"
)

// NativeExecuteCommand executes a command with a specific user.
func NativeExecuteCommand(cmd *exec.Cmd) error {
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

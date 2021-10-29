package josuke

import (
	"os/exec"
)

// NativeExecuteCommand executes a command.
func NativeExecuteCommand(cmd *exec.Cmd) error {
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

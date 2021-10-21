package josuke

import (
	"os/exec"
)

func NativeExecuteCommand(cmd *exec.Cmd) error {
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

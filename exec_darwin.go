package josuke

import (
	"os/exec"
	"syscall"
)

// NativeExecuteCommand executes a command with a specific user.
func NativeExecuteCommand(cmd *exec.Cmd) error {
	// Own process group, so stopping the command also stops the processes it started.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

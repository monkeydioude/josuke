package josuke

import (
	"os/exec"
	"syscall"
)

// NativeExecuteCommand executes a command with a specific user.
func NativeExecuteCommand(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: currentUser.Uid,
			Gid: currentUser.Gid,
			NoSetGroups: true,
		},
		// Own process group, so stopping the command also stops the processes it started.
		Setpgid: true,
	}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

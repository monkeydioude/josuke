package josuke

import (
	"os/exec"
	"syscall"
)

func NativeExecuteCommand(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: currentUser.Uid, Gid: currentUser.Gid}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

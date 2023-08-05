package josuke

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"syscall"
)

// NativeExecuteCommand executes a command with a specific user.
func NativeExecuteCommand(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: currentUser.Uid, Gid: currentUser.Gid},
	}
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err, cmd.Stderr)
	}
	log.Printf("[INFO] %s\n", cmd.Stdout)
	return nil
}

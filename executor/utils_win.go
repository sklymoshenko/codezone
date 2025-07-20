//go:build windows

package executor

import (
	"context"
	"os/exec"
	"strings"
	"syscall"
)

func ExecCommandContext(ctx context.Context, command []string, input string, tempFile, tempDir string) (string, string, error) {
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	cmd.Dir = tempDir

	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

func ExecCommand(command []string) (string, error) {
	cmd := exec.Command(command[0], command[1:]...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	output, err := cmd.Output()

	return string(output), err
}

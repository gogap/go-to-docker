package builder

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

func execCommandToShow(cwd string, cmdStr string) (err error) {

	parts := strings.Fields(cmdStr)
	name := parts[0]
	args := parts[1:len(parts)]

	cmd := exec.Command(name, args...)

	if len(cwd) > 0 {
		cmd.Dir = cwd
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	if err = cmd.Start(); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	return
}

func execCommand(cwd string, cmdStr string) (out []byte, err error) {

	parts := strings.Fields(cmdStr)
	name := parts[0]
	args := parts[1:len(parts)]

	cmd := exec.Command(name, args...)

	if len(cwd) > 0 {
		cmd.Dir = cwd
	}

	return cmd.CombinedOutput()
}

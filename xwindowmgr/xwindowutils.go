package xwindowmgr

import (
	"fmt"
	"os"
	"os/exec"
)

func Spawn(command string, args ...string) {
	go func() {
		cmd := exec.Command(command, args...)

		cmd.Env = os.Environ()

		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		if err := cmd.Start(); err != nil {
			fmt.Println("Error at run command:", command, "; error:", err)
		}
	}()
}

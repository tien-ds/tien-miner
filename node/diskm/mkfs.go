package diskm

import (
	"os"
	"os/exec"
)

//mkfs.ext4 /dev/sdb

func MkfsExt4(dev string) error {
	command := exec.Command("mkfs.ext4", dev)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

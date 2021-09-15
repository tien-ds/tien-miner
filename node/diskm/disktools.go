package diskm

import (
	"github.com/shirou/gopsutil/disk"
	"os"
	"os/exec"
)

func MkFSExt4(dev string) error {
	command := exec.Command("mkfs.ext4", dev)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func E2Label(dev string, label string) error {
	command := exec.Command("e2label", dev, label)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func GetMountpointWithDisk(dev string) string {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return ""
	}
	for _, partition := range partitions {
		if partition.Device == dev {
			return partition.Mountpoint
		}
	}
	return ""
}

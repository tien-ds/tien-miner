package diskm

import (
	"github.com/shirou/gopsutil/disk"
	"strings"
)

const DS_DIR_PREFIX = "/media/root/ds"

// HasDSMounts list all blocks
func HasDSMounts() []disk.PartitionStat {
	devices, err := ListBlockDevices()
	if err != nil {
		return nil
	}
	var origin []disk.PartitionStat
	for _, device := range devices {
		if strings.HasPrefix(device.MountPoint, DS_DIR_PREFIX) {
			origin = append(origin, disk.PartitionStat{
				Device:     "/dev/" + device.DeviceName,
				Mountpoint: device.MountPoint,
				Fstype:     device.FilesystemType,
				Opts:       "",
			})
		}
	}
	return origin
}

// OnMountDisk {"/dev/sdb":"/root/ds"}
var onMountDisk = make(map[string]string)

func UsedMinerDisk(dev string, mountPoint string) {
	if _, b := onMountDisk[dev]; !b {
		onMountDisk[dev] = mountPoint
	}
}

func DiskSize() int {
	return len(onMountDisk)
}

func HasMinedDev(dev string) bool {
	if _, b := onMountDisk[dev]; b {
		return true
	}
	return false
}

func UnMountPlugin(dev string) {
	delete(onMountDisk, dev)
}

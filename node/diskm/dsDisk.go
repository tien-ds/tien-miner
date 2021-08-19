package diskm

import "github.com/shirou/gopsutil/disk"

const DS_DIR = "/media/root/ds"

func HasDSMount() (has bool, devName string, point string) {
	devices, err := disk.Partitions(true)
	if err != nil {
		return false, "", ""
	}
	for _, device := range devices {
		if device.Mountpoint == DS_DIR {
			return true, device.Device, device.Mountpoint
		}
	}
	return false, "", ""
}

var dev = ""

func SetDevName(name string) {
	dev = name
}

package muldisk

import (
	"github.com/shirou/gopsutil/disk"
	"github.com/sirupsen/logrus"
	"math/big"
	"strings"
)

// InPart where path included in Block dev
// eg /media/root/ds-5/.depaas/node
func InPart(path string) (bool, *disk.PartitionStat) {
	var temp disk.PartitionStat

	partitions, err := disk.Partitions(true)
	if err != nil {
		return false, nil
	}

	for _, partition := range partitions {

		// eg. /              /media/root/ds   yes
		// eg  /media/root/ds /media/root/ds-1 yes
		//fmt.Println(partition)
		if partition.Fstype == "ext4" || partition.Fstype == "ntfs" || partition.Fstype == "fuseblk" {
			prefix := strings.HasPrefix(path, partition.Mountpoint+"/")
			//fmt.Println(path, partition.Mountpoint+"/", prefix)
			if prefix && partition.Mountpoint != "/" {
				if len(partition.Mountpoint) > len(temp.Mountpoint) {
					temp = partition
				}
			}
		}
	}

	if temp.Mountpoint == "" {
		return false, nil
	}
	return true, &temp
}

func GetTotalDisk() *big.Int {
	newInt := big.NewInt(0)
	for path := range dbs {
		if b, dPart := InPart(path); b {
			usage, err := disk.Usage(dPart.Mountpoint)
			if err != nil {
				return nil
			}
			newInt.Add(newInt, big.NewInt(int64(usage.Total)))
		}
	}
	return newInt
}

func UnPlugin(dev string) {
	logrus.Debug("UnPlugin", dev)
	for key := range dbs {
		isBlocked, partDev := InPart(key)
		if isBlocked && partDev.Device == dev {
			delete(dbs, key)
			delete(mountPoints, partDev.Mountpoint)
			logrus.Infof("Unplugin %s", key)
		}

	}
}

package muldisk

import (
	"github.com/shirou/gopsutil/disk"
	"math/big"
	"strings"
)

var partitions []disk.PartitionStat

func init() {
	var err error
	partitions, err = disk.Partitions(true)
	if err != nil {
		return
	}
}

// InPart where path included in Block dev
func InPart(path string) (bool, *disk.PartitionStat) {
	var temp disk.PartitionStat
	for _, partition := range partitions {

		// eg. /              /media/root/ds   yes
		// eg  /media/root/ds /media/root/ds-1 yes

		if partition.Fstype == "ext4" || partition.Fstype == "ntfs" {
			prefix := strings.HasPrefix(path, partition.Mountpoint+"/")
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

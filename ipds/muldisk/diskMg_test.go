package muldisk

import (
	"fmt"
	"github.com/shirou/gopsutil/disk"
	"math/rand"
	"path"
	"testing"
)

func TestDiskUsage(t *testing.T) {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return
	}
	for _, partition := range partitions {
		fmt.Println(partition.String())
	}
}

func TestDir(t *testing.T) {
	fmt.Println(path.Base("/mnt/aa/bb"))
}

func TestRand(t *testing.T) {
	for i := 0; i < 100; i++ {
		fmt.Println(rand.Intn(3))
	}
}

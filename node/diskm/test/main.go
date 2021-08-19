package main

import (
	"fmt"
	"github.com/kr/pretty"
	"github.com/shirou/gopsutil/disk"
)

func main() {
	devices, err := disk.Partitions(true)
	if err != nil {
		return
	}
	for _, device := range devices {
		fmt.Println(pretty.Sprint(device))
	}
}

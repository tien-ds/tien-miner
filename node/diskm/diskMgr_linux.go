package diskm

import (
	"fmt"
	"github.com/ds/depaas/closer"
	"github.com/ds/depaas/ipds/muldisk"
	"github.com/dustin/go-humanize"
	"github.com/pilebones/go-udev/netlink"
	"github.com/shirou/gopsutil/disk"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"path"
	"strings"
	"syscall"
	"time"
)

func StartDiskManger() {
	//var rules netlink.RuleDefinitions
	logrus.Info("start StartDiskManger")
	go monitor(nil)
}

// fixDevBlock fix err block or unmount mountpoint
func fixDevBlock() {
	blocksDev, _ := disk.Partitions(true)
	fond := func(dev string) bool {
		for _, s := range blocksDev {
			if s.Device == dev {
				return true
			}
		}
		return false
	}
	has := HasDSMounts()
	for _, parStat := range has {
		if !fond(parStat.Device) {
			if err := syscall.Unmount(parStat.Mountpoint, syscall.MNT_FORCE); err != nil {
				logrus.Errorf("fixBlock Unmount %s", err)
			} else {
				logrus.Infof("fix %s", parStat.Mountpoint)
			}
		}
	}
}

func WarpPath(paths []disk.PartitionStat) string {
	if len(paths) == 1 {
		return paths[0].Mountpoint
	} else {
		var p string
		for _, path := range paths {
			p += path.Mountpoint + ";"
		}
		return p[:len(p)-1]
	}
}

// CheckDiskReady gcSize
func CheckDiskReady(ready func(gcSize uint64)) {

	//check disk is mounted and umount unused dev
	fixDevBlock()

	for {
		mounts := HasDSMounts()
		if len(mounts) != 0 {
			warpPath := WarpPath(mounts)
			os.Setenv("DS_PATH", warpPath)
			logrus.Infof("check %s has mounted", warpPath)
			for _, mount := range mounts {
				UsedMinerDisk(mount.Device, mount.Mountpoint)
			}
			if ready != nil {
				//usage, err := disk.Usage(point)
				//if err != nil {
				//	ready(0)
				//}
				//ready(usage.Total)
				bytes, err := humanize.ParseBytes("10G")
				if err != nil {
					return
				}
				//FIXME multi disk ???
				ready(bytes)
			}
			return
		}
		time.Sleep(time.Second * 2)
		logrus.Info("wait for mount disk")
	}
}

// monitor run monitor mode
func monitor(matcher netlink.Matcher) {

	conn := new(netlink.UEventConn)
	if err := conn.Connect(netlink.UdevEvent); err != nil {
		log.Fatalln("Unable to connect to Netlink Kobject UEvent socket")
	}
	defer conn.Close()
	logrus.Info("Monitoring UEvent kernel message to user-space...")
	queue := make(chan netlink.UEvent)
	errors := make(chan error)
	quit := conn.Monitor(queue, errors, matcher)

	closer.RegisterCloser("udev", closer.NewSimpleCloser(func() error {
		close(quit)
		return nil
	}))

	// Handling message from queue
	for {
		select {
		case uevent := <-queue:
			diskInfo(uevent)
		case err := <-errors:
			log.Println("ERROR:", err)
		}
	}
}

func diskInfo(e netlink.UEvent) {
	if typ, b := e.Env["ID_TYPE"]; b && typ == "disk" {
		devType := e.Env["DEVTYPE"]
		fsType := e.Env["ID_FS_TYPE"]
		devName := e.Env["DEVNAME"]
		//add disk /dev/sdb ext4
		logrus.Infof("%s %s %s %s", e.Action.String(), devType, devName, fsType)
		if e.Action.String() == "add" {
			fmt.Println("add dev ", humanize.Bytes(GetBlockSize(devName)))
			fmt.Println(GetLabel(devName))
			if strings.HasPrefix(GetLabel(devName), "ds") {
				go WaitAppend(devName)
			}
		}
		if e.Action.String() == "remove" && HasMinedDev(devName) {
			UnMountPlugin(devName)
			muldisk.UnPlugin(devName)
			diskSize := DiskSize()
			logrus.Infof("remain disksize %d", diskSize)
			if diskSize == 0 {
				closer.CloseWithName("config.db")
				closer.CloseWithName("udev")
				os.Exit(0)
			}
			//pm2 will restart self
		}
	}
}

func WaitAppend(dev string) {
	for i := 0; i < 5; i++ {
		withDiskPath := GetMountpointWithDisk(dev)
		if withDiskPath != "" {
			pathDB := path.Join(withDiskPath, ".depaas", os.Getenv("repo"))
			muldisk.AppendDataStore(pathDB)
			UsedMinerDisk(dev, withDiskPath)
			logrus.Infof("Append DB %s", pathDB)
			break
		}
		logrus.Debugf("wait for %s", dev)
		time.Sleep(time.Second * 30)
	}
}

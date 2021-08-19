package diskm

import (
	"github.com/ds/depaas/closer"
	"github.com/pilebones/go-udev/netlink"
	"github.com/shirou/gopsutil/disk"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"syscall"
	"time"
)

func StartDiskManger() {
	//var rules netlink.RuleDefinitions
	logrus.Info("start StartDiskManger")
	go monitor(nil)
}

func fixDev() {
	point := GetDevWithMountPoint(DS_DIR)
	has, _, _ := HasDSMount()
	if point == "" && has {
		if err := syscall.Unmount(DS_DIR, syscall.MNT_FORCE); err != nil {
			logrus.Errorf("fixBlock Unmount %s", err)
		} else {
			logrus.Infof("fix %s", DS_DIR)
		}
	}
}

func CheckDiskReady(ready func(uint642 uint64)) {
	fixDev()
	for {
		has, name, point := HasDSMount()
		if has {
			os.Setenv("DS_PATH", DS_DIR)
			logrus.Infof("check %s has mounted", DS_DIR)
			SetDevName(name)
			if ready != nil {
				usage, err := disk.Usage(point)
				if err != nil {
					ready(0)
				}
				ready(usage.Total)
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
		logrus.Infof("%s %s %s %s", e.Action.String(), devType, devName, fsType)
		if e.Action.String() == "remove" && devName == dev {
			os.Exit(0)
		}
	}
}

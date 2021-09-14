package diskm

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var pairsRE = regexp.MustCompile(`([A-Z:]+)=(?:"(.*?)")`)

const (
	// values for the TYPE column that we care about

	typeDisk = "disk"
	typeLoop = "loop"
	typePart = "part"

	// listBlockDevicesPeriod is the time period between block device listings.
	// Unfortunately Linux's inotify does not work with virtual filesystems, so
	// polling it is.
	listBlockDevicesPeriod = time.Second * 30

	// bytesInMiB is the number of bytes in a MiB.
	bytesInMiB = 1024 * 1024
)

// BlockDevice describes a block device discovered on a machine.
type BlockDevice struct {
	// DeviceName is the block device's OS-specific name (e.g. "sdb").
	DeviceName string `yaml:"devicename,omitempty"`

	// DeviceLinks is a collection of symlinks to the block device
	// that the OS maintains (e.g. "/dev/disk/by-id/..."). Storage
	// provisioners can match volume attachments to device links if
	// they know ahead of time how the OS will name them.
	DeviceLinks []string `yaml:"devicelinks,omitempty"`

	// Label is the label for the filesystem on the block device.
	//
	// This will be empty if the block device does not have a filesystem,
	// or if the filesystem is not yet known to Juju.
	Label string `yaml:"label,omitempty"`

	// UUID is a unique identifier for the filesystem on the block device.
	//
	// This will be empty if the block device does not have a filesystem,
	// or if the filesystem is not yet known to Juju.
	//
	// The UUID format is not necessarily uniform; for example, LVM UUIDs
	// differ in format to the standard v4 UUIDs.
	UUID string `yaml:"uuid,omitempty"`

	// HardwareId is the block device's hardware ID, which is composed of
	// a serial number, vendor and model name. Not all block devices have
	// these properties, so HardwareId may be empty. This is used to identify
	// a block device if it is available, in preference to UUID or device
	// name, as the hardware ID is immutable.
	HardwareId string `yaml:"hardwareid,omitempty"`

	// WWN is the block device's World Wide Name (WWN) unique identifier.
	// Not all block devices have one, so WWN may be empty. This is used
	// to identify a block device if it is available, in preference to
	// UUID or device name, as the WWN is immutable.
	WWN string `yaml:"wwn,omitempty"`

	// BusAddress is the bus address: where the block device is attached
	// to the machine. This is currently only populated for disks attached
	// to the SCSI bus.
	//
	// The format for this is <bus>@<bus-specific-address> as according to
	// "lshw -businfo". For example, for a SCSI disk with Host=1, Bus=2,
	// Target=3, Lun=4, we populate this field with "scsi@1:2.3.4".
	BusAddress string `yaml:"busaddress,omitempty"`

	// Size is the size of the block device, in MiB.
	Size uint64 `yaml:"size"`

	// FilesystemType is the type of the filesystem present on the block
	// device, if any.
	FilesystemType string `yaml:"fstype,omitempty"`

	// InUse indicates that the block device is in use (e.g. mounted).
	InUse bool `yaml:"inuse"`

	// MountPoint is the path at which the block devices is mounted.
	MountPoint string `yaml:"mountpoint,omitempty"`

	// SerialId is the block devices serial id used for matching.
	SerialId string `yaml:"serialid,omitempty"`
}

func GetBlockDevs() []string {
	devices, err := ListBlockDevices()
	if err != nil {
		logrus.Debug(devices)
		return nil
	}

	var re []string
	for _, device := range devices {
		if strings.HasPrefix(device.MountPoint, DS_DIR_PREFIX) {
			re = append(re, device.DeviceName)
		}
	}
	return re
}

func GetBlockSize(dev string) uint64 {
	devices, err := ListBlockDevices()
	if err != nil {
		return 0
	}
	for _, device := range devices {
		if dev == ("/dev/" + device.DeviceName) {
			return device.Size
		}
	}
	return 0
}

func ListBlockDevices() ([]BlockDevice, error) {
	columns := []string{
		"KNAME",      // kernel name
		"SIZE",       // size
		"LABEL",      // filesystem label
		"UUID",       // filesystem UUID
		"FSTYPE",     // filesystem type
		"TYPE",       // device type
		"MOUNTPOINT", // moint point
		"MAJ:MIN",    // major/minor device numbers
	}

	logrus.Tracef("executing lsblk")
	output, err := exec.Command(
		"lsblk",
		"-b", // output size in bytes
		"-P", // output fields as key=value pairs
		"-o", strings.Join(columns, ","),
	).Output()
	if err != nil {
		return nil, errors.New(
			"cannot list block devices: lsblk failed",
		)
	}

	var devices []BlockDevice
	s := bufio.NewScanner(bytes.NewReader(output))
	for s.Scan() {
		pairs := pairsRE.FindAllStringSubmatch(s.Text(), -1)
		var dev BlockDevice
		var deviceType string
		var majorMinor string
		for _, pair := range pairs {
			switch pair[1] {
			case "KNAME":
				dev.DeviceName = pair[2]
			case "SIZE":
				size, err := strconv.ParseUint(pair[2], 10, 64)
				if err != nil {
					logrus.Errorf(
						"invalid size %q from lsblk: %v", pair[2], err,
					)
				} else {
					dev.Size = size
				}
			case "LABEL":
				dev.Label = pair[2]
			case "UUID":
				dev.UUID = pair[2]
			case "FSTYPE":
				dev.FilesystemType = pair[2]
			case "TYPE":
				deviceType = pair[2]
			case "MOUNTPOINT":
				dev.MountPoint = pair[2]
			case "MAJ:MIN":
				majorMinor = pair[2]
			default:
				logrus.Debugf("unexpected field from lsblk: %q", pair[1])
			}
		}

		// We may later want to expand this, e.g. to handle lvm,
		// dmraid, crypt, etc., but this is enough to cover bases
		// for now.
		switch deviceType {
		case typeLoop:
		case typePart:
		case typeDisk:
			// Floppy disks, which have major device number 2,
			// should be ignored.
			if strings.HasPrefix(majorMinor, "2:") {
				logrus.Tracef("ignoring flopping disk device: %+v", dev)
				continue
			}
		default:
			logrus.Tracef("ignoring %q type device: %+v", deviceType, dev)
			continue
		}

		// Check if the block device is in use. We need to know this so we can
		// issue an error if the user attempts to allocate an in-use disk to a
		// unit.
		dev.InUse, err = blockDeviceInUse(dev)
		if os.IsNotExist(err) {
			// In LXC containers, lsblk will show the block devices of the
			// host, but the devices will typically not be present.
			continue
		} else if err != nil {
			logrus.Debugf("could not check if %q is in use: %v", dev.DeviceName, err)
			// We cannot detect, so err on the side of caution and default to
			// "in use" so the device cannot be used.
			dev.InUse = true
		}

		// Add additional information from sysfs.
		if err := addHardwareInfo(&dev); err != nil {
			logrus.Errorf(
				"error getting hardware info for %q from sysfs: %v",
				dev.DeviceName, err,
			)
		}
		devices = append(devices, dev)
	}
	if err := s.Err(); err != nil {
		return nil, errors.New("cannot parse lsblk output")
	}
	return devices, nil
}

// blockDeviceInUse checks if the specified block device
// is in use by attempting to open the device exclusively.
//
// If the error returned satisfies os.IsNotExists, then
// the device will be ignored altogether.
var blockDeviceInUse = func(dev BlockDevice) (bool, error) {
	f, err := os.OpenFile("/dev/"+dev.DeviceName, os.O_EXCL, 0)
	if err == nil {
		f.Close()
		return false, nil
	}
	perr, ok := err.(*os.PathError)
	if !ok {
		return false, err
	}
	// open(2): "In general, the behavior of O_EXCL is undefined if
	// it is used without O_CREAT. There is one exception: on Linux
	// 2.6 and later, O_EXCL can be used without O_CREAT if pathname
	// refers to a block device. If the block device is in use by the
	// system  (e.g., mounted), open() fails with the error EBUSY."
	if errno, _ := perr.Err.(syscall.Errno); errno == syscall.EBUSY {
		return true, nil
	}
	return false, err
}

// addHardwareInfo adds additional information about the hardware, and how it is
// attached to the machine, to the given BlockDevice.
func addHardwareInfo(dev *BlockDevice) error {
	logrus.Tracef(`executing "udevadm info" for %s`, dev.DeviceName)
	output, err := exec.Command(
		"udevadm", "info",
		"-q", "property",
		"--name", dev.DeviceName,
	).CombinedOutput()
	if err != nil {
		msg := "udevadm failed"
		if output := bytes.TrimSpace(output); len(output) > 0 {
			msg += fmt.Sprintf(" (%s)", output)
		}
		return errors.New(msg)
	}

	var devpath, idBus, idSerial, wwnWithExtension string

	s := bufio.NewScanner(bytes.NewReader(output))
	for s.Scan() {
		line := s.Text()
		sep := strings.IndexRune(line, '=')
		if sep == -1 {
			logrus.Debugf("unexpected udevadm output line: %q", line)
			continue
		}
		key, value := line[:sep], line[sep+1:]
		switch key {
		case "DEVPATH":
			devpath = value
		case "DEVLINKS":
			dev.DeviceLinks = strings.Split(value, " ")
		case "ID_BUS":
			idBus = value
		case "ID_SERIAL":
			idSerial = value
		case "ID_WWN":
			dev.WWN = value
		case "ID_WWN_WITH_EXTENSION":
			wwnWithExtension = value
		default:
			logrus.Tracef("ignoring line: %q", line)
		}
	}
	if err := s.Err(); err != nil {
		return errors.New("cannot parse udevadm output")
	}

	// For cases where there are logical disks attached to a
	// controller (eg RAID), the controller itself has a WWN
	// which is the ID_WWN value and each disk has a WWN with
	// a vendor extension ID_WWN_WITH_EXTENSION added to
	// identify the disk via the /dev/disk/by-id path.
	if wwnWithExtension != "" {
		dev.WWN = wwnWithExtension
	}
	if idBus != "" && idSerial != "" {
		// ID_BUS will be something like "scsi" or "ata";
		// ID_SERIAL will be something like ${MODEL}_${SERIALNO};
		// and together they make up the symlink in /dev/disk/by-id.
		dev.HardwareId = idBus + "-" + idSerial
	}
	if idSerial != "" {
		dev.SerialId = idSerial
	}

	// For devices on the SCSI bus, we include the address. This is to
	// support storage providers where the SCSI address may be specified,
	// but the device name can not (and may change, depending on timing).
	if idBus == "scsi" && devpath != "" {
		// DEVPATH will be "<uninteresting stuff>/<SCSI address>/block/<device>".
		re := regexp.MustCompile(fmt.Sprintf(
			`^.*/(\d+):(\d+):(\d+):(\d+)/block/(?:\w+/|)%s$`,
			regexp.QuoteMeta(dev.DeviceName),
		))
		submatch := re.FindStringSubmatch(devpath)
		if submatch != nil {
			// We use the address scheme used by lshw: bus@address. We don't use
			// lshw because it does things we don't need, and that slows it down.
			//
			// In DEVPATH, the address format is "H:C:T:L" ([H]ost, [C]hannel,
			// [T]arget, [L]un); the lshw address format is "H:C.T.L"
			dev.BusAddress = fmt.Sprintf(
				"scsi@%s:%s.%s.%s",
				submatch[1], submatch[2], submatch[3], submatch[4],
			)
		} else {
			logrus.Debugf(
				"non matching DEVPATH for %q: %q",
				dev.DeviceName, devpath,
			)
		}
	}

	return nil
}

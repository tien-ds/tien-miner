package utils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenRand() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%d", r.Uint64())
}

func GetContextDir(name string) string {
	repoPath := path.Join(BestPoolPath(), name)
	if !Exist(repoPath) {
		os.MkdirAll(repoPath, 0777)
	}
	return repoPath
}

func BestPoolPath() string {
	dir := os.Getenv("DS_PATH")
	if dir == "" {
		dir, _ = os.UserHomeDir()
	}
	return path.Join(dir, ".depaas")
}

func GBToBytes(gb float64) int64 {
	return int64(1024 * 1024 * 1024 * gb)
}

func IsPortOpen(addr string) bool {
	c, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return false
	} else {
		defer c.Close()
		return true
	}
}

func GetMacWifi() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, inter := range interfaces {
		mac := strings.Replace(inter.HardwareAddr.String(), ":", "", -1) //获取本机MAC地址
		if inter.Name == "eth0" || inter.Name == "enp4s0" {
			logrus.Debug("GetMacWifi mac = ", mac)
			return mac
		}
	}
	return ""
}

func Shutdown(f func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

	// Block until a signal is received.
	<-c
	f()
	os.Exit(0)
}

func RandArray(times, len int) []int {
	var rs []int
	for i := 0; i < times; i++ {
		rs = append(rs, rand.Intn(len))
	}
	return rs
}

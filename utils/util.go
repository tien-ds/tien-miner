package utils

import (
	"fmt"
	"github.com/ds/depaas/logger"
	"github.com/ds/depaas/node/env"
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

func GetConfigDir() string {
	dir, _ := os.UserHomeDir()
	repoPath := path.Join(path.Join(dir, ".depaas"), "config")
	if !Exist(repoPath) {
		os.MkdirAll(repoPath, 0777)
	}
	return repoPath
}

func GetContextDir(name string) string {
	poolPath := BestPoolPath()
	if strings.Contains(poolPath, ";") {
		var cDir string
		for _, dir := range strings.Split(poolPath, ";") {
			cDir += path.Join(dir, name) + ";"
		}
		return cDir[:len(cDir)-1]
	}
	repoPath := path.Join(poolPath, name)
	return repoPath
}

func BestPoolPath() string {
	dir := os.Getenv("DS_PATH")
	if dir == "" {
		dir, _ = os.UserHomeDir()
	}
	if strings.Contains(dir, ";") {
		var cDir string
		for _, d := range strings.Split(dir, ";") {
			cDir += path.Join(d, ".depaas") + ";"
		}
		return cDir[:len(cDir)-1]
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

func SetLog(isGw bool) {

	log := env.GetEnv("LOG")
	switch log {
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	default:
		logrus.SetLevel(logrus.TraceLevel)
	}

	if env.GetEnv("TRACEFILE") != "" {
		logrus.SetReportCaller(true)
	}

	logger.InjectLogrus(logger.File("/var/log/depaas/out.log"), logger.File("/var/log/depaas/err.log"))
}

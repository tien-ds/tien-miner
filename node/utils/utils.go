package utils

import (
	"bytes"
	"github.com/ds/depaas/protocol"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func HttpPostClient(url string, query []byte) ([]byte, error) {
	client := new(http.Client)
	reader := bytes.NewReader(query)

	logrus.Debugf("httpPostClient url %s query %s", url, string(query))
	request, err := http.NewRequest(http.MethodPost, url, reader)
	defer request.Body.Close()

	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	resp, err := client.Do(request)
	if err != nil || resp.StatusCode != http.StatusOK {
		logrus.Errorf("httpPostClient err %s", err)
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	logrus.Debugf("httpPostClient respBytes %s", string(respBytes))
	return respBytes, err
}

// GetDsType 2家庭矿机  1云矿机 3超级矿机
func GetDsType() protocol.MinerType {
	dsType, err := strconv.Atoi(os.Getenv("DS_TYPE"))
	if err != nil {
		return protocol.MINER_CLOUD
	}
	return protocol.MinerType(dsType)
}

type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

// DiskUsage disk usage of path/disk
func DiskUsage(path string) (disk DiskStatus) {
	//fs := syscall.Statfs_t{}
	//err := syscall.Statfs(path, &fs)
	//if err != nil {
	//	return
	//}
	//disk.All = fs.Blocks * uint64(fs.Bsize)
	//disk.Free = fs.Bfree * uint64(fs.Bsize)
	//disk.Used = disk.All - disk.Free
	return
}

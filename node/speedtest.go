package node

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func DownloadSpeed() (float64, error) {
	sysInfo := make(map[string]interface{})
	sysInfo["len"] = true

	data, _ := json.Marshal(sysInfo)
	resp, time, err := httpPost("http://192.168.6.35:3000/users/test", data)
	respBytes, err := ioutil.ReadAll(resp.Body)

	fsize := int64(len(respBytes))
	speed := float64((fsize * int64(1000000000)) / int64(time))
	logrus.Debug("DownloadSpeed fsize:", fsize, ",time:", time, ",speed:", speed)
	//logrus.Debug(string(respBytes))
	return speed, err
}

func DownloadSpeedString() (float64, error) {
	sysInfo := make(map[string]interface{})
	sysInfo["len"] = "true"

	data, _ := json.Marshal(sysInfo)
	resp, time, err := httpPost("http://192.168.6.35:3000/users/test", data)
	respBytes, err := ioutil.ReadAll(resp.Body)

	fsize := int64(len(respBytes))
	speed := float64((fsize * int64(1000000000)) / int64(time))
	logrus.Debug("DownloadSpeed fsize:", fsize, ",time:", time, ",speed:", speed)
	//logrus.Debug(string(respBytes))
	return speed, err
}

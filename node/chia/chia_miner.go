package chia

import (
	"encoding/json"
	"fmt"
	"github.com/ds/depaas/protocol"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func WalkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		if err != nil { //忽略错误
			return err
		}

		if fi.IsDir() { // 忽略目录
			return nil
		}

		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}

		return nil
	})

	return files, err
}

func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0)

	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}

		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}

	return files, nil
}

type PlotStruct struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ChiaMinerInfo struct {
	protocol.MsgType
	Plots  []PlotStruct `json:"plots"`
	PeerID string       `json:"peerId"`
}

func ListPlots(dirPth string) (plots []PlotStruct, err error) {
	plots = make([]PlotStruct, 0)
	files, err := ListDir(dirPth, "plot")
	if err != nil {
		return plots, err
	}

	for _, value := range files {
		var plot PlotStruct
		file, err := os.Open(value)
		if err == nil {
			fi, _ := file.Stat()
			plot.Size = fi.Size()
			plot.Name = path.Base(value)
			plots = append(plots, plot)
			logrus.Debug("ListPlots plot.Size:", plot.Size, ",plot.Name:", plot.Name)
		}
	}

	logrus.Debug("ListPlots plots:", plots)
	return plots, nil
}

func ChiaInfoMsg(id string, peerId string) []byte {
	plots, _ := ListPlots("/media/root/ds/plots")

	logrus.Debug("ChiaInfoMsg plots:", plots)
	var info ChiaMinerInfo
	info.Plots = plots
	info.ID = id
	info.PeerID = peerId
	info.Type = 30

	for i := 0; i <= 15; i++ {
		p, _ := ListPlots(fmt.Sprintf("/media/root/chia%d/plots", i))
		info.Plots = append(info.Plots, p...)
	}

	logrus.Debug("ChiaInfoMsg info:", info)
	data, _ := json.Marshal(info)
	logrus.Debug("ChiaInfoMsg data", string(data))
	return data
}

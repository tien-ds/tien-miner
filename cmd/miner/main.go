package main

import (
	"context"
	"flag"
	"gitee.com/fast_api/api"
	"github.com/ds/depaas/closer"
	"github.com/ds/depaas/database/config"
	"github.com/ds/depaas/ipds"
	"github.com/ds/depaas/ipds/web"
	"github.com/ds/depaas/node"
	"github.com/ds/depaas/node/diskm"
	"github.com/ds/depaas/node/utils"
	"github.com/ds/depaas/protocol"
	gutils "github.com/ds/depaas/utils"
	"github.com/sirupsen/logrus"
	"os"
)

var (
	bindAddr = flag.String("bind", "", "bind ds addr")
	data     = flag.String("data", "", "ds store dir")
)

func init() {
	os.Setenv("repo", "node")
}

func main() {
	doArgs()

	gutils.SetLog(false)

	//hook shutdown
	go gutils.Shutdown(func() {
		closer.CloseAll()
	})

	ipds.InitNodeConfig()

	//start init ipfs
	if utils.GetDsType() == protocol.MINER_HOME {
		diskm.StartDiskManger()
		go diskm.CheckDiskReady(func(uint642 uint64) {
			InitIpds()
			//set repo Max size
			ipds.SetMaxSize(uint642)
		})
	} else {
		InitIpds()
	}

	go api.StartService("127.0.0.1:8888")

	err := node.StartMinerWithNode(context.Background())
	if err != nil {
		logrus.Error(err)
		os.Exit(0)
	}
}

func InitIpds() {
	ipds.SetNode(ipds.InitNode())
	//start file server
	web.StartFileServer()
	//start gc
	ipds.StartGC(ipds.GetNode())
	//check blocks
	go node.StartCheckBlock(context.Background())
}

func doArgs() {
	flag.Parse()
	if *bindAddr == "" {
		config.SetBindAddr(*bindAddr)
	}

	//set DS_PATH
	if *data != "" {
		os.Setenv("DS_PATH", *data)
	}
}

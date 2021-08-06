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
	"github.com/ds/depaas/node/utils"
	gutils "github.com/ds/depaas/utils"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

var (
	bindAddr = flag.String("bind", "", "bind ds addr")
	data     = flag.String("data", "", "ds store dir")
	size     = flag.String("size", "", "set max size (default 10GB)")
)

func init() {
	os.Setenv("repo", "node")
}

func ReSize() {
	if *size != "" {
		c, err := ipds.GetNode().Repo.Config()
		if err != nil {
			return
		}
		logrus.Infof("resize repo %s", *size)
		c.Datastore.StorageMax = *size
		err = ipds.GetNode().Repo.SetConfig(c)
		if err != nil {
			logrus.Error(err)
			os.Exit(0)
		}
	}
}

func main() {
	doArgs()

	utils.SetLog()

	var g sync.WaitGroup
	g.Add(1)
	//start init ipfs
	ipfsNode, _ := ipds.IPfsInit()
	ipds.SetNode(ipfsNode)
	//start miner
	err := node.StartMinerWithNode(context.Background(), ipds.GetNode())
	if err != nil {
		logrus.Error(err)
		os.Exit(0)
	}

	//start file server
	web.StartFileServer()

	//resize
	ReSize()

	//start gc
	ipds.StartGC(ipds.GetNode())
	//check blocks
	go node.StartCheckBlock(context.Background())
	//hook shutdown
	go gutils.Shutdown(func() {
		closer.Close()
	})

	go api.StartService("127.0.0.1:8888")

	g.Wait()
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

package ipds

import (
	"context"
	"gitee.com/fast_api/api"
	"github.com/ds/depaas/closer"
	"github.com/ds/depaas/utils"
	"github.com/dustin/go-humanize"
	"github.com/ipfs/go-ipfs/core/corerepo"
	"github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo"
	"github.com/libp2p/go-libp2p-core/peer"
	"os"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	dsLog "github.com/ipfs/go-log/v2"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/sirupsen/logrus"
)

var dirMap = make(map[string]string)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

// OpenOrCreateRepo repoPath only create or open db
func OpenOrCreateRepo(repoPath string, configDir string) repo.Repo {
	plugins, err := loader.NewPluginLoader(repoPath)
	checkError(err)
	err = plugins.Initialize()
	checkError(err)
	err = plugins.Inject()
	checkError(err)
	repo, err := OpenRepo(repoPath, configDir)
	checkError(err)
	return repo
}

func InitNodeConfig() {
	dsLog.SetLogLevel("corerepo", "debug")
	if !IsInitialized() {
		config, err := GenConfig(os.Stdout, 2048)
		err = DSInitConfig(config)
		checkError(err)
	}
}

func InitNode() *core.IpfsNode {
	repoPath := utils.GetContextDir(os.Getenv("repo"))
	repo := OpenOrCreateRepo(repoPath, utils.GetConfigDir())
	//register closer
	checkError(closer.RegisterCloser("ipds", repo))
	ctx := context.Background()
	ipfsNode, err := core.NewNode(ctx, &core.BuildCfg{
		Online:    true,
		Repo:      repo,
		Permanent: true,
		Routing:   libp2p.DHTOption, //libp2p.DHTServerOption
		//Host:   libp2p.DefaultHostOption,
		ExtraOpts: map[string]bool{
			"pubsub": true,
		},
	})
	checkError(err)
	return ipfsNode
}

func IPfsInit() (*core.IpfsNode, iface.CoreAPI) {
	InitNodeConfig()
	ipfsNode := InitNode()
	//start gc
	StartGC(ipfsNode)
	SetNode(ipfsNode)

	cApi, err := coreapi.NewCoreAPI(ipfsNode)
	checkError(err)
	api.GET(func() interface{} {
		aa, _ := cApi.Swarm().Peers(context.Background())
		var res []peer.AddrInfo
		for _, info := range aa {
			addrInfo, err := cApi.Dht().FindPeer(context.Background(), info.ID())
			if err != nil {
				continue
			}
			res = append(res, addrInfo)
		}
		return res
	}, "/api/swarm/peer")

	api.GET(func() interface{} {
		key, err := cApi.Key().Self(context.Background())
		if err != nil {
			panic(err)
		}
		addLists, _ := cApi.Swarm().LocalAddrs(context.Background())
		return map[string]interface{}{
			"id":    key.ID().String(),
			"addrs": addLists,
		}
	}, "/api/id")

	api.GET(func() interface{} {
		u, _ := ipfsNode.Repo.GetStorageUsage()
		return map[string]uint64{
			"diskUse": u,
		}
	}, "/api/diskUsed")

	return ipfsNode, cApi

}

func StartGC(node *core.IpfsNode) {
	//GC
	go func() {
		cfg, err := node.Repo.Config()
		logrus.Infof("start GC OK max = %s", cfg.Datastore.StorageMax)
		err = corerepo.PeriodicGC(context.Background(), node)
		checkError(err)
	}()

}

func SetMaxSize(uint642 uint64) {
	c, err := GetNode().Repo.Config()
	if err != nil {
		return
	}
	logrus.Infof("resize repo %s", humanize.Bytes(uint642))
	c.Datastore.StorageMax = humanize.Bytes(uint642)
	err = GetNode().Repo.SetConfig(c)
	if err != nil {
		logrus.Error(err)
		os.Exit(0)
	}
}

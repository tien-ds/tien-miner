package ipds

import (
	"context"
	"gitee.com/fast_api/api"
	"github.com/ds/depaas/closer"
	"github.com/ds/depaas/utils"
	"github.com/ipfs/go-ipfs/core/corerepo"
	"github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"os"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/plugin/loader"
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

func InitNode() *core.IpfsNode {

	dsLog.SetLogLevel("corerepo", "debug")

	ctx := context.Background()

	repoPath := utils.GetContextDir(os.Getenv("repo"))

	plugins, err := loader.NewPluginLoader(repoPath)
	checkError(err)

	err = plugins.Initialize()
	checkError(err)

	err = plugins.Inject()
	checkError(err)

	if !IsInitialized() {
		config, err := Init(os.Stdout, 2048)
		err = DSInit(repoPath, config)
		checkError(err)
	}

	repo, err := OpenRepo(repoPath)
	checkError(err)

	//register closer
	checkError(closer.RegisterCloser("ipds", repo))

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

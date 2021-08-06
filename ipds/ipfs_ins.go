package ipds

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/commands/keyencode"
	"github.com/ipfs/go-ipfs/core/coreapi"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/sirupsen/logrus"
)

var mNode *core.IpfsNode

func GetPeerID() string {
	//mNode.Identity
	encoder, err := keyencode.KeyEncoderFromString("b58mh")
	if err != nil {
		return ""
	}
	return encoder.FormatID(mNode.Identity)
}

func SetNode(node *core.IpfsNode) {
	mNode = node
}

func GetNode() *core.IpfsNode {
	return mNode
}

func GetApi() coreiface.CoreAPI {
	api, err := coreapi.NewCoreAPI(GetNode())
	if err != nil {
		logrus.Error(err)
		return nil
	}
	return api
}

type BlockUnit struct {
	Block cid.Cid
	Size  uint64
}

func LsCid(sCid cid.Cid) []BlockUnit {
	var blocks []BlockUnit
	d, _ := GetApi().Unixfs().Ls(context.Background(), path.IpfsPath(sCid))
	for {
		kk := <-d
		if kk.Size == 0 {
			break
		} else {
			blocks = append(blocks, BlockUnit{
				Block: kk.Cid,
				Size:  kk.Size,
			})
		}
	}
	return blocks
}

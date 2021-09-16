package service

import (
	"context"
	"encoding/json"
	"github.com/ds/depaas/database/dispatchdb"
	"github.com/ds/depaas/ipds"
	"github.com/ds/depaas/pools"
	"github.com/ds/depaas/protocol"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"math/rand"
	"sort"

	"github.com/ipfs/go-cid"
	"github.com/sirupsen/logrus"
)

const SaveTimes = 3

// ReceiveCid
// 矿机类型 2家庭矿机  1云矿机 3超级矿机
//
///**
func ReceiveCid(cid cid.Cid) {
	logrus.Debug(cid)
	blocks := ipds.LsCid(cid)
	online2 := pools.Online().TypeKeys(protocol.MINER_HOME)  //家庭矿机
	online1 := pools.Online().TypeKeys(protocol.MINER_CLOUD) //云矿机
	//file is small
	sort.Strings(online2)
	if blocks == nil {
		logrus.Infof("file is small origin to %d machine ", SaveTimes)
		mi := GenRand(SaveTimes, len(online2))
		var ser3 []string
		for _, index := range mi {
			ser3 = append(ser3, online2[index])
		}
		DispatchBlock(cid, ser3)
		return
	}
	//家庭矿机4份
	SaveTimesType2 := 4
	if len(online2) > SaveTimesType2 { //online must > SaveTimesType2
		for _, block := range blocks {
			mi := GenRand(SaveTimesType2, len(online2))
			var ser3 []string
			for _, index := range mi {
				ser3 = append(ser3, online2[index])
			}
			logrus.Infof("block %s in %s ", block.Block.String(), ser3)
			DispatchBlock(block.Block, ser3)
		}
	} else {
		DispatchBlock(cid, online2)
	}
	//云矿机2份
	SaveTimesType1 := 2
	if len(online1) > SaveTimesType1 { //online must > SaveTimesType1
		for _, block := range blocks {
			mi := GenRand(SaveTimesType1, len(online1))
			var ser3 []string
			for _, index := range mi {
				ser3 = append(ser3, online1[index])
			}
			logrus.Infof("block %s in %s ", block.Block.String(), ser3)
			DispatchBlock(block.Block, ser3)
		}
	}
}

// DispatchBlock dispatch block must store CID in gw database
// because miner everyone may invoke contract along.
func DispatchBlock(block cid.Cid, peers []string) {
	size := GetCidSize(block)
	for _, peer := range peers {
		NewPinServer(peer).PinAdd(block.String(), func(id string, result interface{}) {
			bytes, _ := json.Marshal(result)
			dispatchdb.Store(block.String(), peer, size)
			logrus.Debugf("PinAdd %s %s", id, string(bytes))
		})
	}
}

func GetCidSize(cid cid.Cid) uint64 {
	background := context.Background()
	resolvePath, err := ipds.GetApi().ResolvePath(background, path.IpfsPath(cid))
	entries, err := ipds.GetApi().Unixfs().Ls(background, resolvePath, options.Unixfs.ResolveChildren(true))
	if err != nil {
		return 0
	}
	var t uint64
	for entry := range entries {
		if entry.Err != nil {
			break
		}
		t += entry.Size
	}
	logrus.Infof("cid %s size %d", cid.String(), t)
	return t
}

func CleanCid(cid cid.Cid, peers []string) {
	for _, peer := range peers {
		NewPinServer(peer).PinRm(cid.String(), func(id string, result interface{}) {
			bytes, _ := json.Marshal(result)
			logrus.Debugf("PinRemove %s %s", id, string(bytes))
		})
	}
}

func GenRand(times, len int) []int {
	var rs []int
	for i := 0; i < times; i++ {
		rs = append(rs, rand.Intn(len))
	}
	return rs
}

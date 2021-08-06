package service

import (
	"context"
	"github.com/ds/depaas/database"
	"github.com/ds/depaas/database/config"
	"github.com/ds/depaas/ipds"
	"github.com/ds/depaas/ipds/service"
	"github.com/ds/depaas/pools"
	"github.com/ds/depaas/utils"
	"github.com/ipfs/go-cid"
	ipath "github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

func CheckBlocks(sCid string, peerId string) {
	decode, err := cid.Decode(sCid)
	if err != nil {
		logrus.Error(err)
		return
	}
	db := path.Join(utils.GetTempDir(), sCid)
	logrus.Tracef("db temp %s", db)
	err = service.GetFile(decode, db)
	if err != nil {
		logrus.Error(err)
		return
	}
	cacheDB, err := database.OpenDBFile(db)
	if err != nil {
		logrus.Error(err)
		return
	}

	defer func() {
		cacheDB.Close()
		os.Remove(db)
	}()

	config.NewMBlockWith(cacheDB).ListAllBlock(func(scid string) {
		count := pools.Online().Count()
		if count > 6 {
			cid, err := cid.Decode(sCid)
			if err != nil {
				return
			}
			sPeer := ListBlock(cid)
			has := len(sPeer)
			if has >= 8 {
				times := has - 6
				ints := utils.RandArray(times, has)
				for _, v := range ints {
					NewPinServer(sPeer[v].String()).PinRm(cid.String(), func(id string, result interface{}) {
						logrus.Tracef("block rm %s", scid)
					})
				}

			}
			if len(sPeer) < 6 {
				times := has - 6
				randArray := utils.RandArray(times, count)
				for _, v := range randArray {
					peer := pools.Online().Keys()[v]
					NewPinServer(peer).PinAdd(scid, func(id string, result interface{}) {
						logrus.Tracef("block patch %s", scid)
					})
				}
			}
		} else {
			logrus.Warnf("connect peers too little %d", count)
		}

	})
}

func ListBlock(block cid.Cid) []peer.AddrInfo {
	aa, _ := ipds.GetApi().Dht().FindProviders(context.Background(), ipath.New(block.String()))
	var blocksPeers []peer.AddrInfo
	for {
		tt := <-aa
		if tt.ID == "" {
			break
		} else {
			blocksPeers = append(blocksPeers, tt)
		}
	}
	return blocksPeers
}

package rest

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gitee.com/fast_api/api"
	"gitee.com/fast_api/api/def"
	"github.com/ds/depaas/ipds"
	ip2region "github.com/ds/depaas/ipregion"
	"github.com/ds/depaas/persistence"
	"github.com/ds/depaas/pools"
	"github.com/ds/depaas/protocol"
	"github.com/ds/depaas/service"
	"github.com/ds/depaas/utils"
	"github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	"mime/multipart"
	"reflect"
	"strings"
	"time"
)

var checkSize = options.Unixfs.Chunker("size-1048576")

func InitRest() {
	//app rest
	apiRest()

	test()
}

func apiRest() {
	api.GET(getOnlineDev, "/api/pool/online")
	api.GET(GetOnlineCdnPeers, "/api/pool/cdnPeers")
	api.GET(getMinerInfo, "/api/pool/peer")
	api.POST(uploadFile1, "/api/pool/upload")
	api.POST(uploadFile, "/api/pool/uploadFile")
	api.POST(uploadFile2, "/api/pool/upload2")
	//api.POST(uploadFile1, "/api/pool/upload")
	//api.GetApi().POST(uploadDir, "/api/pool/uploadDir")
	//api.GetApi().POST(uploadData, "/api/pool/speed")
	api.GET(GetBinds, "/api/pool/getBinds")
	api.GET(func(peerId def.StringReq, offset, number int) []interface{} {
		return persistence.GetStoreDB().FindByPage(protocol.InfoType{PeerId: peerId.String()}, offset, number)
	}, "/api/pool/getWorkHistory")

	api.GET(func(peerId string, begin, end int64) []interface{} {
		return persistence.GetStoreDB().FindByIndexValueFilter(protocol.InfoType{PeerId: peerId}, func(indexValue interface{}) bool {
			aa := indexValue.(*protocol.InfoType)
			if aa.Time > begin && aa.Time < end {
				return true
			} else {
				return false
			}
		})
	}, "/api/pool/getWorkHistoryWithTime")
	//api.GetApi().POST(server.SendMessage, "/api/pool/send")

	api.GET(func(sCid def.StringReq) interface{} {
		dht := ipds.GetApi().Dht()
		ci, err := cid.Decode(sCid.String())
		if err != nil {
			panic(err)
		}
		blocks := ipds.LsCid(ci)
		var kvs = make(map[string][]peer.AddrInfo)
		for _, block := range blocks {
			aa, _ := dht.FindProviders(context.Background(), path.New(block.Block.String()))
			var blocksPeers []peer.AddrInfo
			for {
				tt := <-aa
				if tt.ID == "" {
					break
				} else {
					blocksPeers = append(blocksPeers, tt)
				}
			}
			kvs[block.Block.String()] = blocksPeers
		}
		return kvs
	}, "/api/pool/listBlock")

	api.GET(func(peer def.StringReq) interface{} {
		var kks []string
		persistence.GetOrigDB().Range(func(key []byte, value []byte) bool {
			k := string(key)
			if strings.Contains(k, peer.String()) {
				kks = append(kks, k)
			}
			return false
		})
		return kks
	}, "/api/pool/all")

	api.GET(func(peerId def.StringReq) interface{} {
		return persistence.GetStoreDB().Get(peerId.String(), reflect.TypeOf(protocol.BlockchainInfo{}))
	}, "/api/pool/chia")

	api.GET(func(peerId def.StringReq) interface{} {
		return persistence.GetStoreDB().FindAll(protocol.ActionHistory{
			PeerId: peerId.String(),
		})
	}, "/api/pool/OffLineRate")

	api.POST(func(a struct {
		Cmd  string `json:"cmd" req:"true"`
		Peer string `json:"peer" req:"true"`
	}) interface{} {
		var v interface{}
		con, cancel := context.WithTimeout(context.Background(), time.Second*5)
		err := pools.Online().Get(a.Peer, func(msg protocol.Msg) {
			rand := utils.GenRand()
			msg.SendEncryptMsg(protocol.CmdMsg(rand, a.Cmd)).
				Subscriber(func(id string, f interface{}) {
					res := f.(*protocol.CmderResult)
					v = res.Text
					cancel()
				})
		})
		if err != nil {
			panic(err)
		}
		<-con.Done()
		if con.Err() != context.Canceled {
			panic("time out")
		}
		return v
	}, "/api/test")

	api.GET(func(peerId def.StringReq) interface{} {
		rand := utils.GenRand()
		aaa := time.Now()
		con, cancel := context.WithTimeout(context.Background(), time.Second*15)
		var v interface{}
		err := pools.Online().Get(string(peerId), func(msg protocol.Msg) {
			msg.SendEncryptMsg(protocol.CmdSeed(rand)).Subscriber(func(id string, f interface{}) {
				res := f.(*protocol.Mnemonic)
				v = res.Mnemonic
				cancel()
			})
		})
		if err != nil {
			panic(err)
		}
		<-con.Done()
		fmt.Println(time.Since(aaa))
		if con.Err() != context.Canceled {
			panic(con.Err())
		}
		return map[string]interface{}{
			"seed": v,
		}

	}, "/api/pool/seed")

	api.POST(func(code struct {
		ArmUrl  string `json:"armUrl"`
		AmdUrl  string `json:"amdUrl"`
		Version int    `json:"version"`
	}) interface{} {
		keys := pools.Online().Keys()
		for _, v := range keys {
			pools.Online().Get(v, func(msg protocol.Msg) {
				randID := utils.GenRand()
				msg.SendEncryptMsg(protocol.CmdUpdate(randID, code.ArmUrl, code.AmdUrl, code.Version)).
					Subscriber(func(id string, f interface{}) {
						fmt.Println(f)
					})

			})
		}
		return code
	}, "/api/pool/update")

	api.GET(func(name def.StringReq) interface{} {
		return persistence.GetStoreDB().Get(string(name), reflect.TypeOf(protocol.Wifi{}))
	}, "/api/pool/wifi")

	api.POST(func(a struct {
		Addr   string `json:"addr"`
		PeerId string `json:"peerId"`
	}) interface{} {
		err := pools.Online().Get(a.PeerId, func(msg protocol.Msg) {
			msg.SendEncryptMsg(protocol.GETChiaMinerInfo{MsgType: protocol.MsgType{
				Type: protocol.CHIA_INFO_CODE,
				ID:   "7894531",
			}})
		})
		if err != nil {
			logrus.Error(err)
		}
		ms := getBindsWithAddr(a.Addr)
		ms = append(ms, a.PeerId)
		saveBinds(a.Addr, ms)
		return map[string]bool{
			"state": true,
		}
	}, "/api/pool/bind")

	api.GET(func(a struct {
		Addr   string `json:"addr"`
		PeerId string `json:"peerId"`
	}) interface{} {
		ms := getBindsWithAddr(a.Addr)
		for i, v := range ms {
			if v == a.PeerId {
				ms = append(ms[:i], ms[i:]...)
				break
			}
		}
		saveBinds(a.Addr, ms)
		return map[string]bool{
			"state": true,
		}
	}, "/api/pool/unbind")

	api.GET(func(head def.Header, sCid def.StringReq) interface{} {
		sign := head.Get("sign")
		fmt.Println(sign)
		cid, err := cid.Parse(sCid.String())
		if err != nil {
			panic(err)
		}
		go service.ReceiveCid(cid)
		return map[string]bool{
			"state": true,
		}
	}, "/api/pool/dispatcher")

	api.POST(func(param struct {
		PeerId    string `json:"peerId"`
		SefAmount string `json:"sefAmount"`
		SefTo     string `json:"sefTo"`
		FeeAmount string `json:"feeAmount"`
		FeeTo     string `json:"feeTo"`
	}) interface{} {
		//TODO ethurl
		con, cancel := context.WithTimeout(context.Background(), time.Second*30)
		id := utils.GenRand()
		res := make(map[string]string)
		err := pools.Online().Get(param.PeerId, func(msg protocol.Msg) {
			msg.SendEncryptMsg(protocol.BeeCmd(id, protocol.BEE_TRANSFER, "", []protocol.BeeParamArgs{
				{
					Amount: param.SefAmount,
					To:     param.SefTo,
				}, {
					Amount: param.FeeAmount,
					To:     param.FeeTo,
				},
			})).Subscriber(func(id string, f interface{}) {
				if v, b := f.(*protocol.CmderResult); b {
					res["result"] = v.Text
					if v.Type == protocol.CMD_RESP_SUCCESS {
						res["state"] = "success"
					} else {
						res["state"] = "fail"
					}
				}
				cancel()
			})
		})
		if err != nil {
			panic(err)
		}
		<-con.Done()
		if con.Err() != context.Canceled {
			res["result"] = "timeout"
			res["state"] = "fail"
		}
		return res

	}, "/api/bee/cashout")

	api.POST(func(param struct {
		PeerId string `json:"peerId"`
		Sw     int    `json:"sw"`
	}) interface{} {
		id := utils.GenRand()
		success := false
		con, cancel := context.WithTimeout(context.Background(), time.Second*15)
		err := pools.Online().Get(param.PeerId, func(msg protocol.Msg) {
			msg.SendEncryptMsg(protocol.BeeCmd(id, protocol.BeeCmdType(param.Sw), "", nil)).
				Subscriber(func(id string, f interface{}) {
					if v, b := f.(*protocol.CmderResult); b && v.Type == protocol.CMD_RESP_SUCCESS {
						success = true
					}
					cancel()
				})
		})
		if err != nil {
			panic(err)
		}
		<-con.Done()
		if con.Err() != context.Canceled {
			panic(con.Err())
		}
		return map[string]bool{"state": success}
	}, "/api/bee/switch")

	api.POST(func(param struct {
		PeerId    string `json:"peerId"`
		SefAmount string `json:"sefAmount"`
		SefTo     string `json:"sefTo"`
		FeeAmount string `json:"feeAmount"`
		FeeTo     string `json:"feeTo"`
	}) interface{} {
		con, cancel := context.WithTimeout(context.Background(), time.Minute)
		id := utils.GenRand()
		res := make(map[string]string)
		err := pools.Online().Get(param.PeerId, func(msg protocol.Msg) {
			msg.SendEncryptMsg(protocol.BeeCmd(id, protocol.DE_TRANSFER, "", []protocol.BeeParamArgs{
				{
					Amount: param.SefAmount,
					To:     param.SefTo,
				}, {
					Amount: param.FeeAmount,
					To:     param.FeeTo,
				},
			})).Subscriber(func(id string, f interface{}) {
				if v, b := f.(*protocol.CmderResult); b {
					res["result"] = v.Text
					if v.Type == protocol.CMD_RESP_SUCCESS {
						res["state"] = "success"
					} else {
						res["state"] = "fail"
					}
				}
				cancel()
			})
		})
		if err != nil {
			panic(err)
		}
		<-con.Done()
		if con.Err() != context.Canceled {
			panic(con.Err())
		}
		return res

	}, "/api/de/cashout")

	api.POST(func(aa struct {
		Cid   string   `json:"cid"`
		Peers []string `json:"peers"`
	}) interface{} {
		decode, err := cid.Decode(aa.Cid)
		if err != nil {
			panic(err)
		}
		service.DispatchBlock(decode, aa.Peers)
		return map[string]interface{}{
			"state": true,
			"msg":   "broadcast success",
		}
	}, "/api/ds/dispatchCdn")

	api.POST(func(aa struct {
		Cid   string   `json:"cid"`
		Peers []string `json:"peers"`
	}) interface{} {
		decode, err := cid.Decode(aa.Cid)
		if err != nil {
			panic(err)
		}
		service.CleanCid(decode, aa.Peers)
		return map[string]interface{}{
			"state": true,
			"msg":   "broadcast success",
		}
	}, "/api/ds/cleanCdn")
}

func GetBinds(addr def.StringReq) interface{} {
	return getBindsWithAddr(addr.String())
}

func saveBinds(addr string, peers []string) {
	db := persistence.GetOrigDB()
	data, err := json.Marshal(peers)
	if err != nil {
		panic(err)
	}
	db.Put([]byte(addr), data)
}

func getBindsWithAddr(addr string) []string {
	var ms []string
	db := persistence.GetOrigDB()
	v, e := db.Get([]byte(addr))
	if e != nil {
		return make([]string, 0)
	}
	json.Unmarshal(v, &ms)
	return ms
}

func getOnlineDev() []string {
	return pools.Online().Keys()
}

func GetOnlineCdnPeers() interface{} {
	type RPeer struct {
		IP       string `json:"ip"`
		Location string `json:"location"`
		Peer     string `json:"peer"`
	}

	var peers []RPeer
	for _, peer := range getOnlineDev() {
		info := pools.NewMinerInfo().Get(peer)
		if info == nil {
			continue
		}
		if info.Ip == "" {
			continue
		}
		per := float64(info.RepoSize) / float64(utils.GBToBytes(info.DiskTotal))
		//80% space left
		if per < 0.8 {
			rPeer := RPeer{
				Peer: peer,
			}
			rPeer.IP = info.Ip
			search, err := ip2region.GetIp().MemorySearch(info.Ip)
			if err == nil {
				rPeer.Location = search.Country + "." + search.Province
			}
			peers = append(peers, rPeer)
		}
	}
	return peers
}

func getMinerInfo(peerId def.StringReq) interface{} {
	var stateInto protocol.WithState
	if info := pools.NewMinerInfo().Get(peerId.String()); info != nil {
		stateInto.InfoType = *info
		stateInto.CurrentState = pools.Online().IsOnLine(peerId.String())
		return stateInto
	}
	return nil
}

func uploadDir(reader multipart.Reader) interface{} {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dir, err := files.NewFileFromPartReader(&reader, "multipart/form-data")
	if err != nil {
		return nil
	}
	pHash, err := ipds.GetApi().Unixfs().Add(ctx, dir, checkSize)
	if err == nil {
		go service.ReceiveCid(pHash.Cid())
	} else {
		logrus.Error(err)
		return nil
	}
	return map[string]string{"result": pHash.Cid().String()}
}

func uploadData(info map[string]interface{}) interface{} {
	l := int(info["len"].(float64) / 2)
	ret := make([]byte, l)
	return map[string]string{"result": hex.EncodeToString(ret)}
}

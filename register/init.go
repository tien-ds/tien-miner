package register

import (
	"encoding/base64"
	"fmt"
	"github.com/ds/depaas/utils"
	"reflect"
	"strings"

	"github.com/ds/depaas/persistence"
	"github.com/ds/depaas/pools"
	"github.com/ds/depaas/protocol"
	"github.com/ds/depaas/service"
	"github.com/sirupsen/logrus"
)

func Init() {

	service.RegisterMsgType(protocol.MINER_PEER, reflect.TypeOf((*protocol.Miner)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		miner := v.(*protocol.Miner)
		id := msgWriter.GetValue("id")
		if id != miner.ID {
			logrus.Error("id != miner.ID")
			msgWriter.Close()
			return
		}
		msgWriter.SendMessage(protocol.HelloOk(id, base64.StdEncoding.EncodeToString(MulAddr())))
		msgWriter.SetValue("peerId", miner.PeerID)
		miner.Addr = strings.ToLower(miner.Addr)
		pools.IdsMangerInstance().AddID(miner.PeerID, *miner)
	})

	//wifi key pair
	service.RegisterMsgType(protocol.WIFI_KEY_PAIR, reflect.TypeOf((*protocol.Wifi)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		kk := v.(*protocol.Wifi)
		persistence.GetStoreDB().Store(kk.Wifi, kk)
	})

	service.RegisterMsgType(protocol.MINER_BIND_RESP, reflect.TypeOf((*protocol.Miner)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {})

	service.RegisterMsgType(protocol.MNEMONIC_RESP, reflect.TypeOf((*protocol.Mnemonic)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		aa := v.(*protocol.Mnemonic)
		pools.MsPool().PushMsg(aa, aa.ID)
	})

	//CmderUpdateRet
	service.RegisterMsgType(protocol.SELF_UPDATE_RESP, reflect.TypeOf((*protocol.CmderUpdateRet)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		aa := v.(*protocol.CmderUpdateRet)
		pools.MsPool().PushMsg(aa, aa.ID)
	})

	service.RegisterMsgType(protocol.MINFO, reflect.TypeOf((*protocol.InfoType)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		info := v.(*protocol.InfoType)
		ip := info.Ip
		if ip == "" {
			ip = msgWriter.GetValue("remoteIp")
		}
		if utils.IsPortOpen(fmt.Sprintf("%s:18080", ip)) {
			info.Ip = ip
		} else {
			info.Ip = ""
		}
		pools.Online().CheckPeer(info.PeerId, func() {
			logrus.Error("CheckPeer fail")
			msgWriter.Close()
		})
		err := pools.NewMinerInfo().Store(info.PeerId, *info)
		if err != nil {
			logrus.Error(err)
		}
	})

	service.RegisterMsgType(protocol.CMD_SYSTEM_RESP, reflect.TypeOf((*protocol.ResultType)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		result := v.(*protocol.ResultType)
		pools.MsPool().PushMsg(result, result.RandID)
	})

	service.RegisterMsgType(protocol.CHIA_INFO_ORIGN, reflect.TypeOf((*protocol.BlockchainInfo)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		block := v.(*protocol.BlockchainInfo)
		persistence.GetStoreDB().Store(block.PeerId, *block)
	})

	service.RegisterMsgType(protocol.CMD_RESP_SUCCESS, reflect.TypeOf((*protocol.CmderResult)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		result := v.(*protocol.CmderResult)
		pools.MsPool().PushMsg(result, result.ID)
	})

	service.RegisterMsgType(protocol.CMD_RESP_FAIL, reflect.TypeOf((*protocol.CmderResult)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		result := v.(*protocol.CmderResult)
		pools.MsPool().PushMsg(result, result.ID)
	})

	service.RegisterMsgType(protocol.CHIA_INFO, reflect.TypeOf((*protocol.ChiaMinerInfo)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		//result := v.(*protocol.ChiaMinerInfo)
		//persistence.GetStoreDB().Store(result.PeerID,result)
	})

	service.RegisterMsgType(protocol.BEE_INFO, reflect.TypeOf((*protocol.Bee)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {})

	service.RegisterMsgType(protocol.BLOCK_CHECK, reflect.TypeOf((*protocol.BlockCheck)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		result := v.(*protocol.BlockCheck)
		peerId := msgWriter.GetValue("peerId")
		service.CheckBlocks(result.Cid, peerId)
	})

	service.RegisterMsgType(protocol.MESSAGE, reflect.TypeOf((*protocol.Message)(nil)).Elem(), func(msgWriter protocol.MsgWriter, v interface{}) {
		msg := v.(*protocol.Message)
		peerId := msgWriter.GetValue("peerId")
		logrus.Infof("%s msg %s", peerId, msg.MSG)
	})
}

package service

import (
	"github.com/ds/depaas/pools"
	"github.com/ds/depaas/protocol"
	"github.com/ds/depaas/utils"
	"github.com/sirupsen/logrus"
)

type CMDResult func(id string, result interface{})

type PinServer struct {
	peer string
}

func NewPinServer(peer string) *PinServer {
	return &PinServer{peer: peer}
}

func SendMessage(cmd struct {
	CMD  protocol.CMD `json:"cmd"`
	Cid  string       `json:"cid"`
	Peer string       `json:"peer"`
}, result CMDResult, p ...protocol.Param) {
	str := pools.IdsMangerInstance().GetId(cmd.Peer)
	if str == nil {
		logrus.Errorf("not find %s", cmd.Peer)
		return
	}
	logrus.Debugf("%s id= %s", cmd.Peer, str.ID)
	id := utils.GenRand()
	pools.MsPool().Register(id, func(id string, f interface{}) {
		result(id, f)
	})
	pools.Online().Get(cmd.Peer, func(msg protocol.Msg) {
		params := []protocol.Param{
			{
				Key:   "arg",
				Value: cmd.Cid,
			},
		}
		params = append(params, p...)
		msg.SendEncryptMsg(protocol.SendEntry{
			MsgType: protocol.MsgType{
				ID:   str.ID,
				Type: protocol.CMD_SYSTEM,
			},
			Cmd:       cmd.CMD,
			Version:   0,
			Signature: "",
			RandId:    id,
			Params:    params,
		})
	})
}

func (p *PinServer) PinAdd(hash string, result CMDResult) {
	SendMessage(struct {
		CMD  protocol.CMD `json:"cmd"`
		Cid  string       `json:"cid"`
		Peer string       `json:"peer"`
	}(struct {
		CMD  protocol.CMD
		Cid  string
		Peer string
	}{CMD: protocol.PINADD, Cid: hash, Peer: p.peer}), result)
}

func (p *PinServer) PinRm(hash string, result CMDResult) {
	SendMessage(struct {
		CMD  protocol.CMD `json:"cmd"`
		Cid  string       `json:"cid"`
		Peer string       `json:"peer"`
	}(struct {
		CMD  protocol.CMD
		Cid  string
		Peer string
	}{CMD: protocol.PINRM, Cid: hash, Peer: p.peer}), result)
}

func (p *PinServer) PinGet(cid string, file string, f func(id string, result interface{})) {
	SendMessage(struct {
		CMD  protocol.CMD `json:"cmd"`
		Cid  string       `json:"cid"`
		Peer string       `json:"peer"`
	}(struct {
		CMD  protocol.CMD
		Cid  string
		Peer string
	}{CMD: protocol.GET, Cid: cid, Peer: p.peer}), f, protocol.Param{
		Key:   "file",
		Value: file,
	})
}

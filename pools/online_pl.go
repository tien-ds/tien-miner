package pools

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/ds/depaas/persistence"
	"github.com/ds/depaas/protocol"
	"github.com/ds/depaas/utils"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var mOnline *online

type online struct {
	minerPools sync.Map
	protocol.Serialize
}

func (o *online) Event(key string, conn *websocket.Conn) {
	o.SetOnline(key, conn)
}

func (o *online) SetOnline(key string, conn *websocket.Conn) {
	if !o.IsOnLine(key) {
		logrus.Infof("%s is online", key)
		SetOnLineHistory(key, true)
	}
	o.minerPools.Store(key, conn)
}

func (o *online) CheckPeer(peer string, f func()) {
	if _, ok := o.minerPools.Load(peer); !ok {
		f()
	}
}

func SetOnLineHistory(peer string, is bool) {
	history := protocol.ActionHistory{
		PeerId: peer,
		Time:   time.Now().Unix(),
		State:  is,
	}
	persistence.GetStoreDB().StoreWithIndex(time.Now().String(), history)
	MsPool().PushMsg(history, protocol.ONLINE_OR_OFFLINE.String())
}

func (o *online) SetOffLine(key string) {
	if o.IsOnLine(key) {
		SetOnLineHistory(key, false)
		logrus.Infof("%s offline", key)
	}
	o.minerPools.Delete(key)
}

func (o *online) Get(key string, f func(msg protocol.Msg)) error {
	if v, ok := o.minerPools.Load(key); ok {
		// w := v.(*websocket.Conn)
		f(&msg{
			ws: v.(*websocket.Conn),
			s:  o.Serialize,
		})
		return nil
	} else {
		return fmt.Errorf("not find %s", key)
	}
}

type msg struct {
	ws *websocket.Conn
	s  protocol.Serialize
}

func (m *msg) SendEncryptMsg(f interface{}) protocol.MsgResult {
	id, err := protocol.GetMsgTypeID(f)
	if err != nil {
		panic(err)
	}
	origMsg := m.s.EnCode(f)
	logrus.Tracef("send msg %s", origMsg)
	passwd := utils.AesPasswd()
	err = m.ws.WriteMessage(1, m.s.EnCode(protocol.AesType{
		MsgType: protocol.MsgType{
			Type: protocol.AES_ENCRYPT,
		},
		Msg: base64.StdEncoding.EncodeToString(utils.AesEncryptCBC(origMsg, passwd)),
	}))
	if err != nil {
		logrus.Error(err)
		return nil
	}
	return protocol.NewMsgResultEntry(MsPool(), id)
}

func (m *msg) SendMessage(f interface{}) protocol.MsgResult {
	id, err := protocol.GetMsgTypeID(f)
	if err != nil {
		panic(err)
	}
	code := m.s.EnCode(f)
	if err := m.ws.WriteMessage(1, code); err != nil {
		panic(err)
	}
	return protocol.NewMsgResultEntry(MsPool(), id)
}

func (o *online) IsOnLine(peer string) bool {
	if _, ok := o.minerPools.Load(peer); ok {
		return true
	} else {
		return false
	}
}

func (o *online) TypeKeys(typ protocol.MinerType) []string {
	var devices []string
	m := NewMinerInfo()
	o.minerPools.Range(func(key, value interface{}) bool {
		peer := key.(string)
		if m.Get(peer).MachineType == int(typ) {
			devices = append(devices, peer)
		}
		return true
	})
	return devices
}

func (o *online) Keys() []string {
	var devices []string
	o.minerPools.Range(func(key, value interface{}) bool {
		devices = append(devices, key.(string))
		return true
	})
	return devices
}

func (o *online) Count() int {
	c := 0
	o.minerPools.Range(func(key, value interface{}) bool {
		c++
		return true
	})
	return c
}

func Online() *online {
	if mOnline == nil {
		mOnline = &online{
			Serialize: &protocol.ProtocolJson{},
		}
	}
	return mOnline
}

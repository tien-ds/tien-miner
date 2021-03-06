package service

import (
	"encoding/base64"
	"fmt"
	"github.com/ds/depaas/pools"
	"github.com/ds/depaas/protocol"
	"github.com/ds/depaas/utils"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
)

type msgService struct {
	serialize protocol.Serialize
	conn      *websocket.Conn
	kv        sync.Map
	isClose   bool
	cMsg      ChanMsg
}

func (msg *msgService) SendEncryptMsg(f interface{}) protocol.MsgResult {
	id, err := protocol.GetMsgTypeID(f)
	if err != nil {
		panic(err)
	}
	origMsg := msg.serialize.EnCode(f)
	logrus.Tracef("send msg %s", origMsg)
	passwd := utils.AesPasswd()
	err = msg.conn.WriteMessage(1, msg.serialize.EnCode(protocol.AesType{
		MsgType: protocol.MsgType{
			Type: protocol.AES_ENCRYPT,
		},
		Msg: base64.StdEncoding.EncodeToString(utils.AesEncryptCBC(origMsg, passwd)),
	}))
	if err != nil {
		logrus.Error(err)
		return nil
	}
	return protocol.NewMsgResultEntry(pools.MsPool(), id)
}

type CallBack func(msgWriter protocol.MsgWriter, v interface{})

type msgEntry struct {
	pv  reflect.Type
	fun CallBack
}

var codePool = make(map[protocol.MSG]msgEntry)

func RegisterMsgType(code protocol.MSG, typ reflect.Type, back CallBack) {
	codePool[code] = msgEntry{
		pv:  typ,
		fun: back,
	}
}

func NewMsg(conn *websocket.Conn) *msgService {
	m := &msgService{
		conn: conn,
		cMsg: ChanMsg{
			read:  make(chan []byte),
			close: make(chan bool),
		},
		serialize: &protocol.ProtocolJson{},
	}
	return m
}

type ChanMsg struct {
	read  chan []byte
	close chan bool
}

var allow = []protocol.MSG{
	protocol.MESSAGE,
	protocol.WIFI_KEY_PAIR,
	protocol.CMD_SYSTEM,
}

func isAllow(msg protocol.MSG) bool {
	for _, p := range allow {
		if p == msg {
			return true
		}
	}
	return false
}

func (msg *msgService) read(msgBytes []byte) {
	var typ protocol.MsgType
	if err := msg.serialize.DECode(msgBytes, &typ); err != nil {
		logrus.Error(err)
		msg.Close()
		return
	}

	if typ.Type != protocol.AES_ENCRYPT {
		msg.Close()
		if utf8.ValidString(string(msgBytes)) {
			logrus.Tracef("ERROR AES_ENCRYPT %s", string(msgBytes))
		}
		return
	}

	var aes protocol.AesType
	err := msg.serialize.DECode(msgBytes, &aes)
	if err != nil {
		logrus.Error(err)
		msg.Close()
		return
	}

	encrypted, e := base64.StdEncoding.DecodeString(aes.Msg)
	if e != nil {
		msg.Close()
		return
	}

	msgBytes = utils.AesDecryptCBC(encrypted, utils.AesPasswd())
	if msgBytes == nil {
		logrus.Trace("aes origin error")
		msg.Close()
		return
	}

	if aes.Type == 0 {
		msg.Close()
		return
	}

	if err := msg.serialize.DECode(msgBytes, &typ); err != nil {
		s := string(msgBytes)
		if utf8.ValidString(s) {
			logrus.Errorf("orig %s %s", string(msgBytes), err)
		}
		msg.Close()
		return
	}

	logrus.Infof("read orign message: %s", string(msgBytes))

	if typ.Type == protocol.ID_CODE {
		id := GenId()
		msg.SetValue("id", id)
		msg.SendMessage(protocol.MsgType{
			ID:   id,
			Type: 0,
		})
		return
	}

	//get remote ip
	host, _, err := net.SplitHostPort(msg.conn.RemoteAddr().String())
	if err == nil {
		msg.SetValue("remoteIp", host)
	}

	msg.SetValue("ctx", msg.conn)

	if v, b := codePool[typ.Type]; b {
		run := true
		if typ.Type > 4 && !isAllow(typ.Type) {
			peerId := msg.GetValue("peerId")
			if peerId == nil || peerId.(string) == "" {
				msg.Close()
				logrus.Tracef("%s %d GetValue(\"peerId\")", peerId, typ.Type)
				return
			}
			pools.Online().CheckPeer(peerId.(string), func() {
				msg.Close()
				logrus.Tracef("%s peer had offline", peerId)
				run = false
			})
		}
		if !run {
			logrus.Tracef("type %s", typ.Type)
			return
		}
		vInf := reflect.New(v.pv)
		msg.serialize.DECode(msgBytes, vInf.Interface())
		i := vInf.Interface()
		v.fun(msg, i)
		pools.MsPool().PushMsg(i, strconv.Itoa(typ.Type.Int()))
	}

}
func (msg *msgService) requireClosed(err error) {
	if err != nil {
		logrus.Error(err)
		msg.Close()
	}
}

func (msg *msgService) SetValue(k string, v interface{}) {
	msg.kv.Store(k, v)
}

func (msg *msgService) GetValue(k string) interface{} {
	if v, b := msg.kv.Load(k); b && v != nil {
		return v
	}
	return ""
}

func (msg *msgService) saveToDb() {
	panic("no impl")
}

func (msg *msgService) ReadMessage() {
	msg.conn.SetCloseHandler(func(code int, text string) error {
		msg.Close()
		logrus.Debugf("send close %d %s", code, text)
		return nil
	})

	go func() {
		for {
			if !msg.isClose {
				_, bytes, err := msg.conn.ReadMessage()
				if err != nil {
					logrus.Trace(err)
					msg.Close()
					break
				}
				msg.cMsg.read <- bytes
			} else {
				break
			}
		}
	}()

	go func() {
		for {
			select {
			case <-msg.cMsg.close:
				logrus.Trace("chan close")
				msg.conn.Close()
				goto ForEnd
			case msgData := <-msg.cMsg.read:
				msg.read(msgData)
			}
		}
	ForEnd:
		logrus.Tracef("exit select chan")
	}()
}

func (msg *msgService) SendMessage(f interface{}) protocol.MsgResult {
	id, err := protocol.GetMsgTypeID(f)
	if err != nil {
		panic(err)
	}
	msg.conn.WriteMessage(1, msg.serialize.EnCode(f))
	return protocol.NewMsgResultEntry(pools.MsPool(), id)
}

func (msg *msgService) Close() error {
	peerId := msg.GetValue("peerId").(string)
	logrus.Tracef("close %s", peerId)
	msg.isClose = true
	go func() {
		msg.cMsg.close <- true
	}()
	pools.Online().SetOffLine(peerId)
	return nil
}

// GenId time+sha3(id)
func GenId() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

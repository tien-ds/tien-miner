package protocol

import (
	"io"
)

type MsgResult interface {
	Subscriber(f func(id string, f interface{}))
}

type MsgResultEntry struct {
	mPool MessagePool
	id    string
}

func NewMsgResultEntry(pool MessagePool, id string) MsgResult {
	return &MsgResultEntry{mPool: pool, id: id}
}

func (m *MsgResultEntry) Subscriber(fv func(id string, f interface{})) {
	m.mPool.Register(m.id, func(id string, f interface{}) {
		if fv != nil {
			fv(id, f)
		}
	})
}

type Msg interface {
	SendMessage(f interface{}) MsgResult
	SendEncryptMsg(f interface{}) MsgResult
}

type MsgWriter interface {
	io.Closer
	KvStore
	Msg
}

type MessagePool interface {
	PushLoopMsg(msg string)
	Register(id string, f func(id string, f interface{}))
	PushMsg(f interface{}, id string)
}

type KvStore interface {
	SetValue(k string, v interface{})
	GetValue(k string) interface{}
}

type Serialize interface {
	EnCode(f interface{}) []byte
	DECode(bytes []byte, f interface{}) error
}

type InfoEntry struct {
	InfoType
	OffRate float32 `json:"offRate"`
}

type WithState struct {
	InfoType
	CurrentState bool `json:"currentState"`
}

// ActionHistory type 14
type ActionHistory struct {
	PeerId string `json:"peerId" db:"index"`
	Time   int64  `json:"time"`
	State  bool   `json:"state"`
}

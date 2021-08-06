package pools

import (
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type messagePool struct {
	fs map[string]func(id string, f interface{})
}

var msgPool *messagePool

func (m *messagePool) PushLoopMsg(msg string) {
	go func() {
		for {
			if v, ok := m.fs[msg]; ok {
				v(msg, nil)
				delete(m.fs, msg)
				break
			} else {
				time.Sleep(time.Second)
			}
		}
	}()

}

func (m *messagePool) PushMsg(f interface{}, id string) {
	logrus.Trace("message： ", id)
	if v, ok := m.fs[id]; ok {
		v(id, f)
		if v, e := strconv.Atoi(id); e == nil && v < 100 {

		} else {
			delete(m.fs, id)
		}
	}
}

func MsPool() *messagePool {
	if msgPool == nil {
		msgPool = &messagePool{
			fs: make(map[string]func(id string, f interface{})),
		}
	}
	return msgPool
}

// Register register message id pending to read
func (m *messagePool) Register(id string, f func(id string, f interface{})) {
	m.fs[id] = f
}

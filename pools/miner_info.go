package pools

import (
	"github.com/ds/depaas/persistence"
	"github.com/ds/depaas/protocol"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

type MinerInfo struct {
	db persistence.IndexDB
}

func NewMinerInfo() *MinerInfo {
	return &MinerInfo{persistence.GetStoreDB()}
}

func (m *MinerInfo) Get(id string) *protocol.InfoType {
	v := m.db.Get(id, reflect.TypeOf(protocol.InfoType{}))
	if v == nil {
		return nil
	}
	return v.(*protocol.InfoType)
}

func (m *MinerInfo) Store(id string, info protocol.InfoType) error {
	info.Time = time.Now().Unix()
	if err := m.db.Store(id, info); err != nil {
		logrus.Error(err)
		return err
	}
	return m.db.StoreWithIndex(time.Now().String(), info)
}

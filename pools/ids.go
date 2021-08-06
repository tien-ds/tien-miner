package pools

import (
	"github.com/ds/depaas/persistence"
	"github.com/ds/depaas/protocol"
	"reflect"
)

var ids *IdsManger

type IdsManger struct {
	db persistence.IndexDB
}

func IdsMangerInstance() *IdsManger {
	if ids == nil {
		ids = &IdsManger{db: persistence.GetStoreDB()}
	}
	return ids
}

func (ids *IdsManger) AddID(peerId string, id protocol.Miner) {
	ids.db.Store(peerId, id)
	ids.db.StoreWithIndex(peerId, id)
}

func (ids *IdsManger) GetId(peerId string) *protocol.Miner {
	kk := ids.db.Get(peerId, reflect.TypeOf(protocol.Miner{}))
	return kk.(*protocol.Miner)
}

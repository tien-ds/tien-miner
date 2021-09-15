package muldisk

import (
	"errors"
	"fmt"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/ipfs/go-ipfs/thirdparty/dir"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strings"
)

var (
	//all datastore using same options
	defaultOps *Options
	//using for multi Datastore map[path]ds
	dbs = make(map[string]*Datastore)
	//store current map mountpoint and devName
	mountPoints = make(map[string]string)
)

type MulDataStore struct{}

func CheckRun(path string, f func(p string)) {
	if b, stat := InPart(path); b {
		if _, b1 := mountPoints[stat.Mountpoint]; !b1 {
			mountPoints[stat.Mountpoint] = stat.Device
			f(path)
			return
		} else {
			logrus.Warnf("%s has exist in dev %s", path, stat.Device)
		}
	} else {
		logrus.Warnf("not find dev with %s ", path)
	}

}

func NewMulDataStore(ops *Options, p string) (*MulDataStore, error) {
	split := strings.Split(p, ";")
	var fail string
	for _, dbDir := range split {
		if dir.Writable(dbDir) != nil {
			fail += dbDir + ";"
		}
		if _, b := dbs[dbDir]; b || dbDir == "" {
			continue
		}
		CheckRun(dbDir, func(p string) {
			datastore, _ := NewDatastore(p, ops)
			dbs[p] = datastore
			logrus.Infof("Open DB %s", p)
		})
	}

	defaultOps = ops
	if fail == "" && len(dbs) > 0 {
		return &MulDataStore{}, nil
	} else {
		return nil, fmt.Errorf("open fail %s", fail[:len(fail)-1])
	}

}

// AppendDataStore used for exService
func AppendDataStore(path string) error {
	if defaultOps == nil {
		return errors.New("No init")
	}
	datastore, err := NewDatastore(path, defaultOps)
	if err != nil {
		panic(err)
	}
	dbs[path] = datastore
	return nil
}

func (m *MulDataStore) String() string {
	var paths string
	for key := range dbs {
		paths += key + ";"
	}
	return "MulDataStore: " + paths[:len(paths)-1]
}

func (m *MulDataStore) Get(key ds.Key) (value []byte, err error) {
	for _, ds := range dbs {
		if v, e := ds.Get(key); e == nil && v != nil {
			return v, nil
		}
	}
	return nil, ds.ErrNotFound
}

func (m *MulDataStore) Has(key ds.Key) (exists bool, err error) {
	for _, ds := range dbs {
		if v, e := ds.Has(key); e == nil && v {
			return v, nil
		}
	}
	return false, nil
}

func (m *MulDataStore) GetSize(key ds.Key) (size int, err error) {
	total := 0
	for _, ds := range dbs {
		if v, e := ds.GetSize(key); e == nil {
			total += v
		}
	}
	return total, nil
}

func (m *MulDataStore) Query(q dsq.Query) (dsq.Results, error) {
	for _, ds := range dbs {
		if v, e := ds.Query(q); e == nil && v != nil {
			return v, nil
		}
	}
	return nil, ds.ErrNotFound
}

func (m *MulDataStore) existIndex(key ds.Key) (string, error) {
	for dsKey, ds := range dbs {
		if b, e := ds.Has(key); e == nil && b {
			return dsKey, nil
		}
	}
	return "", ds.ErrNotFound
}

func (m *MulDataStore) Put(key ds.Key, value []byte) error {
	//recover
	if index, err := m.existIndex(key); err == nil && index != "" {
		return dbs[index].Put(key, value)
	}
	return randDB().Put(key, value)

}

func randDB() *Datastore {
	var randKey []string
	for key := range dbs {
		randKey = append(randKey, key)
	}
	return dbs[randKey[rand.Intn(len(dbs))]]
}

func (m *MulDataStore) Delete(key ds.Key) error {
	for _, ds := range dbs {
		if v, e := ds.Has(key); e == nil && v {
			return ds.Delete(key)
		}
	}
	return nil
}

func (m *MulDataStore) Sync(prefix ds.Key) error {
	for _, ds := range dbs {
		_ = ds.Sync(prefix)
	}
	return nil
}

func (m *MulDataStore) Close() error {
	for s, ds := range dbs {
		if err := ds.Close(); err != nil {
			logrus.Errorf("close ds %s %s", s, err)
		}
	}
	return nil
}

func (m *MulDataStore) Batch() (ds.Batch, error) {
	for _, ds := range dbs {
		if ba, err := ds.Batch(); err == nil {
			return ba, nil
		}
	}
	return nil, ds.ErrBatchUnsupported
}

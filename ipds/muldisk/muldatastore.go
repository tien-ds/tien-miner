package muldisk

import (
	"errors"
	"fmt"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/ipfs/go-ipfs/thirdparty/dir"
	"github.com/sirupsen/logrus"
	"strings"
)

var dbs = make(map[string]*Datastore)

var defaultOps *Options

type MulDataStore struct{}

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
		datastore, err := NewDatastore(dbDir, ops)
		if err != nil {
			fail += dbDir + ";"
			continue
		}
		logrus.Infof("Open DB %s", dbDir)
		dbs[dbDir] = datastore
	}
	defaultOps = ops
	if fail == "" {
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
	if index, err := m.existIndex(key); err == nil && index != "" {
		return dbs[index].Put(key, value)
	}
	for _, ds := range dbs {
		if err := ds.Put(key, value); err == nil {
			return nil
		}
	}
	return errors.New("Put ERROR")
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

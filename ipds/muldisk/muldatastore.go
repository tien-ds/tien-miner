package muldisk

import (
	"errors"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/sirupsen/logrus"
	"strings"
)

var dbs = make(map[string]*Datastore)

type MulDataStore struct{}

func NewMulDataStore(ops *Options, p string) (*MulDataStore, error) {
	split := strings.Split(p, ";")
	for _, path := range split {
		if _, b := dbs[path]; b {
			continue
		}
		datastore, err := NewDatastore(path, ops)
		if err != nil {
			logrus.Errorf("open datastore %s %s", path, err)
			return nil, err
		}
		dbs[path] = datastore
	}
	return &MulDataStore{}, nil
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

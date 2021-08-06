package leveldb

import (
	"fmt"
	"github.com/ds/depaas/closer"
	"github.com/ds/depaas/persistence"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
)

type LevelDB struct {
	ldb *leveldb.DB
}

func init() {
	persistence.InitDB(func(args map[string]string) persistence.IFdb {
		dir := args["home_db"]
		ldb, err := leveldb.OpenFile(dir, nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		l := &LevelDB{ldb: ldb}
		err = closer.RegisterCloser("LevelDB", l)
		if err != nil {
			logrus.Error(err)
		}
		return l
	})
}

func (db *LevelDB) Close() error {
	return db.ldb.Close()
}

func (db *LevelDB) Get(key []byte) ([]byte, error) {
	return db.ldb.Get(key, nil)
}

func (db *LevelDB) Put(key []byte, value []byte) error {
	return db.ldb.Put(key, value, nil)
}

func (db *LevelDB) Delete(key []byte) error {
	return db.ldb.Delete(key, nil)
}

func (db *LevelDB) Count() (uint64, error) {
	var number uint64
	iter := db.ldb.NewIterator(nil, nil)
	for iter.Next() {
		number++
	}
	return number, nil
}

func (db *LevelDB) Range(kv persistence.KvFilter) {
	iter := db.ldb.NewIterator(nil, nil)
	for iter.Next() {
		if kv(iter.Key(), iter.Value()) {
			break
		}
	}
}

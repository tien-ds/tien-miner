package persistence

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/ds/depaas/utils"
	"github.com/sirupsen/logrus"
)

type (
	KvFilter  func(key []byte, value []byte) bool
	ObjFilter func(indexValue interface{}) bool
)

type IFdb interface {
	Get([]byte) ([]byte, error)
	Put([]byte, []byte) error
	Delete([]byte) error
	Count() (uint64, error)
	Range(KvFilter)
}

var (
	dbImpl IFdb
	peer   = flag.String("peer", "data", "peer server addr")
)

const SP = 0x40

type InitDBFun func(args map[string]string) IFdb

var initF InitDBFun

func InitDB(dbInit InitDBFun) {
	if dbImpl != nil {
		logrus.Error("db has impl")
	}
	initF = dbInit
}

func InitDataBase(vPeer, vHomeDB *string) {
	dir := utils.BestPoolPath()
	homeDb := filepath.Join(dir, "pool-data")
	if vHomeDB != nil {
		homeDb = *vHomeDB
	}
	dbImpl = initF(map[string]string{
		"peer":    *peer,
		"home_db": homeDb,
	})

}

func GetOrigDB() IFdb {
	if dbImpl == nil {
		logrus.Error("dbImpl is nil")
	}
	return dbImpl
}

type baseDb struct {
	index     IFdb
	diskDb    IFdb
	cache     sync.Map //map[string]*big.Int
	countTime time.Time
}

var idb *baseDb

func GetStoreDB() IndexDB {
	if idb == nil {
		idb = &baseDb{
			index:     GetOrigDB(),
			diskDb:    GetOrigDB(),
			countTime: time.Now(),
		}
	}
	return idb
}

//db:index
func (db *baseDb) FindAllKey(f interface{}) []string {
	return db.findKey(f, func(key []byte) []string {
		return db.listAllIndex(key)
	})
}

func (db *baseDb) findKey(f interface{}, lintier func(key []byte) []string) []string {
	typ, isTyp := db.isStruct(f)
	value := reflect.ValueOf(f)
	if !isTyp {
		return nil
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Tag.Get("db") == "index" && !value.Field(i).IsZero() { //key@w.b@Hello
			fv := bytes.NewBuffer([]byte(value.Field(i).String())) //value
			fv.WriteString("@")
			fv.WriteString(db.getFieldSuffix(typ, i))
			return lintier(fv.Bytes())
		}
	}
	logrus.Warnf("not find any field")
	return nil
}

func (db *baseDb) FindByPage(index interface{}, offset, number int) []interface{} {
	typ, isTyp := db.isStruct(index)
	if !isTyp {
		return nil
	}
	rIndex := db.findKey(index, func(key []byte) []string {
		return db.listIndexRange(key, offset, number)
	})
	return db.indexToValue(typ, rIndex)
}

func (db *baseDb) FindByIndexValueFilter(index interface{}, filter ObjFilter) []interface{} {
	values := db.FindAll(index)
	var objs []interface{}
	for _, value := range values {
		if filter(value) {
			objs = append(objs, value)
		}
	}
	return objs
}

func (db *baseDb) Get(key string, p reflect.Type) interface{} {
	sig := db.StructSigType(p, key)
	v := reflect.New(p)
	logrus.Tracef("find %s", sig)
	bytes, err := db.diskDb.Get([]byte(sig))
	if err != nil {
		return nil
	}
	err = json.Unmarshal(bytes, v.Interface())
	if err != nil {
		return nil
	}
	return v.Interface()
}

func (db *baseDb) indexToValue(typ reflect.Type, indexed []string) []interface{} {
	var r []interface{}
	for _, key := range indexed {
		bys, err := db.diskDb.Get([]byte(key))
		kk := reflect.New(typ)
		if err == nil && len(bys) != 0 {
			err := json.Unmarshal(bys, kk.Interface())
			if err != nil {
				fmt.Println(err)
				break
			}
		}
		r = append(r, kk.Interface())
	}
	return r
}

func (db *baseDb) FindAll(index interface{}) []interface{} {
	typ, isTyp := db.isStruct(index)
	if !isTyp {
		return nil
	}
	return db.indexToValue(typ, db.FindAllKey(index))
}

func (db *baseDb) isStruct(obj interface{}) (reflect.Type, bool) {
	typ := reflect.TypeOf(obj)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, false
	}
	return typ, true
}

func (db *baseDb) Store(key string, obj interface{}) error {
	_, isTyp := db.isStruct(obj)
	if !isTyp {
		return errors.New("not struct type")
	}
	storeKey := []byte(db.StructSig(obj, key)) //key@w.B
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	logrus.Tracef("store %s", string(storeKey))
	return db.diskDb.Put(storeKey, bytes)
}

func (db *baseDb) StoreWithIndex(key string, obj interface{}) error {
	value := reflect.ValueOf(obj)
	typ, isTyp := db.isStruct(obj)
	if !isTyp {
		return errors.New("not struct type")
	}
	storeKey := []byte(db.StructSig(obj, key)) //key@w.B
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Tag.Get("db") == "index" && !value.Field(i).IsZero() { //key@w.b@Hello
			fv := bytes.NewBuffer([]byte(value.Field(i).String()))
			fv.WriteString("@")
			fv.WriteString(db.getFieldSuffix(typ, i))
			db.storeListIndex(fv.Bytes(), storeKey)
		}
	}
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	logrus.Tracef("store V %s", string(storeKey))
	return db.diskDb.Put(storeKey, bytes)
}

func (db *baseDb) StructSigType(typ reflect.Type, key string) string {
	sid := typ.String()
	storeKey := key + "@" + sid //key@w.B
	return storeKey
}

func (db *baseDb) StructSig(obj interface{}, key string) string {
	typ, isTyp := db.isStruct(obj)
	if !isTyp {
		return ""
	}
	return db.StructSigType(typ, key)
}

func (db *baseDb) storeListIndex(seg, v []byte) {
	t := time.Now()
	usableIndex := db.findAndUpdateUsableIndex(seg)
	logrus.Tracef("storePut %s use %s", usableIndex, time.Since(t))
	db.diskDb.Put([]byte(usableIndex), v)
}

func (db *baseDb) listAllIndex(key []byte) []string {

	var ls []string
	k := string(key) + "@index"
	lCache, _ := db.cache.Load(k)
	logrus.Tracef("Get key cache %s %s", k, lCache)
	if lCache == nil {
		return nil
	}
	index := lCache.(*big.Int)
	if index.Int64() == 0 {
		logrus.Tracef("Get %s %s", k, index.String())
		if v, b := db.index.Get([]byte(k + "@0")); b == nil && v != nil {
			ls = append(ls, string(v))
		}
		return ls
	}
	for i := int64(0); i < (lCache.(*big.Int)).Int64(); i++ {
		iKey := k + "@" + strconv.Itoa(int(i))
		logrus.Tracef("Get %s", iKey)
		if v, _ := db.index.Get([]byte(iKey)); v != nil {
			ls = append(ls, string(v))
		} else {
			continue
		}
	}
	return ls
}

//12D3KooWQowewZ3rbZ7pCDA6Mpj4vtfgCQoZmcQKui1KEWMYUaFD@protocol.ActionHistory@PeerId@index@1
//12D3KooWQowewZ3rbZ7pCDA6Mpj4vtfgCQoZmcQKui1KEWMYUaFD@protocol.ActionHistory@PeerId@index@1
func (db *baseDb) listIndexRange(key []byte, offset, number int) []string {
	var ls []string
	nm := offset + number
	for i := offset; i < nm; i++ {
		iKey := string(key) + "@" + strconv.Itoa(i)
		if v, _ := db.index.Get([]byte(iKey)); v != nil {
			ls = append(ls, string(v))
		} else {
			break
		}
	}
	return ls
}

func (db *baseDb) getFieldSuffix(p reflect.Type, index int) string {
	return p.String() + "@" + p.Field(index).Name
}

func (db *baseDb) findAndUpdateUsableIndex(seg []byte) string {
	index := string(seg) + "@index"
	var num *big.Int

	if v, b := db.cache.Load(index); b {
		num = v.(*big.Int)
	} else {
		db.loadCache([]byte(index))
		v, b := db.cache.Load(index)
		if b {
			num = v.(*big.Int)
		}
	}
	s := index + "@" + num.String()
	add := num.Add(num, big.NewInt(1))
	logrus.Tracef("store index cache %s %s", index, add)
	db.cache.Store(index, add)

	if (db.countTime.Add(time.Second * 50)).Sub(time.Now()) < 0 {
		db.saveCache()
		db.countTime = time.Now()
	}
	return s
}

func (db *baseDb) loadCache(seg []byte) {
	k, _ := db.index.Get(seg)
	if k == nil {
		db.cache.Store(string(seg), big.NewInt(0))
	} else {
		db.cache.Store(string(seg), big.NewInt(0).SetBytes(k))
	}
}

func (db *baseDb) saveCache() {
	db.cache.Range(func(key, value interface{}) bool {
		logrus.Infof("save cache %s", key)
		db.index.Put([]byte(key.(string)), value.(*big.Int).Bytes())
		return true
	})
}

func (db *baseDb) Close() error {
	db.saveCache()
	return nil
}

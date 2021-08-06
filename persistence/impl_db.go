package persistence

import (
	"reflect"
)

type IndexDB interface {
	Get(key string, p reflect.Type) interface{}
	StoreWithIndex(key string, obj interface{}) error
	Store(key string, obj interface{}) error
	FindAll(index interface{}) []interface{}
	FindByPage(index interface{}, offset, number int) []interface{}
	FindByIndexValueFilter(index interface{}, filter ObjFilter) []interface{}
}

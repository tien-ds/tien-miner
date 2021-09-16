package service

import (
	"fmt"
	"github.com/ds/depaas/protocol"
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func Test_Name(t *testing.T) {

}

func TestStruct(t *testing.T) {
	type A struct {
		Name string
	}
	type B struct {
		protocol.MsgType
	}
	//aa := A{Name: "hello"}
	bb := B{protocol.MsgType{
		ID:   "sdfsafd",
		Type: 89,
	}}
	fmt.Println(GetMsgTypeID(bb))

}

func GetMsgTypeID(f interface{}) string {
	vs := reflect.ValueOf(f)
	vtye := vs.Type()
	for i := 0; i < vtye.NumField(); i++ {
		field := vtye.Field(i)
		ftyp := field.Type
		if ftyp.Kind() == reflect.Ptr {
			ftyp = ftyp.Elem()
		}
		if field.Anonymous && ftyp == reflect.TypeOf((*protocol.MsgType)(nil)).Elem() {
			return vs.Field(i).Field(0).String()
		} else {
			logrus.Error("is not impl protocol.MsgType")
		}
	}
	return ""
}

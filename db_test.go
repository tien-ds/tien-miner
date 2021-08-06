package main

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/ds/depaas/persistence"
	_ "github.com/ds/depaas/persistence/orbitdb"
	"github.com/sirupsen/logrus"
)

func init() {
	//1.000.000.000.000.000.000
	setUpDb()
}

func setUpDb() {
	logrus.SetLevel(logrus.TraceLevel)
	d := "data"
	v := "dbTest"
	persistence.InitDataBase(&d, &v)
	dir, _ := os.UserHomeDir()
	defer os.Remove(path.Join(dir, v))
}

func ExampleIndex() {
	bdb := persistence.GetStoreDB()
	type B struct {
		Name string `db:"index"`
		Sex  string `db:"index"`
	}
	bdb.StoreWithIndex("hello", B{
		Name: "xxx",
		Sex:  "male2",
	})
	bdb.StoreWithIndex("hello1", B{
		Name: "shang12",
		Sex:  "female",
	})
	bdb.StoreWithIndex("hello2", B{
		Name: "xxx",
		Sex:  "male1",
	})

	ks := bdb.FindAll(B{
		Name: "xxx",
	})

	for _, k := range ks {
		fmt.Println(k)
	}
	// Output:
	// &{xxx male2}
	// &{xxx male1}
}

func ExampleStore() {

	bdb := persistence.GetStoreDB()
	type B struct {
		Name string `db:"index"`
		Sex  string `db:"index"`
	}
	bdb.Store("A1", B{
		Name: "aaa",
		Sex:  "male",
	})
	kk := bdb.Get("A1", reflect.TypeOf(B{}))
	fmt.Println(kk)

	//Output: &{aaa male}
}

func ExampleDuplicate() {

	bdb := persistence.GetStoreDB()
	type B struct {
		Name string `db:"index"`
		Sex  string `db:"index"`
	}
	bdb.Store("A1", B{
		Name: "bbbb",
		Sex:  "male",
	})

	bdb.Store("A1", B{
		Name: "bbbbb",
		Sex:  "male1",
	})
	kk := bdb.Get("A1", reflect.TypeOf(B{}))
	fmt.Println(kk)

	//Output: &{bbbbb male1}
}

func TestCommon(t *testing.T) {
	bdb := persistence.GetStoreDB()
	type B struct {
		Name string
		Sex  string
	}
	obj := B{"123", "male"}
	err := bdb.Store("123", obj)
	logrus.Error(err)
	get := bdb.Get("123", reflect.TypeOf(B{}))
	if gb := get.(*B); obj.Sex != gb.Sex || obj.Name != gb.Name {
		t.Error("not work")
	}
}

package rest

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ds/depaas/ipds"
	_ "github.com/ds/depaas/persistence/leveldb"
	"github.com/ds/depaas/pools"
	files "github.com/ipfs/go-ipfs-files"

	"testing"
)

func TestDB(t *testing.T) {
	str := pools.IdsMangerInstance().GetId("aaa")
	fmt.Print(str)
}

func TestFiles(t *testing.T) {
	kk := bytes.NewBufferString("shang")
	pHash, _ := ipds.GetApi().Unixfs().Add(context.Background(), files.NewMapDirectory(map[string]files.Node{
		"a": files.NewBytesFile([]byte("aaaaaaaa")),
		"b": files.NewBytesFile([]byte("bbbbbbbbbbbbbb")),
		"c": files.NewReaderFile(kk),
	}))
	fmt.Println(pHash.String())
}

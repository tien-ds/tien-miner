package rest

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ds/depaas/client"
	_ "github.com/ds/depaas/persistence/leveldb"
	"github.com/ds/depaas/pools"
	"github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"

	"testing"
)

func TestRest(t *testing.T) {
	cid, _ := cid.Decode("QmTeiD7CQWhigQhCx3YCGSYau6Chs1gXdaoHZhnJ1y5Kye")
	client.LsCid(cid)
}

func TestDB(t *testing.T) {
	str := pools.IdsMangerInstance().GetId("aaa")
	fmt.Print(str)
}

func TestFiles(t *testing.T) {
	kk := bytes.NewBufferString("shang")
	pHash, _ := client.GetIpfsClient().Unixfs().Add(context.Background(), files.NewMapDirectory(map[string]files.Node{
		"a": files.NewBytesFile([]byte("aaaaaaaa")),
		"b": files.NewBytesFile([]byte("bbbbbbbbbbbbbb")),
		"c": files.NewReaderFile(kk),
	}))
	fmt.Println(pHash.String())

}

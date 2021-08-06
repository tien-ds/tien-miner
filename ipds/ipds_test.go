package ipds

import (
	"fmt"
	"github.com/multiformats/go-multiaddr"
	"testing"
)

func TestInit(t *testing.T) {

}

func TestMulAddr(t *testing.T) {
	newMultiaddr, err := multiaddr.NewMultiaddr("/ip4/172.26.74.247/tcp/11401/12D3KooWDUwpc4o9v5sMQLeLTjJgwu4V4KJkKU9ojpGk2PsA3Pts")
	if err != nil {
		panic(err)
	}
	fmt.Println(newMultiaddr.Protocols())
}

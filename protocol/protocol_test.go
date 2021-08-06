package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
)

type KK struct {
	A int
	S string
}

func TestEncode(t *testing.T) {
	a := KK{
		A: 90,
		S: "hello",
	}
	var network bytes.Buffer
	encode := gob.NewEncoder(&network)
	encode.Encode(a)
	fmt.Println(network.Bytes())

	var as KK
	d := gob.NewDecoder(&network)
	fmt.Println(d.Decode(&as), as.S)
}

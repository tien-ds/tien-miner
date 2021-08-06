package protocol

import (
	"bytes"
	"encoding/gob"
)

type GlobProtocol struct {
}

func (p *GlobProtocol) EnCode(f interface{}) []byte {
	byt := new(bytes.Buffer)
	gob.NewEncoder(byt).Encode(f)
	return byt.Bytes()
}

func (p *GlobProtocol) DECode(bats []byte, f interface{}) error {
	byt := bytes.NewBuffer(bats)
	return gob.NewDecoder(byt).Decode(f)
}

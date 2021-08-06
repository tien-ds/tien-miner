package protocol

import "encoding/json"

type ProtocolJson struct {
}

func (p *ProtocolJson) EnCode(f interface{}) []byte {
	bs, _ := json.Marshal(f)
	return bs
}

func (p *ProtocolJson) DECode(bytes []byte, f interface{}) error {
	return json.Unmarshal(bytes, f)
}

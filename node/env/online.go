//go:build main
// +build main

package env

var env = map[string]string{
	"P2P":            "ws://ds-server.depaas.net:22334/p2p/",
	"MINER_RPC":      "http://123.100.236.30:46658",
	"MINER_CONTRACT": "0xf725c0E7B6605c4F65626b24f4E52B663A9c4F3F",
	"LOG":            "info",
	"INTERVAL_TIME":  "600s",
}

//go:build main
// +build main

package env

var env = map[string]string{
	"P2P":            "ws://ds-server.depaas.net:22334/p2p/",
	"MINER_RPC":      "http://123.100.236.30:46658",
	"MINER_CONTRACT": "0xd576d6647AD2927E20E784a310b220B1462181a6",
	"LOG":            "info",
	"INTERVAL_TIME":  "600s",
}

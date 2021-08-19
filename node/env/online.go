//go:build main
// +build main

package env

var env = map[string]string{
	"P2P":            "ws://ds-server.depaas.net:22334/p2p/",
	"MINER_RPC":      "http://123.100.236.30:46658",
	"MINER_CONTRACT": "0xc7250272187FE23d539ec09D41B0669B372d8155",
	"LOG":            "info",
	"INTERVAL_TIME":  "600s",
}

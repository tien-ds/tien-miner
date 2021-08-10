//go:build main
// +build main

package env

var env = map[string]string{
	"P2P":            "ws://ds-server.depaas.net:22334/p2p/",
	"MINER_RPC":      "http://123.100.236.30:46658",
	"MINER_CONTRACT": "0x45183a2F908f8975dCB1171c175D10FaC3604E38",
	"LOG":            "info",
	"INTERVAL_TIME":  "600s",
}

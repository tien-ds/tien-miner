//go:build dev
// +build dev

package env

var env = map[string]string{
	//"P2P":            "ws://127.0.0.1:8099/p2p/",
	"P2P":            "ws://172.26.112.1:8099/p2p/",
	"MINER_RPC":      "http://123.100.236.30:46658",
	"MINER_CONTRACT": "0xf725c0E7B6605c4F65626b24f4E52B663A9c4F3F",
	"LOG":            "trace",
	"INTERVAL_TIME":  "600s",
}

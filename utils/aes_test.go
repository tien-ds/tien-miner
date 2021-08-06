package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestAes(t *testing.T) {
	origData := []byte("Hello World")                                                  // 待加密的数据
	key := []byte(fmt.Sprintf("%s%d", "0HNToLr1j37vkGoiLgsiKF2", time.Now().Unix()/5)) // 加密的密钥
	log.Println("原文：", string(origData))

	log.Println("------------------ CBC模式 --------------------")
	encrypted := AesEncryptCBC(origData, key)
	log.Println("密文(hex)：", hex.EncodeToString(encrypted))
	log.Println("密文(base64)：", base64.StdEncoding.EncodeToString(encrypted))
	decrypted := AesDecryptCBC(encrypted, key)
	log.Println("解密结果：", string(decrypted))

	log.Println("------------------ ECB模式 --------------------")
	encrypted = AesEncryptECB(origData, key)
	log.Println("密文(hex)：", hex.EncodeToString(encrypted))
	log.Println("密文(base64)：", base64.StdEncoding.EncodeToString(encrypted))
	decrypted = AesDecryptECB(encrypted, key)
	log.Println("解密结果：", string(decrypted))

	log.Println("------------------ CFB模式 --------------------")
	encrypted = AesEncryptCFB(origData, key)
	log.Println("密文(hex)：", hex.EncodeToString(encrypted))
	log.Println("密文(base64)：", base64.StdEncoding.EncodeToString(encrypted))
	decrypted = AesDecryptCFB(encrypted, key)
	log.Println("解密结果：", string(decrypted))
}

func TestName(t *testing.T) {
	fmt.Println(time.Now().Format("2006-01-02T15:04"))
	fmt.Println(fmt.Sprintf("%s%d", "0HNToLr1j37vkGoiLgsiKF2", time.Now().Unix()/5))
}

func TestBase64(t *testing.T) {
	a := `{"Identity":{"PeerID":"12D3KooWMbH85g8p7cxgTMRXbYDbV8eHHzEfXbxE6HD4MnwrXnQi","PrivKey":"CAESQEoPU4ga2rYCuKRNfTTVcDs9WMnE3/hPEebRAP6q6J6MrvH8ttL3w1VzEJxsnSMgcQYhGx4FoxLxaHwDyt2gc3M="},"Datastore":{"StorageMax":"10GB","StorageGCWatermark":90,"GCPeriod":"1h","Spec":{"child":{"path":"dsblocks","syncWrites":false,"truncate":true,"type":"badgerds"},"prefix":"badger.datastore","type":"measure"},"HashOnRead":false,"BloomFilterSize":0},"Addresses":{"Swarm":["/ip4/0.0.0.0/tcp/11401","/ip6/::/tcp/11401","/ip4/0.0.0.0/udp/11401/quic","/ip6/::/udp/11401/quic"],"Announce":[],"NoAnnounce":[],"API":"/ip4/127.0.0.1/tcp/5001","Gateway":"/ip4/127.0.0.1/tcp/8080"},"Mounts":{"IPFS":"/ipfs","IPNS":"/ipns","FuseAllowOther":false},"Discovery":{"MDNS":{"Enabled":true,"Interval":10}},"Routing":{"Type":"dht"},"Ipns":{"RepublishPeriod":"","RecordLifetime":"","ResolveCacheSize":128},"Bootstrap":["/ip4/39.99.129.137/tcp/4001/p2p/12D3KooWNxdXLNTheDK6WrPh1x46EBYHBPbAoGFNsb3mJFrheVKZ","/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWNxdXLNTheDK6WrPh1x46EBYHBPbAoGFNsb3mJFrheVKZ"],"Gateway":{"HTTPHeaders":{"Access-Control-Allow-Headers":["X-Requested-With","Range","User-Agent"],"Access-Control-Allow-Methods":["GET"],"Access-Control-Allow-Origin":["*"]},"RootRedirect":"","Writable":false,"PathPrefixes":[],"APICommands":[],"NoFetch":false,"NoDNSLink":false,"PublicGateways":null},"API":{"HTTPHeaders":{}},"Swarm":{"AddrFilters":null,"DisableBandwidthMetrics":false,"DisableNatPortMap":false,"EnableRelayHop":false,"EnableAutoRelay":false,"Transports":{"Network":{},"Security":{},"Multiplexers":{}},"ConnMgr":{"Type":"basic","LowWater":600,"HighWater":900,"GracePeriod":"20s"}},"AutoNAT":{},"Pubsub":{"Router":"","DisableSigning":false},"Peering":{"Peers":null},"Provider":{"Strategy":""},"Reprovider":{"Interval":"12h","Strategy":"all"},"Experimental":{"FilestoreEnabled":false,"UrlstoreEnabled":false,"ShardingEnabled":false,"GraphsyncEnabled":false,"Libp2pStreamMounting":false,"P2pHttpProxy":false,"StrategicProviding":false},"Plugins":{"Plugins":null},"Pinning":{"RemoteServices":null}}`
	decodeString := base64.StdEncoding.EncodeToString(AesEncryptCBC([]byte(a), AesPasswd()))
	fmt.Println(decodeString)

	//
	bytes, err := base64.StdEncoding.DecodeString(decodeString)
	if err != nil {
		fmt.Println(err)
		return
	}
	cbc := AesDecryptCBC(bytes, AesPasswd())
	fmt.Println(string(cbc))
}

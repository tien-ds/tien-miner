module github.com/ds/depaas

go 1.16

require (
	berty.tech/go-orbit-db v1.11.4
	gitee.com/aifuturewell/gojni v0.0.0-20210507105514-3201d9b6ae5d
	gitee.com/fast_api/api v0.0.0-20210420101017-8f0953c2255c
	github.com/StackExchange/wmi v1.2.0 // indirect
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/buger/jsonparser v0.0.0-20181115193947-bf1c66bbce23
	github.com/dustin/go-humanize v1.0.0
	github.com/ethereum/go-ethereum v1.8.20
	github.com/fatih/color v1.10.0 // indirect
	github.com/gabriel-vasile/mimetype v1.1.2
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/golang/mock v1.5.0 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-cidutil v0.0.2
	github.com/ipfs/go-datastore v0.4.5
	github.com/ipfs/go-ds-measure v0.1.0
	github.com/ipfs/go-filestore v0.0.3
	github.com/ipfs/go-fs-lock v0.0.6
	github.com/ipfs/go-ipfs v0.8.0
	github.com/ipfs/go-ipfs-cmds v0.6.0
	github.com/ipfs/go-ipfs-config v0.12.0
	github.com/ipfs/go-ipfs-files v0.0.8
	github.com/ipfs/go-log v1.0.4
	github.com/ipfs/go-log/v2 v2.1.1
	github.com/ipfs/go-merkledag v0.3.2
	github.com/ipfs/go-mfs v0.1.2
	github.com/ipfs/go-path v0.0.9
	github.com/ipfs/interface-go-ipfs-core v0.4.0
	github.com/jbenet/go-is-domain v1.0.5
	github.com/libp2p/go-libp2p v0.13.0
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/mattn/go-sqlite3 v1.14.8
	github.com/mitchellh/go-homedir v1.1.0
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/multiformats/go-multiaddr-dns v0.2.0
	github.com/multiformats/go-multibase v0.0.3
	github.com/shirou/gopsutil v3.21.6+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.1
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca
	github.com/tien-ds/contract-miner v0.0.0
	github.com/tklauser/go-sysconf v0.3.7 // indirect
	github.com/whyrusleeping/cbor-gen v0.0.0-20200715143311-227fab5a2377 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420 // indirect
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210604141403-392c879c8b08 // indirect
)

replace (
	github.com/ethereum/go-ethereum v1.8.20 => github.com/loomnetwork/go-ethereum v1.8.17-0.20191122084538-6128fa1a8c76

	//must be
	github.com/loomnetwork/go-loom v0.0.0 => github.com/tien-ds/go-loom v0.0.0-20210806092349-e916fc4e73d1

	github.com/phonkee/go-pubsub v0.0.0 => github.com/loomnetwork/go-pubsub v0.0.0-20180626134536-2d1454660ed1

	github.com/tien-ds/contract-miner v0.0.0 => ../contract-miner
)

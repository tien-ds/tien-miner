package ipds

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	config "github.com/ipfs/go-ipfs-config"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/ipfs/interface-go-ipfs-core/options"
	ci "github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

func GenConfig(out io.Writer, nBitsForKeypair int) (*config.Config, error) {
	identity, err := CreateIdentity(out, []options.KeyGenerateOption{
		options.Key.Size(nBitsForKeypair),
		options.Key.Type(options.Ed25519Key),
	})
	if err != nil {
		return nil, err
	}
	return InitWithIdentity(identity)
}

func ParseBootstrapPeers(addrs []string) ([]peer.AddrInfo, error) {
	maddrs := make([]ma.Multiaddr, len(addrs))
	for i, addr := range addrs {
		var err error
		maddrs[i], err = ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
	}
	return peer.AddrInfosFromP2pAddrs(maddrs...)
}

var DefaultBootstrapAddresses = []string{
	//"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
	//"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
	//"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
	//"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
	//"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",       // mars.i.ipfs.io
	//"/ip4/39.99.129.137/tcp/11401/p2p/12D3KooWDUwpc4o9v5sMQLeLTjJgwu4V4KJkKU9ojpGk2PsA3Pts",
}

func DefaultBootstrapPeers() ([]peer.AddrInfo, error) {
	ps, err := ParseBootstrapPeers(DefaultBootstrapAddresses)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse hardcoded bootstrap peers: %s
			This is a problem with the ipfs codebase. Please report it to the dev team.`, err)
	}
	return ps, nil
}

func InitWithIdentity(identity config.Identity) (*config.Config, error) {
	bootstrapPeers, err := DefaultBootstrapPeers()
	if err != nil {
		return nil, err
	}

	datastore := DefaultDatastoreConfig()

	conf := &config.Config{
		Addresses: addressesConfig(11401),
		Datastore: datastore,
		Bootstrap: config.BootstrapPeerStrings(bootstrapPeers),
		Identity:  identity,
		Discovery: config.Discovery{
			MDNS: config.MDNS{
				Enabled:  true,
				Interval: 5,
			},
		},

		Routing: config.Routing{
			Type: "dht",
		},

		//// setup the node mount points.
		//Mounts: config.Mounts{
		//	IPFS: "/ipfs",
		//	IPNS: "/ipns",
		//},

		Ipns: config.Ipns{
			ResolveCacheSize: 128,
		},

		Gateway: config.Gateway{
			RootRedirect: "",
			Writable:     false,
			NoFetch:      false,
			PathPrefixes: []string{},
			HTTPHeaders: map[string][]string{
				"Access-Control-Allow-Origin":  {"*"},
				"Access-Control-Allow-Methods": {"GET"},
				"Access-Control-Allow-Headers": {"X-Requested-With", "Range", "User-Agent"},
			},
			APICommands: []string{},
		},
		Reprovider: config.Reprovider{
			Interval: "12h",
			Strategy: "all",
		},
		Swarm: config.SwarmConfig{
			ConnMgr: config.ConnMgr{
				LowWater:    DefaultConnMgrLowWater,
				HighWater:   DefaultConnMgrHighWater,
				GracePeriod: DefaultConnMgrGracePeriod.String(),
				Type:        "basic",
			},
		},
		AutoNAT: config.AutoNATConfig{
			ServiceMode: config.AutoNATServiceEnabled,
		},
	}

	return conf, nil
}

// DefaultConnMgrHighWater is the default value for the connection managers
// 'high water' mark
const DefaultConnMgrHighWater = 900

// DefaultConnMgrLowWater is the default value for the connection managers 'low
// water' mark
const DefaultConnMgrLowWater = 600

// DefaultConnMgrGracePeriod is the default value for the connection managers
// grace period
const DefaultConnMgrGracePeriod = time.Second * 20

func addressesConfig(port int) config.Addresses {

	return config.Addresses{
		Swarm: []string{
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),
			fmt.Sprintf("/ip6/::/tcp/%d", port),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port),
			fmt.Sprintf("/ip6/::/udp/%d/quic", port),
		},
		Announce:   []string{},
		NoAnnounce: []string{},
		API:        config.Strings{"/ip4/127.0.0.1/tcp/5001"},
		Gateway:    config.Strings{"/ip4/127.0.0.1/tcp/8080"},
	}
}

// DefaultDatastoreConfig is an internal function exported to aid in testing.
func DefaultDatastoreConfig() config.Datastore {
	return config.Datastore{
		StorageMax:         "10GB",
		StorageGCWatermark: 90, // 90%
		GCPeriod:           "1h",
		BloomFilterSize:    0,
		Spec:               badgerSpec(),
	}
}

func badgerSpec() map[string]interface{} {
	return map[string]interface{}{
		"type":   "measure",
		"prefix": "badger.datastore",
		"child": map[string]interface{}{
			"type":       "badgerds",
			"path":       "dsblocks",
			"syncWrites": false,
			"truncate":   true,
		},
	}
}

func flatfsSpec() map[string]interface{} {
	return map[string]interface{}{
		"type": "mount",
		"mounts": []interface{}{
			map[string]interface{}{
				"mountpoint": "/blocks",
				"type":       "measure",
				"prefix":     "flatfs.datastore",
				"child": map[string]interface{}{
					"type":      "flatfs",
					"path":      "blocks",
					"sync":      true,
					"shardFunc": "/repo/flatfs/shard/v1/next-to-last/2",
				},
			},
			map[string]interface{}{
				"mountpoint": "/",
				"type":       "measure",
				"prefix":     "leveldb.datastore",
				"child": map[string]interface{}{
					"type":        "levelds",
					"path":        "datastore",
					"compression": "none",
				},
			},
		},
	}
}

// CreateIdentity initializes a new identity.
func CreateIdentity(out io.Writer, opts []options.KeyGenerateOption) (config.Identity, error) {
	// TODO guard higher up
	ident := config.Identity{}

	settings, err := options.KeyGenerateOptions(opts...)
	if err != nil {
		return ident, err
	}

	var sk ci.PrivKey
	var pk ci.PubKey

	switch settings.Algorithm {
	case "rsa":
		if settings.Size == -1 {
			settings.Size = options.DefaultRSALen
		}
		fmt.Fprintf(out, "generating %d-bit RSA keypair...", settings.Size)
		priv, pub, err := ci.GenerateKeyPair(ci.RSA, settings.Size)
		if err != nil {
			return ident, err
		}
		sk = priv
		pk = pub
	case "ed25519":
		fmt.Fprintf(out, "generating ED25519 keypair...")
		priv, pub, err := ci.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return ident, err
		}
		sk = priv
		pk = pub
	default:
		return ident, fmt.Errorf("unrecognized key type: %s", settings.Algorithm)
	}
	fmt.Fprintf(out, "done\n")

	// currently storing key unEncrypted. in the future we need to encrypt it.
	// TODO(security)
	skBytes, err := sk.Bytes()
	if err != nil {
		return ident, err
	}
	ident.PrivKey = base64.StdEncoding.EncodeToString(skBytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return ident, err
	}
	ident.PeerID = id.Pretty()
	fmt.Fprintf(out, "peer identity: %s\n", ident.PeerID)
	return ident, nil
}

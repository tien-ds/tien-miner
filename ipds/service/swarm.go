package service

import (
	"context"
	"fmt"
	"github.com/ds/depaas/ipds"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"
	"sync"
	"time"
)

const (
	dnsResolveTimeout = 10 * time.Second
)

func ConnectPeer(addrs []string) error {
	background := context.Background()
	pis, err := parseAddresses(background, addrs)
	if err != nil {
		return err
	}
	api := ipds.GetApi()
	output := make([]string, len(pis))
	for i, pi := range pis {
		output[i] = "connect " + pi.ID.Pretty()
		err := api.Swarm().Connect(background, pi)
		if err != nil {
			return fmt.Errorf("%s failure: %s", output[i], err)
		}
		output[i] += " success"
	}
	return nil
}

func parseAddresses(ctx context.Context, addrs []string) ([]peer.AddrInfo, error) {
	// resolve addresses
	maddrs, err := resolveAddresses(ctx, addrs)
	if err != nil {
		return nil, err
	}

	return peer.AddrInfosFromP2pAddrs(maddrs...)
}

// resolveAddresses resolves addresses parallelly
func resolveAddresses(ctx context.Context, addrs []string) ([]ma.Multiaddr, error) {
	ctx, cancel := context.WithTimeout(ctx, dnsResolveTimeout)
	defer cancel()

	var maddrs []ma.Multiaddr
	var wg sync.WaitGroup
	resolveErrC := make(chan error, len(addrs))

	maddrC := make(chan ma.Multiaddr)

	for _, addr := range addrs {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}

		// check whether address ends in `ipfs/Qm...`
		if _, last := ma.SplitLast(maddr); last.Protocol().Code == ma.P_IPFS {
			maddrs = append(maddrs, maddr)
			continue
		}
		wg.Add(1)
		go func(maddr ma.Multiaddr) {
			defer wg.Done()
			raddrs, err := madns.Resolve(ctx, maddr)
			if err != nil {
				resolveErrC <- err
				return
			}
			// filter out addresses that still doesn't end in `ipfs/Qm...`
			found := 0
			for _, raddr := range raddrs {
				if _, last := ma.SplitLast(raddr); last != nil && last.Protocol().Code == ma.P_IPFS {
					maddrC <- raddr
					found++
				}
			}
			if found == 0 {
				resolveErrC <- fmt.Errorf("found no ipfs peers at %s", maddr)
			}
		}(maddr)
	}
	go func() {
		wg.Wait()
		close(maddrC)
	}()

	for madder := range maddrC {
		maddrs = append(maddrs, madder)
	}

	select {
	case err := <-resolveErrC:
		return nil, err
	default:
	}

	return maddrs, nil
}

package register

import (
	"fmt"
	"github.com/ds/depaas/ipds"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

var mIp string

func SetAppIp(ip string) {
	mIp = ip
}

func GetAppIp() string {
	return mIp
}

func MulAddr() []byte {
	multiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/11401/p2p/%s", mIp, ipds.GetPeerID()))
	if err != nil {
		logrus.Error(err)
		return nil
	}
	return multiAddr.Bytes()
}

package web

import (
	"github.com/ds/depaas/ipds"
	"github.com/ipfs/go-ipfs/core/corehttp"
)

func StartFileServer() {
	//http server
	addr := "/ip4/0.0.0.0/tcp/18080"
	var opts = []corehttp.ServeOption{
		GatewayOption(false, "/ds", "/ipns"),
	}
	go corehttp.ListenAndServe(ipds.GetNode(), addr, opts...)
}

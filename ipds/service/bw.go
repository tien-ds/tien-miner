package service

import (
	"github.com/ds/depaas/ipds"
	"github.com/libp2p/go-libp2p-core/metrics"
)

func StatsBw() metrics.Stats {
	reporter := ipds.GetNode().Reporter
	return reporter.GetBandwidthTotals()
}

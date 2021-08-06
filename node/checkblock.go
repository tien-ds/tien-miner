package node

import (
	"context"
	"github.com/ds/depaas/database"
	"github.com/ds/depaas/ipds/service"
	"github.com/ds/depaas/protocol"
	"github.com/sirupsen/logrus"
	"time"
)

func StartCheckBlock(ctx context.Context) {
	period, err := time.ParseDuration("1440h")
	if err != nil {
		logrus.Error(err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			break
		case <-time.After(period):
			// the private func maybeGC doesn't compute storageMax, storageGC, slackGC so that they are not re-computed for every cycle
			Check()
		}
	}
}

func Check() {
	logrus.Debug("check block run .....")
	cid, err := service.AddFile(database.GetDBPath())
	if err != nil {
		logrus.Error(err)
		return
	}
	err = Current().SendEncryptMessage(protocol.BlockCheck{
		MsgType: protocol.MsgType{
			Type: protocol.BLOCK_CHECK,
		},
		Cid: cid,
	})
	if err != nil {
		logrus.Error(err)
		return
	}
}

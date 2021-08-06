package main

import (
	"github.com/ds/depaas/ipds/service"
	"github.com/sirupsen/logrus"
)

func ConnectBoot(p string) {
	err := service.ConnectPeer([]string{p})
	if err == nil {
		logrus.Infof("connect to %s success", p)
	}
}

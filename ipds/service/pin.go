package service

import (
	"context"
	"github.com/ds/depaas/ipds"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-cidutil/cidenc"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/sirupsen/logrus"
)

func PinAdd(p cid.Cid) (string, error) {
	api := ipds.GetApi()
	ctx := context.Background()
	resolvePath, err := api.ResolvePath(ctx, path.IpfsPath(p))
	if err != nil {
		return "", err
	}
	if err := api.Pin().Add(ctx, resolvePath, options.Pin.Recursive(true)); err != nil {
		return "", err
	}
	logrus.Debugf("pin add %s success", p)
	e := cidenc.Default()
	return e.Encode(resolvePath.Cid()), nil
}

func PinRm(p cid.Cid) (string, error) {
	api := ipds.GetApi()
	ctx := context.Background()
	resolvePath, err := api.ResolvePath(ctx, path.IpfsPath(p))
	if err != nil {
		return "", err
	}
	logrus.Debugf("pin add %s success", p)
	if err := api.Pin().Rm(ctx, resolvePath, options.Pin.RmRecursive(true)); err != nil {
		return "", err
	}
	e := cidenc.Default()
	return e.Encode(resolvePath.Cid()), nil
}

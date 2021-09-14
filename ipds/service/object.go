package service

import (
	"context"
	"github.com/ds/depaas/ipds"
	"github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

func GetFile(cid cid.Cid, fPath string) error {
	getED, err := ipds.GetApi().Unixfs().Get(context.Background(), path.IpfsPath(cid))
	if err != nil {
		logrus.Error(err)
		return err
	}
	//FIXME fPath must limit
	openFile, err := os.OpenFile(fPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer openFile.Close()
	_, err = io.Copy(openFile, files.ToFile(getED))
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func AddFile(fPath string) (string, error) {
	open, err := os.Open(fPath)
	if err != nil {
		return "", err
	}
	defer open.Close()
	add, err := ipds.GetApi().Unixfs().Add(context.Background(), files.NewReaderFile(open))
	if err != nil {
		return "", err
	}
	return add.Cid().String(), nil
}

package service

import (
	"context"
	"github.com/ds/depaas/ipds"

	"github.com/ipfs/go-ipfs/core/corerepo"
)

func RepoSize() *corerepo.SizeStat {
	repoSize, err := corerepo.RepoSize(context.Background(), ipds.GetNode())
	if err != nil {
		return nil
	}
	return &repoSize
}

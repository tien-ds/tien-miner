package dbfs

import (
	"github.com/ipfs/go-filestore"
	config "github.com/ipfs/go-ipfs-config"
	keystore "github.com/ipfs/go-ipfs/keystore"
	"github.com/ipfs/go-ipfs/repo"
	ma "github.com/multiformats/go-multiaddr"
)

type DS struct {
}

func (D *DS) Config() (*config.Config, error) {
	panic("implement me")
}

func (D *DS) BackupConfig(prefix string) (string, error) {
	panic("implement me")
}

func (D *DS) SetConfig(config *config.Config) error {
	panic("implement me")
}

func (D *DS) SetConfigKey(key string, value interface{}) error {
	panic("implement me")
}

func (D *DS) GetConfigKey(key string) (interface{}, error) {
	panic("implement me")
}

func (D *DS) Datastore() repo.Datastore {
	panic("implement me")
}

func (D *DS) GetStorageUsage() (uint64, error) {
	panic("implement me")
}

func (D *DS) Keystore() keystore.Keystore {
	panic("implement me")
}

func (D *DS) FileManager() *filestore.FileManager {
	panic("implement me")
}

func (D *DS) SetAPIAddr(addr ma.Multiaddr) error {
	panic("implement me")
}

func (D *DS) SwarmKey() ([]byte, error) {
	panic("implement me")
}

func (D *DS) Close() error {
	panic("implement me")
}
